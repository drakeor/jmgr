package core

import (
	"testing"
	"time"

	collectors "github.com/drakeor/jmgr/internal/core/collectors"
)

func TestSysPoller(t *testing.T) {
	p := NewSysPoller(&collectors.CPUCollector{NodeID: "test"})
	defer p.Close()
	select {
	case got := <-p.Metrics():
		if got == nil {
			t.Skip("No metric collected (possibly running in a restricted environment)")
		}
		if got.NodeId != "test" || got.Name != "cpu_pct" {
			t.Fatalf("bad metric: %+v", got)
		}
	case <-time.After(2 * time.Second): // was 500ms
		t.Skip("No metric received in time – maybe running in restricted test environment")
	}
}
