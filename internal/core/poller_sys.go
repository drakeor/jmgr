package core

import (
	"sync"
	"time"

	pb "github.com/drakeor/jmgr/api"
	"github.com/drakeor/jmgr/internal/core/collectors"
)

// SysPoller runs a set of collectors and emits their output.
type SysPoller struct {
	out  chan *pb.Metric
	wg   sync.WaitGroup
	stop chan struct{}
}

// Creates a new SysPoller with the provided collectors.
func NewSysPoller(cs ...collectors.Collector) *SysPoller {
	p := &SysPoller{
		out:  make(chan *pb.Metric, 128),
		stop: make(chan struct{}),
	}
	for _, c := range cs {
		p.wg.Add(1)
		go func(col collectors.Collector) {
			defer p.wg.Done()
			tick := time.NewTicker(col.Interval())
			defer tick.Stop()
			for {
				select {
				case <-tick.C:
					if m, err := col.Collect(); err == nil {
						select {
						case p.out <- m:
						default:
						}
					}
				case <-p.stop:
					return
				}
			}
		}(c)
	}
	return p
}

// Gets the channel for metrics emitted by the SysPoller.
func (p *SysPoller) Metrics() <-chan *pb.Metric { return p.out }

func (p *SysPoller) Close() {
	close(p.stop)
	p.wg.Wait()
	close(p.out)
}
