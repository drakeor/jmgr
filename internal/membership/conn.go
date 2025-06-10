package membership

import (
	"context"
	"time"

	pb "github.com/drakeor/jmgr/api"
)

// clusterStream is the smallest common subset that both
// pb.Cluster_StreamClient *and* pb.Cluster_StreamServer satisfy.
type clusterStream interface {
	Send(*pb.ClusterMsg) error
	Recv() (*pb.ClusterMsg, error)
	Context() context.Context
}

// peerConn owns one bidirectional stream plus two goroutines.
type peerConn struct {
	id     string
	stream clusterStream // ← was *pb.Cluster_StreamClient
	send   chan *pb.ClusterMsg
	done   chan struct{}
}

// newPeerConn creates a new peerConn instance.
func newPeerConn(id string, s clusterStream) *peerConn {
	return &peerConn{
		id:     id,
		stream: s,
		send:   make(chan *pb.ClusterMsg, 64),
		done:   make(chan struct{}),
	}
}

// run launches reader + writer, each exiting when ctx, pc.done or stream fails.
func (pc *peerConn) run(ctx context.Context, b *MembershipNode) {
	go pc.reader(ctx, b)
	go pc.writer(ctx, b)
}

// Implementation of reader for peerConn
func (pc *peerConn) reader(ctx context.Context, b *MembershipNode) {
	defer close(pc.done)
	for {
		in, err := pc.stream.Recv()
		if err != nil {
			return
		}
		b.handleMsg(pc.id, in)
	}
}

// Implementation of writer for peerConn
func (pc *peerConn) writer(ctx context.Context, b *MembershipNode) {
	ticker := time.NewTicker(b.tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pc.done:
			return
		case <-ticker.C:
			pc.send <- &pb.ClusterMsg{
				Payload: &pb.ClusterMsg_Digest{Digest: b.view.ToDigest()},
			}
		case msg := <-pc.send:
			_ = pc.stream.Send(msg) // ignore send errors—reader will close pc.done
		}
	}
}
