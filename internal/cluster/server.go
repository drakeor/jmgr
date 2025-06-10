package cluster

import (
	"context"

	pb "github.com/drakeor/jmgr/api"
)

// Server is the root gRPC server for the cluster.
// It delegates each set of RPCs to the respective module implementation.
type Server struct {
	pb.UnimplementedClusterServer

	Membership  MembershipRPC
	Store       StoreRPC
	AntiEntropy AntiEntropyRPC
}

// Membership
func (s *Server) Join(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error) {
	return s.Membership.Join(ctx, req)
}
func (s *Server) Stream(srv pb.Cluster_StreamServer) error {
	return s.Membership.Stream(srv)
}

// Store
func (s *Server) ListBuckets(ctx context.Context, req *pb.Empty) (*pb.BucketList, error) {
	return s.Store.ListBuckets(ctx, req)
}
func (s *Server) GetMetricPoints(ctx context.Context, req *pb.GetMetricRequest) (*pb.MetricPoints, error) {
	return s.Store.GetMetricPoints(ctx, req)
}

// Anti-entropy
func (s *Server) GetNodeSummary(ctx context.Context, req *pb.Empty) (*pb.NodeSummary, error) {
	return s.AntiEntropy.GetNodeSummary(ctx, req)
}
