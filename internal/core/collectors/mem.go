package collectors

import (
	"time"

	pb "github.com/drakeor/jmgr/api"
	"github.com/shirou/gopsutil/v4/mem"
)

type MemFreeCollector struct {
	NodeID string
}

func (c *MemFreeCollector) Interval() time.Duration { return time.Second }

func (c *MemFreeCollector) Collect() (*pb.Metric, error) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	return &pb.Metric{
		NodeId: c.NodeID,
		Name:   "mem_free_bytes",
		Labels: map[string]string{},
		Points: []*pb.MetricPoint{{TsMs: time.Now().UnixMilli(), Value: float64(vm.Available)}},
		Source: c.NodeID,
		Tier:   "hot",
	}, nil
}
