package store

import (
	"time"

	pb "github.com/drakeor/jmgr/api"
)

// Driver is the storage interface for all metric stores.
// Put stores a single metric point.
// Get returns all MetricPoints for nodeID+name+labels newer than 'since'.
// Vacuum should ensure total storage used is <= maxBytes (if supported).
// DiskUsage returns total storage used in bytes.
type Driver interface {
	Put(m *pb.Metric) error
	Get(nodeID, name string, since time.Time, labels map[string]string) ([]*pb.MetricPoint, error)
	Vacuum(maxBytes int64) error
	DiskUsage() int64
	ListBuckets() []string
}
