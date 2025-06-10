package collectors

import (
	"time"

	pb "github.com/drakeor/jmgr/api"
	"github.com/shirou/gopsutil/v4/disk"
)

type DiskFreeCollector struct {
	NodeID string
	Mount  string
}

func (d *DiskFreeCollector) Interval() time.Duration { return time.Second }

func (d *DiskFreeCollector) Collect() (*pb.Metric, error) {
	du, err := disk.Usage(d.Mount)
	if err != nil {
		return nil, err
	}
	return &pb.Metric{
		NodeId: d.NodeID,
		Name:   "disk_free_bytes",
		Labels: map[string]string{"mount": d.Mount},
		Points: []*pb.MetricPoint{{TsMs: time.Now().UnixMilli(), Value: float64(du.Free)}},
		Source: d.NodeID,
		Tier:   "hot",
	}, nil
}
