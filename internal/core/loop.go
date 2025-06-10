package core

import (
	"context"
	"sync"

	pb "github.com/drakeor/jmgr/api"
	"github.com/drakeor/jmgr/internal/alert"
	"github.com/drakeor/jmgr/internal/membership"
	"github.com/drakeor/jmgr/internal/store"
)

// Loop ties together Poller, Store, and Membership driver.
// Idk if this is the best method, or if I should make stuff more async
type Loop struct {
	poller Poller
	store  store.Driver
	memb   membership.Driver
	buf    chan *pb.Metric
	alert  *alert.Service
	once   sync.Once
}

// Creates a new Loop instance with the provided components.
func NewLoop(p Poller, s store.Driver, m membership.Driver, a *alert.Service) *Loop {
	return &Loop{
		poller: p,
		store:  s,
		memb:   m,
		alert:  a,
		buf:    make(chan *pb.Metric, 64),
	}
}

// Starts the loop, initializing the main loop
func (l *Loop) Start(ctx context.Context) {
	l.once.Do(func() {
		go l.pollLoop(ctx)
		go l.sendLoop(ctx)
		go l.recvLoop(ctx)
	})
}

// pollLoop collects local samples and pushes to buffer & store.
func (l *Loop) pollLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case m := <-l.poller.Metrics():
			_ = l.store.Put(m)
			// Process alert if configured
			if l.alert != nil {
				l.alert.Process(m)
			}
			// Enqueue metric for outbound
			select {
			case l.buf <- m:
			// drop if queue full
			default:
			}
		}
	}
}

// sendLoop drains local buf and broadcasts via membership.
func (l *Loop) sendLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case m := <-l.buf:
			// Set/overwrite source and tier before sending
			m.Source = l.memb.SelfID()
			m.Tier = "hot"
			l.memb.Send(m)
		}
	}
}

// recvLoop listens to membership events and stores inbound metrics.
func (l *Loop) recvLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case m := <-l.memb.Metrics():
			_ = l.store.Put(m)
		}
	}
}
