package collectors

import (
	"time"

	pb "github.com/drakeor/jmgr/api"
	"github.com/shirou/gopsutil/v4/cpu"
)

type CPUCollector struct {
	NodeID string
}

func (c *CPUCollector) Interval() time.Duration { return time.Second }

func (c *CPUCollector) Collect() (*pb.Metric, error) {
	vals, err := cpu.Percent(0, false)
	if err != nil || len(vals) == 0 {
		return nil, err
	}
	return &pb.Metric{
		NodeId: c.NodeID,
		Name:   "cpu_pct",
		Labels: map[string]string{},
		Points: []*pb.MetricPoint{{TsMs: time.Now().UnixMilli(), Value: vals[0]}},
		Source: c.NodeID,
		Tier:   "hot",
	}, nil
}
