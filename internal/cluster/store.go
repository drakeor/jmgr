package cluster

import (
	"context"
	"time"

	pb "github.com/drakeor/jmgr/api"
	"github.com/drakeor/jmgr/internal/store"
	"github.com/drakeor/jmgr/internal/utils"
)

type StoreHandler struct {
	Store store.Driver
}

func (s *StoreHandler) ListBuckets(ctx context.Context, _ *pb.Empty) (*pb.BucketList, error) {
	buckets := s.Store.ListBuckets()
	return &pb.BucketList{Buckets: buckets}, nil
}

func (s *StoreHandler) GetMetricPoints(ctx context.Context, req *pb.GetMetricRequest) (*pb.MetricPoints, error) {
	nodeID, name, labels, err := utils.ParseBucketKey(req.Bucket)
	if err != nil {
		return nil, err
	}
	pts, _ := s.Store.Get(nodeID, name, time.Unix(0, 0), labels)
	return &pb.MetricPoints{Points: pts}, nil
}
