package antientropy

import (
	"context"
	"testing"
	"time"

	"fmt"
	"net"

	pb "github.com/drakeor/jmgr/api"
	"github.com/drakeor/jmgr/internal/connections"
	"github.com/drakeor/jmgr/internal/store"
	"github.com/drakeor/jmgr/internal/utils"
	"google.golang.org/grpc"
)

func startGRPCStoreServer(t *testing.T, store store.Driver, addr string, id string) func() {
	srv := grpc.NewServer()
	// This is your real test anti-entropy RPC server:
	pb.RegisterClusterServer(srv, &testServer{store: store})

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	go srv.Serve(lis)
	return func() {
		srv.GracefulStop()
		lis.Close()
	}
}

type testServer struct {
	pb.UnimplementedClusterServer
	store store.Driver
}

func (s *testServer) GetNodeSummary(ctx context.Context, _ *pb.Empty) (*pb.NodeSummary, error) {
	summary := SummarizeStore(s.store)
	return ToProtoNodeSummary(summary), nil
}

func (s *testServer) ListBuckets(ctx context.Context, _ *pb.Empty) (*pb.BucketList, error) {
	return &pb.BucketList{Buckets: s.store.ListBuckets()}, nil
}

func (s *testServer) GetMetricPoints(ctx context.Context, req *pb.GetMetricRequest) (*pb.MetricPoints, error) {
	nodeID, name, labels, err := utils.ParseBucketKey(req.Bucket)
	if err != nil {
		return nil, err
	}
	pts, _ := s.store.Get(nodeID, name, time.Unix(0, 0), labels)
	return &pb.MetricPoints{Points: pts}, nil
}

func TestAntiEntropyRepairsMissingBucket(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	a, _ := store.NewBadgerRing("a", nil, "", 1<<20, true)
	b, _ := store.NewBadgerRing("b", nil, "", 1<<20, true)
	defer a.Close()
	defer b.Close()

	// Put a metric on both
	m := &pb.Metric{
		NodeId: "a", Name: "cpu_pct", Labels: map[string]string{"core": "0"},
		Points: []*pb.MetricPoint{{TsMs: time.Now().UnixMilli(), Value: 50}},
		Source: "a", Tier: "hot",
	}
	_ = a.Put(m)
	_ = b.Put(m)

	bucket := fmt.Sprintf("%s|%s|%s", m.NodeId, m.Name, utils.SerializeLabels(m.Labels))
	if !contains(a.ListBuckets(), bucket) || !contains(b.ListBuckets(), bucket) {
		t.Fatalf("expected both to have the bucket")
	}

	// Now, simulate "bucket missing on B"
	b2, err := store.NewBadgerRing("b", nil, "", 1<<20, true)
	if err != nil {
		t.Fatalf("failed to create new badger ring for b: %v", err)
	}
	defer b2.Close()

	if contains(b2.ListBuckets(), bucket) {
		t.Fatalf("expected bucket missing on B")
	}

	// Start gRPC server for A
	const peerID = "a"
	const peerAddr = "127.0.0.1:50051"
	cleanup := startGRPCStoreServer(t, a, peerAddr, peerID)
	defer cleanup()

	connMgr := connections.NewConnManager()
	connMgr.AddPeer(peerID, peerAddr)

	engine := NewSyncEngine(16)
	go engine.Run(ctx, 2, func(job SyncJob) {
		// This is your actual sync logic: fetch data from peer for (bucket, bin)
		// For the test, just pull the points from A and put into B
		nodeID, name, labels, _ := utils.ParseBucketKey(job.Bucket)
		pts, _ := a.Get(nodeID, name, time.Unix(0, 0), labels)
		m := &pb.Metric{NodeId: nodeID, Name: name, Labels: labels, Points: pts, Source: "a", Tier: "hot"}
		_ = b2.Put(m)
	})

	anti := NewService(
		"b",
		b2,
		connMgr,
		1*time.Second,
		engine,
	)
	anti.syncOnce(ctx)

	utils.AssertEventually(t, 2*time.Second, func() bool {
		return contains(b2.ListBuckets(), bucket)
	})

	utils.AssertEventually(t, 2*time.Second, func() bool {
		nodeID, name, labels, _ := utils.ParseBucketKey(bucket)
		pts, _ := b2.Get(nodeID, name, time.Unix(0, 0), labels)
		return len(pts) > 0
	})
}

func contains(xs []string, x string) bool {
	for _, s := range xs {
		if s == x {
			return true
		}
	}
	return false
}
