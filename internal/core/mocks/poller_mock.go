//go:build mock

package core

import (
	"math/rand"
	"time"

	pb "github.com/drakeor/jmgr/api"
)

type MockPoller struct {
	nodeID string
	out    chan *pb.Metric
}

func NewMockPoller(nodeID string) *MockPoller {
	out := make(chan *pb.Metric, 64)
	go func() {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		t := time.NewTicker(200 * time.Millisecond)
		defer t.Stop()
		for now := range t.C {
			out <- &pb.Metric{
				NodeId: nodeID,
				Name:   "cpu_pct",
				Labels: map[string]string{},
				Points: []*pb.MetricPoint{
					{TsMs: now.UnixMilli(), Value: rng.Float64() * 100},
				},
			}
		}
	}()
	return &MockPoller{nodeID, out}
}

func (m *MockPoller) Metrics() <-chan *pb.Metric { return m.out }
func (m *MockPoller) Close()                     { close(m.out) }
