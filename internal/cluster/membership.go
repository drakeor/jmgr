package cluster

import (
	"context"

	pb "github.com/drakeor/jmgr/api"
	"github.com/drakeor/jmgr/internal/membership"
)

type MembershipHandler struct {
	Driver membership.Driver
}

func (m *MembershipHandler) Join(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error) {
	return m.Driver.HandleJoin(ctx, req)
}

func (m *MembershipHandler) Stream(srv pb.Cluster_StreamServer) error {
	return m.Driver.HandleStream(srv)
}
