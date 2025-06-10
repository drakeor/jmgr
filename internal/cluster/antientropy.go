package cluster

import (
	"context"
	"fmt"

	pb "github.com/drakeor/jmgr/api"
	"github.com/drakeor/jmgr/internal/antientropy"
	"github.com/drakeor/jmgr/internal/store"
)

type AntiEntropyHandler struct {
	Store store.Driver
}

func (h *AntiEntropyHandler) GetNodeSummary(ctx context.Context, _ *pb.Empty) (*pb.NodeSummary, error) {
	if h.Store == nil {
		return nil, fmt.Errorf("store not initialized on server")
	}
	summary := antientropy.SummarizeStore(h.Store)
	return antientropy.ToProtoNodeSummary(summary), nil
}
