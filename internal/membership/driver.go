package membership

import (
	"context"

	pb "github.com/drakeor/jmgr/api"
)

// Driver defines the minimum membership/broadcast API required by the core loop.
// All implementations must be goroutine-safe.
type Driver interface {
	Start(ctx context.Context, seeds []string) error
	Stop()
	Send(m *pb.Metric)
	Metrics() <-chan *pb.Metric
	MergeEvents() <-chan MergeEvent
	MetricsEvents() <-chan *pb.Metric
	Events() <-chan Event
	SelfID() string // must return local node ID
	View() *View

	HandleJoin(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error)
	HandleStream(srv pb.Cluster_StreamServer) error
}
