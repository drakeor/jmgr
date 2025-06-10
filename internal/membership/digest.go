package membership

import (
	"sync"
	"time"

	pb "github.com/drakeor/jmgr/api"
	"google.golang.org/protobuf/proto"
)

// Peer wraps the protobuf PeerInfo with runtime metadata.
type Peer struct {
	Info     *pb.PeerInfo
	LastSeen time.Time
}

// View holds our current opinion of the cluster.
// It is safe for concurrent use by multiple goroutines.
type View struct {
	mu    sync.RWMutex
	peers map[string]*Peer // key = peer id
}

// NewView returns an empty view.
func NewView() *View {
	return &View{peers: make(map[string]*Peer)}
}

// ToDigest converts the current view to a protobuf digest for broadcasting.
func (v *View) ToDigest() *pb.MembershipDigest {
	v.mu.RLock()
	defer v.mu.RUnlock()
	out := &pb.MembershipDigest{Peers: make([]*pb.PeerInfo, 0, len(v.peers))}
	for _, p := range v.peers {
		out.Peers = append(out.Peers, p.Info)
	}
	return out
}

// MergeDigest merges an incoming digest. It returns true if the view changed.
func (v *View) MergeDigest(d *pb.MembershipDigest) bool {
	if d == nil {
		return false
	}
	changed := false
	v.mu.Lock()
	defer v.mu.Unlock()
	now := time.Now()
	for _, pin := range d.Peers {
		cur, ok := v.peers[pin.Id]
		switch {
		case !ok:
			// brand‑new peer
			v.peers[pin.Id] = &Peer{Info: clonePeerInfo(pin), LastSeen: now}
			changed = true
		case pin.Incarnation > cur.Info.Incarnation:
			// newer incarnation wins
			cur.Info = clonePeerInfo(pin)
			cur.LastSeen = now
			changed = true
		default:
			// peer exists with same or newer incarnation; just bump LastSeen
			cur.LastSeen = now
		}
	}
	return changed
}

// clonePeerInfo performs a deep copy to avoid races.
func clonePeerInfo(pi *pb.PeerInfo) *pb.PeerInfo {
	if pi == nil {
		return nil
	}
	return proto.Clone(pi).(*pb.PeerInfo)
}

// AlivePeerIDs returns a slice of peer ids that are *not* tombstoned.
func (v *View) AlivePeerIDs() []string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	ids := make([]string, 0, len(v.peers))
	for id, p := range v.peers {
		if !p.Info.Tombstone {
			ids = append(ids, id)
		}
	}
	return ids
}

// PeerInfos returns a slice of peer infos that are *not* tombstoned.
func (v *View) PeerInfos() []*pb.PeerInfo {
	v.mu.RLock()
	defer v.mu.RUnlock()
	out := make([]*pb.PeerInfo, 0, len(v.peers))
	for _, p := range v.peers {
		if !p.Info.Tombstone {
			out = append(out, clonePeerInfo(p.Info))
		}
	}
	return out
}
