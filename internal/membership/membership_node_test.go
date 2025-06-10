package membership

import (
	"context"
	"testing"
	"time"

	"github.com/drakeor/jmgr/internal/connections"
	"github.com/drakeor/jmgr/internal/utils"
)

func TestTwoNodeJoin(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	aAddr := "localhost:18001"
	bAddr := "localhost:18002"

	connMgrA := connections.NewConnManager()
	connMgrB := connections.NewConnManager()

	a := NewMembershipNode("a", aAddr, connMgrA)
	if err := a.Start(ctx, nil); err != nil {
		t.Fatalf("start A: %v", err)
	}
	defer a.Stop()

	b := NewMembershipNode("b", bAddr, connMgrB)
	if err := b.Start(ctx, []string{aAddr}); err != nil {
		t.Fatalf("start B: %v", err)
	}
	defer b.Stop()

	utils.AssertEventually(t, 5*time.Second, func() bool {
		return len(a.ConnMgr().Peers()) == 1 && len(b.ConnMgr().Peers()) == 1
	})
}
