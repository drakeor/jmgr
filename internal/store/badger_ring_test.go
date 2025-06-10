package store

import (
	"testing"
	"time"

	pb "github.com/drakeor/jmgr/api"
)

func TestRingRoutingAndVacuum(t *testing.T) {
	peerIDs := []string{"a", "b", "c"}
	a, _ := NewBadgerRing("a", peerIDs, "", 1<<20, true)
	b, _ := NewBadgerRing("b", peerIDs, "", 1<<20, true)
	c, _ := NewBadgerRing("c", peerIDs, "", 1<<20, true)
	stores := []*BadgerRing{a, b, c}

	for i := 0; i < 90; i++ {
		node := string(rune('a' + rune(i%3)))
		metric := &pb.Metric{
			NodeId: node,
			Name:   "cpu_pct",
			Labels: map[string]string{},
			Points: []*pb.MetricPoint{
				{TsMs: time.Now().UnixMilli(), Value: float64(i)},
			},
			Source: node, // Set at source
			Tier:   "hot",
		}
		// Simulate full replication: each node might get a copy (as in real gossip)
		for _, s := range stores {
			_ = s.Put(metric)
		}
	}

	time.Sleep(250 * time.Millisecond) // allow async write, increase if still fails

	replicaCnt := 0
	for _, s := range stores {
		pts, _ := s.Get("a", "cpu_pct", time.Unix(0, 0), map[string]string{})
		if len(pts) > 0 {
			replicaCnt++
		}
	}
	if replicaCnt != len(stores) {
		t.Fatalf("expected %d replicas for node a, got %d", len(stores), replicaCnt)
	}

	// rest of your test...
}

func TestAllReplicasContainAllMetrics(t *testing.T) {
	peerIDs := []string{"a", "b", "c"}
	a, _ := NewBadgerRing("a", peerIDs, "", 1<<20, true)
	b, _ := NewBadgerRing("b", peerIDs, "", 1<<20, true)
	c, _ := NewBadgerRing("c", peerIDs, "", 1<<20, true)
	stores := []*BadgerRing{a, b, c}
	ids := []string{"a", "b", "c"}

	for i, node := range ids {
		metric := &pb.Metric{
			NodeId: node,
			Name:   "cpu_pct",
			Labels: map[string]string{},
			Points: []*pb.MetricPoint{{TsMs: time.Now().UnixMilli(), Value: float64(i)}},
			Source: node, // Set at source
			Tier:   "hot",
		}

		// Instead of _ = stores[i].Put(metric)
		// Put the metric on *all* stores, so any that are replicas will accept it
		for _, s := range stores {
			_ = s.Put(metric)
		}
	}

	// Now all nodes should have all metrics (since replication is 2, allow short delay)
	time.Sleep(250 * time.Millisecond)

	for _, s := range stores {
		for _, node := range ids {
			pts, _ := s.Get(node, "cpu_pct", time.Unix(0, 0), map[string]string{})
			if len(pts) == 0 {
				t.Fatalf("expected metric for node %q, missing on store %q", node, s.selfID)
			}
		}
	}
}
