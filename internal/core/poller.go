package core

import pb "github.com/drakeor/jmgr/api"

// Poller provides a channel of live Metric points. It can be closed.
type Poller interface {
	Metrics() <-chan *pb.Metric
	Close()
}
