package membership

import (
	"context"
	"sync"
	"time"

	pb "github.com/drakeor/jmgr/api"
	"github.com/drakeor/jmgr/internal/connections"
	"github.com/drakeor/jmgr/internal/store"
	"github.com/drakeor/jmgr/internal/utils"
)

// MembershipNode is the membership driver
type MembershipNode struct {
	selfID  string
	addr    string
	view    *View
	connMgr *connections.ConnManager

	events     chan Event
	merges     chan MergeEvent
	metricsTap chan *pb.Metric
	metricsCh  chan *pb.Metric

	connsMu sync.RWMutex
	conns   map[string]*peerConn

	tickInterval   time.Duration
	dialerInterval time.Duration
	incarnation    uint64
	store          store.Driver
}

// Options for configuring the MembershipNode
type MembershipOption func(*MembershipNode)

// NewMembershipNode creates a new MembershipNode instance with the given selfID and addr.
func NewMembershipNode(selfID, addr string, connMgr *connections.ConnManager, opts ...MembershipOption) *MembershipNode {
	b := &MembershipNode{
		selfID:         selfID,
		addr:           addr,
		view:           NewView(),
		connMgr:        connMgr,
		events:         make(chan Event, 32),
		merges:         make(chan MergeEvent, 64),
		metricsTap:     make(chan *pb.Metric, 128),
		metricsCh:      make(chan *pb.Metric, 128),
		conns:          make(map[string]*peerConn),
		tickInterval:   1 * time.Second,
		dialerInterval: 3 * time.Second,
		incarnation:    1,
	}
	for _, opt := range opts {
		opt(b)
	}
	b.view.MergeDigest(&pb.MembershipDigest{
		Peers: []*pb.PeerInfo{{Id: selfID, Addr: addr, Incarnation: b.incarnation}},
	})
	b.connMgr.AddPeer(selfID, addr)
	return b
}

func (b *MembershipNode) Metrics() <-chan *pb.Metric        { return b.metricsCh }
func (b *MembershipNode) MetricsEvents() <-chan *pb.Metric  { return b.metricsTap }
func (b *MembershipNode) MergeEvents() <-chan MergeEvent    { return b.merges }
func (b *MembershipNode) Events() <-chan Event              { return b.events }
func (b *MembershipNode) SelfID() string                    { return b.selfID }
func (b *MembershipNode) ConnMgr() *connections.ConnManager { return b.connMgr }
func (b *MembershipNode) AlivePeers() []string              { return b.view.AlivePeerIDs() }
func (n *MembershipNode) View() *View                       { return n.view }

func (b *MembershipNode) Start(ctx context.Context, seeds []string) error {
	// Only launches client-side seeding, not any gRPC server
	go b.seedClient(ctx, seeds)
	return nil
}

func (b *MembershipNode) Stop() {
	// Clean up any resources if needed.
}

func (b *MembershipNode) Send(m *pb.Metric) {
	msg := &pb.ClusterMsg{Payload: &pb.ClusterMsg_Metric{Metric: m}}
	b.connsMu.RLock()
	for _, pc := range b.conns {
		utils.NonBlockingSend(pc.send, msg)
	}
	b.connsMu.RUnlock()
	utils.NonBlockingSend(b.metricsTap, m)
}

func (b *MembershipNode) addConn(id string, pc *peerConn) {
	b.connsMu.Lock()
	b.conns[id] = pc
	b.connsMu.Unlock()
	utils.NonBlockingSend(pc.send, &pb.ClusterMsg{
		Payload: &pb.ClusterMsg_Digest{Digest: b.view.ToDigest()},
	})
}

func (b *MembershipNode) handleMsg(src string, msg *pb.ClusterMsg) {
	switch p := msg.Payload.(type) {
	case *pb.ClusterMsg_Digest:
		if b.view.MergeDigest(p.Digest) {
			utils.NonBlockingSend(b.merges, MergeEvent{Source: src, Count: len(p.Digest.Peers)})
			go b.broadcastDigest()
		}
		b.connMgr.MergeDigest(p.Digest, b.selfID)
	case *pb.ClusterMsg_Metric:
		utils.NonBlockingSend(b.metricsCh, p.Metric)
		utils.NonBlockingSend(b.metricsTap, p.Metric)
	}
}

// Option functions for configuring MembershipNode
func WithTickInterval(d time.Duration) MembershipOption {
	return func(b *MembershipNode) { b.tickInterval = d }
}
func WithDialerInterval(d time.Duration) MembershipOption {
	return func(b *MembershipNode) { b.dialerInterval = d }
}
func WithIncarnation(inc uint64) MembershipOption {
	return func(b *MembershipNode) { b.incarnation = inc }
}
func WithStore(s store.Driver) MembershipOption {
	return func(b *MembershipNode) { b.store = s }
}

// Broadcasts the current view digest to all connected peers.
func (b *MembershipNode) broadcastDigest() {
	b.connsMu.RLock()
	defer b.connsMu.RUnlock()
	digest := b.view.ToDigest()
	msg := &pb.ClusterMsg{Payload: &pb.ClusterMsg_Digest{Digest: digest}}
	for _, pc := range b.conns {
		utils.NonBlockingSend(pc.send, msg)
	}
}
