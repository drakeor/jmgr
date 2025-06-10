package membership

import (
	"reflect"
	"sort"
	"testing"
	"time"

	pb "github.com/drakeor/jmgr/api"
)

// helper to compare two string slices regardless of order.
func equalIDs(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	as := append([]string(nil), a...)
	bs := append([]string(nil), b...)
	sort.Strings(as)
	sort.Strings(bs)
	return reflect.DeepEqual(as, bs)
}

func TestMergeCommutative(t *testing.T) {
	t.Helper()
	v1, v2 := NewView(), NewView()

	a := &pb.PeerInfo{Id: "a", Incarnation: 1}
	b := &pb.PeerInfo{Id: "b", Incarnation: 1}

	// Each view sees a different peer first.
	v1.MergeDigest(&pb.MembershipDigest{Peers: []*pb.PeerInfo{a}})
	v2.MergeDigest(&pb.MembershipDigest{Peers: []*pb.PeerInfo{b}})

	// Merge in both directions.
	v2.MergeDigest(v1.ToDigest())
	v1.MergeDigest(v2.ToDigest())

	want := []string{"a", "b"}
	if !equalIDs(v1.AlivePeerIDs(), want) {
		t.Fatalf("view1 mismatch: want %v, got %v", want, v1.AlivePeerIDs())
	}
	if !equalIDs(v2.AlivePeerIDs(), want) {
		t.Fatalf("view2 mismatch: want %v, got %v", want, v2.AlivePeerIDs())
	}

	t.Logf("MergeCommutative passed – views: v1=%v v2=%v", v1.AlivePeerIDs(), v2.AlivePeerIDs())
	time.Sleep(100 * time.Millisecond)

}

func TestIncarnationWins(t *testing.T) {
	t.Helper()
	v := NewView()

	old := &pb.PeerInfo{Id: "x", Incarnation: 1, Addr: "old"}
	v.MergeDigest(&pb.MembershipDigest{Peers: []*pb.PeerInfo{old}})

	newer := &pb.PeerInfo{Id: "x", Incarnation: 2, Addr: "new"}
	v.MergeDigest(&pb.MembershipDigest{Peers: []*pb.PeerInfo{newer}})

	got := v.peers["x"].Info
	if got.Addr != "new" || got.Incarnation != 2 {
		t.Fatalf("incarnation did not win: got %+v", got)
	}

	t.Logf("IncarnationWins passed – final peer: %+v", got)
	time.Sleep(100 * time.Millisecond)

}

func TestTombstonePropagation(t *testing.T) {
	t.Helper()
	v1, v2 := NewView(), NewView()
	live := &pb.PeerInfo{Id: "y", Incarnation: 1}
	v1.MergeDigest(&pb.MembershipDigest{Peers: []*pb.PeerInfo{live}})
	v2.MergeDigest(&pb.MembershipDigest{Peers: []*pb.PeerInfo{live}})

	// Tombstone with higher incarnation.
	dead := &pb.PeerInfo{Id: "y", Incarnation: 2, Tombstone: true}
	v1.MergeDigest(&pb.MembershipDigest{Peers: []*pb.PeerInfo{dead}})

	v2.MergeDigest(v1.ToDigest())

	if got := v2.AlivePeerIDs(); len(got) != 0 {
		t.Fatalf("tombstone failed: expected peer to be removed, alive list: %v", got)
	}

	t.Log("TombstonePropagation passed – peer y removed from view2")
	time.Sleep(100 * time.Millisecond)

}
