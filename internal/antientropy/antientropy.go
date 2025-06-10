package antientropy

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	pb "github.com/drakeor/jmgr/api"
	"github.com/drakeor/jmgr/internal/connections"
	"github.com/drakeor/jmgr/internal/store"
	"github.com/drakeor/jmgr/internal/utils"
)

// Service struct for the anti-entropy service.
type Service struct {
	selfID     string
	store      store.Driver
	connMgr    *connections.ConnManager
	period     time.Duration
	syncEngine *SyncEngine // from queue.go
}

// Create a new anti-entropy service.
func NewService(selfID string, store store.Driver, connMgr *connections.ConnManager, period time.Duration, syncEngine *SyncEngine) *Service {
	return &Service{
		selfID:     selfID,
		store:      store,
		connMgr:    connMgr,
		period:     period,
		syncEngine: syncEngine,
	}
}

// Run starts the anti-entropy service, periodically syncing with peers.
// Can be ran independently async'd in a goroutine.
func (s *Service) Run(ctx context.Context) {
	ticker := time.NewTicker(s.period)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.syncOnce(ctx)
		}
	}
}

// syncOnce does one anti-entropy cycle for all peers.
func (s *Service) syncOnce(ctx context.Context) {
	for _, peerID := range s.connMgr.Peers() {
		if peerID == s.selfID || peerID == "" {
			continue
		}
		go s.syncWithPeer(ctx, peerID)
	}
}

// syncWithPeer: Orchestrates a full round with a peer.This consists of the following steps:
// 1. Fetch peer's summary
// 2. Get local summary
// 3. Diff
// 4. Queue sync jobs.
func (s *Service) syncWithPeer(ctx context.Context, peerID string) {

	// 0. Grab the address and connection for the peer.
	addr, ok := s.connMgr.AddrForPeer(peerID)
	if !ok || addr == "" {
		log.Printf("[anti-entropy] unknown peer or missing address %q (%q)", peerID, addr)
		return
	}
	log.Printf("[anti-entropy] syncing with peer %s at address %s", peerID, addr)
	conn, err := s.connMgr.Get(ctx, peerID, addr, connections.TrafficSync)
	if err != nil {
		log.Printf("[anti-entropy] connMgr.Get %s: %v", addr, err)
		return
	}
	client := pb.NewClusterClient(conn)

	// 1. Fetch peer's NodeSummary (over new RPC call)
	peerSummary, err := s.getPeerSummary(ctx, client)
	if err != nil {
		log.Printf("[anti-entropy] getPeerSummary %s: %v", addr, err)
		return
	}
	log.Printf("[anti-entropy] peer %s summary: %d buckets", peerID, len(peerSummary))

	// 2. Get local summary (from our store)
	localSummary := SummarizeStore(s.store)
	log.Printf("[anti-entropy] local node %s summary: %d buckets", s.selfID, len(localSummary))

	// 3. Diff: compute what to sync
	jobs := DiffSummaries(localSummary, peerSummary, peerID)
	log.Printf("[anti-entropy] %d sync jobs needed with peer %s", len(jobs), peerID)

	// 4. Enqueue jobs to sync engine
	for _, job := range jobs {
		if !s.syncEngine.Enqueue(job) {
			log.Printf("[anti-entropy] sync queue full, dropping job %+v", job)
		}
		log.Printf("[anti-entropy] enqueued sync job for peer %s: %+v", peerID, job)

		// Hacky rate limiting to avoid overwhelming peers.
		// TODO: Create a better method based off data size or peer capacity.
		time.Sleep(50 * time.Millisecond)
	}
}

// getPeerSummary exchanges a summary with a peer (placeholder).
func (s *Service) getPeerSummary(ctx context.Context, client pb.ClusterClient) (NodeSummary, error) {
	resp, err := client.GetNodeSummary(ctx, &pb.Empty{})
	if err != nil {
		return nil, err
	}
	return FromProtoNodeSummary(resp), nil
}

// ParseBucketKey splits "nodeID|name|labels" into parts.
func ParseBucketKey(bucket string) (nodeID, name string, labels map[string]string, err error) {
	parts := strings.SplitN(bucket, "|", 3)
	if len(parts) != 3 {
		return "", "", nil, fmt.Errorf("bad bucket: %q", bucket)
	}
	nodeID, name, labelStr := parts[0], parts[1], parts[2]
	labels = utils.DeserializeLabels(labelStr)
	return nodeID, name, labels, nil
}
