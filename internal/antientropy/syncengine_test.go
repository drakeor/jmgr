package antientropy

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestSyncEngineEnqueueAndRun(t *testing.T) {
	engine := NewSyncEngine(2)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var seen []SyncJob
	var mu sync.Mutex
	go engine.Run(ctx, 2, func(job SyncJob) {
		mu.Lock()
		defer mu.Unlock()
		seen = append(seen, job)
	})

	job := SyncJob{PeerID: "peer1", Bucket: "bucketA", Bin: "bin1"}
	if !engine.Enqueue(job) {
		t.Fatalf("enqueue should succeed")
	}

	// Wait a bit for worker to process
	time.Sleep(100 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if len(seen) != 1 || seen[0] != job {
		t.Errorf("expected to see job processed, got %+v", seen)
	}
}
