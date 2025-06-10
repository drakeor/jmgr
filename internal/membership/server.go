package membership

import (
	"context"
	"fmt"

	pb "github.com/drakeor/jmgr/api"
	"github.com/drakeor/jmgr/internal/utils"
)

// HandleJoin applies a join request to the local membership state.
func (b *MembershipNode) HandleJoin(_ context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error) {
	fmt.Printf("HandleJoin: got join for %q, addr=%q, incarnation=%d\n", req.NodeId, req.Addr, 1)
	info := &pb.PeerInfo{Id: req.NodeId, Addr: req.Addr, Incarnation: 1}

	if b.view.MergeDigest(&pb.MembershipDigest{Peers: []*pb.PeerInfo{info}}) {
		utils.NonBlockingSend(b.events, Event{Type: EventJoin, Peer: info})
		utils.NonBlockingSend(b.merges, MergeEvent{Source: info.Id, Count: 1})
	}
	return &pb.JoinResponse{Peers: b.view.PeerInfos()}, nil

}

// HandleStream handles a remote peer's membership stream to this node.
func (b *MembershipNode) HandleStream(srv pb.Cluster_StreamServer) error {
	first, err := srv.Recv()
	if err != nil {
		return err
	}

	dig, ok := first.Payload.(*pb.ClusterMsg_Digest)
	if !ok || len(dig.Digest.Peers) == 0 {
		return fmt.Errorf("first frame must contain a MembershipDigest")
	}
	peerID := dig.Digest.Peers[0].Id

	pc := newPeerConn(peerID, srv)
	b.addConn(peerID, pc)

	// merge handshake digest + launch goroutines
	b.handleMsg(peerID, first)
	pc.run(srv.Context(), b)

	<-pc.done // block until connection closes
	return nil
}
