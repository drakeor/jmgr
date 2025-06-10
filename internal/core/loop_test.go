//go:build mock
// +build mock

package core

import (
	"context"
	"testing"
	"time"

	"github.com/drakeor/jmgr/internal/membership"
	"github.com/drakeor/jmgr/internal/store"
)

func TestLoopEndToEnd(t *testing.T) {
	// in‑proc membership driver (3 brokers) from membership tests
	brokers := membership.NewTestCluster(t, 3) // helper previously defined

	// create stores
	st := store.NewInMem()

	p := NewMockPoller("nodeA", 50*time.Millisecond)

	loop := NewLoop(p, st, brokers[0])

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	loop.Start(ctx)

	<-ctx.Done()

	// Ensure at least one sample stored
	pts, _ := st.Get("nodeA", time.Now().Add(-time.Second))
	if len(pts) == 0 {
		t.Fatalf("expected >=1 stored sample, got 0")
	}
}
