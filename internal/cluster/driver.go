package cluster

import (
	"context"

	pb "github.com/drakeor/jmgr/api"
)

// Membership-related RPC interface
type MembershipRPC interface {
	Join(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error)
	Stream(srv pb.Cluster_StreamServer) error
}

// Store/metrics-related RPC interface
type StoreRPC interface {
	ListBuckets(ctx context.Context, req *pb.Empty) (*pb.BucketList, error)
	GetMetricPoints(ctx context.Context, req *pb.GetMetricRequest) (*pb.MetricPoints, error)
}

// Anti-entropy/summary RPC interface
type AntiEntropyRPC interface {
	GetNodeSummary(ctx context.Context, req *pb.Empty) (*pb.NodeSummary, error)
}
