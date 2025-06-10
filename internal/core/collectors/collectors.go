package collectors

import (
	"time"

	pb "github.com/drakeor/jmgr/api"
)

// Collector defines a source of metrics sampled at a given interval.
type Collector interface {
	Interval() time.Duration
	Collect() (*pb.Metric, error)
}
