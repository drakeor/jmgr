// internal/cluster/integration_test.go

package cluster

import (
	"context"
	"net"
	"testing"
	"time"

	pb "github.com/drakeor/jmgr/api"
	"github.com/drakeor/jmgr/internal/connections"
	"github.com/drakeor/jmgr/internal/membership"
	"github.com/drakeor/jmgr/internal/store"
	"github.com/drakeor/jmgr/internal/utils"
	"google.golang.org/grpc"
)

const (
	testTick       = 100 * time.Millisecond
	testTickDialer = 300 * time.Millisecond
)

type testClusterNode struct {
	id      string
	addr    string
	server  *grpc.Server
	cluster *Server
	memb    *membership.MembershipNode
	connMgr *connections.ConnManager
	store   store.Driver
	stop    context.CancelFunc
}

func startTestClusterNode(t *testing.T, id string, addr string, seeds []string, ctx context.Context, inc uint64) (*testClusterNode, error) {
	connMgr := connections.NewConnManager()
	st, _ := store.NewBadgerRing(id, nil, "", 1<<20, true)
	memb := membership.NewMembershipNode(
		id,
		addr,
		connMgr,
		membership.WithTickInterval(testTick),
		membership.WithDialerInterval(testTickDialer),
		membership.WithStore(st),
		membership.WithIncarnation(inc),
	)
	membershipHandler := &MembershipHandler{Driver: memb}
	storeHandler := &StoreHandler{Store: st}
	antiHandler := &AntiEntropyHandler{Store: st}

	clusterSrv := &Server{
		Membership:  membershipHandler,
		Store:       storeHandler,
		AntiEntropy: antiHandler,
	}
	grpcSrv := grpc.NewServer()
	pb.RegisterClusterServer(grpcSrv, clusterSrv)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	go grpcSrv.Serve(lis)

	// Wait a short time to ensure server is up
	time.Sleep(100 * time.Millisecond)

	// Now start membership, passing in real seeds (other nodes' addrs)
	if err := memb.Start(ctx, seeds); err != nil {
		return nil, err
	}

	return &testClusterNode{
		id:      id,
		addr:    addr,
		server:  grpcSrv,
		cluster: clusterSrv,
		memb:    memb,
		connMgr: connMgr,
		store:   st,
		stop: func() {
			grpcSrv.GracefulStop()
			st.Close()
		},
	}, nil
}

func stopTestClusterNode(node *testClusterNode) {
	node.stop()
	node.server.GracefulStop()
	// node.store.Close()
}

func getFreePort() string {
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	defer lis.Close()
	return lis.Addr().String()
}

func TestTombstonePropagationOnRejoin(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	addrA := getFreePort()
	addrB := getFreePort()

	a, err := startTestClusterNode(t, "a", addrA, nil, ctx, 1)
	if err != nil {
		t.Fatalf("failed to start node a: %v", err)
	}
	b, err := startTestClusterNode(t, "b", addrB, []string{addrA}, ctx, 1)
	if err != nil {
		t.Fatalf("failed to start node b: %v", err)
	}

	utils.AssertEventually(t, 3*time.Second, func() bool {
		return contains(a.memb.AlivePeers(), "b") && contains(b.memb.ConnMgr().Peers(), "a")
	})

	b.stop()
	time.Sleep(200 * time.Millisecond)
	a.memb.View().MergeDigest(&pb.MembershipDigest{Peers: []*pb.PeerInfo{
		{Id: "b", Incarnation: 2, Tombstone: true},
	}})
	time.Sleep(500 * time.Millisecond)

	utils.AssertEventually(t, 2*time.Second, func() bool {
		return !contains(a.memb.AlivePeers(), "b")
	})

	// Restart b, simulating a higher incarnation
	incarnation := uint64(3) // Must be > tombstone incarnation (which you set to 2)
	b, err = startTestClusterNode(t, "b", addrB, []string{addrA}, ctx, incarnation)
	if err != nil {
		t.Fatalf("failed to restart node b: %v", err)
	}

	utils.AssertEventually(t, 5*time.Second, func() bool {
		return contains(a.memb.AlivePeers(), "b") && contains(b.memb.ConnMgr().Peers(), "a")
	})

	a.stop()
	b.stop()
	time.Sleep(100 * time.Millisecond)
}

func TestIndirectPeerDiscovery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Pick static but unique ports for A/B/C or use ":0" and parse the assigned ports
	aAddr := "127.0.0.1:50001"
	bAddr := "127.0.0.1:50002"
	cAddr := "127.0.0.1:50003"

	a, err := startTestClusterNode(t, "a", aAddr, []string{bAddr}, ctx, 1)
	if err != nil {
		t.Fatalf("Failed to start node A: %v", err)
	}
	//require.NoError(t, err)
	b, err := startTestClusterNode(t, "b", bAddr, nil, ctx, 1)
	if err != nil {
		t.Fatalf("Failed to start node B: %v", err)
	}
	//require.NoError(t, err)
	c, err := startTestClusterNode(t, "c", cAddr, []string{bAddr}, ctx, 1)
	if err != nil {
		t.Fatalf("Failed to start node C: %v", err)
	}
	//require.NoError(t, err)

	defer a.stop()
	defer b.stop()
	defer c.stop()

	// Now run your peer discovery assertions as before, using AssertEventually, etc.
	utils.AssertEventually(t, 5*time.Second, func() bool {
		return contains(a.memb.AlivePeers(), "c") && contains(c.memb.ConnMgr().Peers(), "a")
	})
}
func TestPeerReconnection(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	addrA := getFreePort()
	addrB := getFreePort()

	a, err := startTestClusterNode(t, "a", addrA, nil, ctx, 1)
	if err != nil {
		t.Fatalf("failed to start node a: %v", err)
	}
	b, err := startTestClusterNode(t, "b", addrB, []string{addrA}, ctx, 1)
	if err != nil {
		t.Fatalf("failed to start node b: %v", err)
	}

	utils.AssertEventually(t, 3*time.Second, func() bool {
		return contains(a.memb.AlivePeers(), "b") && contains(b.memb.ConnMgr().Peers(), "a")
	})

	b.stop()
	time.Sleep(200 * time.Millisecond)

	b, err = startTestClusterNode(t, "b", addrB, []string{addrA}, ctx, 1)
	if err != nil {
		t.Fatalf("failed to restart node b: %v", err)
	}

	utils.AssertEventually(t, 5*time.Second, func() bool {
		return contains(a.memb.AlivePeers(), "b") && contains(b.memb.ConnMgr().Peers(), "a")
	})

	a.stop()
	b.stop()
	time.Sleep(100 * time.Millisecond)
}

func TestClusterToleratesNodeUnreachable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	addrA := getFreePort()
	addrB := getFreePort()
	addrC := getFreePort()

	a, err := startTestClusterNode(t, "a", addrA, nil, ctx, 1)
	if err != nil {
		t.Fatalf("failed to start node a: %v", err)
	}
	b, err := startTestClusterNode(t, "b", addrB, []string{addrA}, ctx, 1)
	if err != nil {
		t.Fatalf("failed to start node b: %v", err)
	}
	c, err := startTestClusterNode(t, "c", addrC, []string{addrA}, ctx, 1)
	if err != nil {
		t.Fatalf("failed to start node c: %v", err)
	}

	utils.AssertEventually(t, 3*time.Second, func() bool {
		return contains(a.memb.AlivePeers(), "b") && contains(a.memb.AlivePeers(), "c")
	})

	b.stop()
	time.Sleep(200 * time.Millisecond)

	utils.AssertEventually(t, 3*time.Second, func() bool {
		return contains(a.memb.AlivePeers(), "c") && contains(c.memb.ConnMgr().Peers(), "a")
	})

	b, err = startTestClusterNode(t, "b", addrB, []string{addrA}, ctx, 1)
	if err != nil {
		t.Fatalf("failed to restart node b: %v", err)
	}

	utils.AssertEventually(t, 5*time.Second, func() bool {
		return contains(a.memb.AlivePeers(), "b") && contains(c.memb.ConnMgr().Peers(), "b")
	})

	a.stop()
	b.stop()
	c.stop()
	time.Sleep(100 * time.Millisecond)
}

func TestPartitionHealMergesMembership(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	addrA := getFreePort()
	addrB := getFreePort()
	addrC := getFreePort()

	a, err := startTestClusterNode(t, "a", addrA, nil, ctx, 1)
	if err != nil {
		t.Fatalf("failed to start node a: %v", err)
	}
	b, err := startTestClusterNode(t, "b", addrB, []string{addrA}, ctx, 1)
	if err != nil {
		t.Fatalf("failed to start node b: %v", err)
	}

	utils.AssertEventually(t, 3*time.Second, func() bool {
		return contains(a.memb.AlivePeers(), "b") && contains(b.memb.ConnMgr().Peers(), "a")
	})

	c, err := startTestClusterNode(t, "c", addrC, []string{addrB}, ctx, 1)
	if err != nil {
		t.Fatalf("failed to start node c: %v", err)
	}

	c.stop()
	time.Sleep(200 * time.Millisecond)

	c, err = startTestClusterNode(t, "c", addrC, []string{addrA, addrB}, ctx, 1)
	if err != nil {
		t.Fatalf("failed to restart node c: %v", err)
	}

	utils.AssertEventually(t, 5*time.Second, func() bool {
		ok := contains(a.memb.AlivePeers(), "b") && contains(a.memb.AlivePeers(), "c") &&
			contains(b.memb.ConnMgr().Peers(), "a") && contains(b.memb.ConnMgr().Peers(), "c") &&
			contains(c.memb.ConnMgr().Peers(), "a") && contains(c.memb.ConnMgr().Peers(), "b")
		return ok
	})

	a.stop()
	b.stop()
	c.stop()
	time.Sleep(100 * time.Millisecond)
}

func contains(xs []string, x string) bool {
	for _, s := range xs {
		if s == x {
			return true
		}
	}
	return false
}

// (Repeat for other scenarios: PeerReconnection, PartitionHeal, etc.)
