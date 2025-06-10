package membership

import (
	"context"
	"time"

	pb "github.com/drakeor/jmgr/api"
	"github.com/drakeor/jmgr/internal/connections"
)

// Add seeds to ConnManager and start the membership dialer loop.
func (b *MembershipNode) seedClient(ctx context.Context, seeds []string) {
	// Add all seeds to ConnManager
	for _, addr := range seeds {
		if addr == "" || addr == b.addr {
			continue // skip self & blanks
		}
		peerID := addr // If seeds have only addresses, treat address as ID (fix later if you have mapping)
		b.connMgr.AddPeer(peerID, addr)
	}
	go b.membershipDialerLoop(ctx)
}

// membershipDialerLoop periodically checks for peers
// and ensures a membership stream is established.
func (b *MembershipNode) membershipDialerLoop(ctx context.Context) {
	ticker := time.NewTicker(b.dialerInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, peerID := range b.connMgr.Peers() {
				if peerID == b.selfID || peerID == "" {
					continue
				}
				go b.ensureMembershipStream(ctx, peerID)
			}
		}
	}
}

// Implementation of ensureMembershipStream ensures that a membership stream
func (b *MembershipNode) ensureMembershipStream(ctx context.Context, peerID string) {

	// Get the address for the peer
	addr, ok := b.connMgr.AddrForPeer(peerID)
	if !ok || addr == "" || peerID == b.selfID {
		return
	}
	conn, err := b.connMgr.Get(ctx, peerID, addr, connections.TrafficHeartbeat)
	if err != nil {
		time.Sleep(2 * time.Second)
		return
	}
	client := pb.NewClusterClient(conn)

	// Attempt to join the cluster
	resp, _ := client.Join(ctx, &pb.JoinRequest{NodeId: b.selfID, Addr: b.addr})
	if resp != nil {
		for _, pin := range resp.Peers {
			if pin.Id == b.selfID || pin.Addr == "" {
				continue
			}
			b.view.MergeDigest(&pb.MembershipDigest{Peers: []*pb.PeerInfo{pin}})
			// Make sure all discovered peers are registered in ConnMgr
			b.connMgr.AddPeer(pin.Id, pin.Addr)
		}
	}

	// Ensure we have a stream to the peer
	stream, err := client.Stream(ctx)
	if err != nil {
		time.Sleep(2 * time.Second)
		return
	}

	pc := newPeerConn(peerID, stream)
	b.addConn(peerID, pc)

	_ = stream.Send(&pb.ClusterMsg{Payload: &pb.ClusterMsg_Digest{Digest: b.view.ToDigest()}})
	pc.run(ctx, b)
	<-pc.done // block until broken, then reconnect
}
