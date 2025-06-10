package antientropy

import (
	"context"
	"log"
	"time"

	pb "github.com/drakeor/jmgr/api"
	"github.com/drakeor/jmgr/internal/connections"
	"github.com/drakeor/jmgr/internal/store"
	"github.com/drakeor/jmgr/internal/utils"
)

// SyncEngine is a job queue for syncing metric bins from peers.
type SyncEngine struct {
	Queue chan SyncJob
	quit  chan struct{}
}

// NewSyncEngine: create with hard queue size; workers are launched via Run()
// Note that exceeding the queue size will drop jobs silently.
// TODO: Maybe change that up later... might be bad design
func NewSyncEngine(queueSize int) *SyncEngine {
	return &SyncEngine{
		Queue: make(chan SyncJob, queueSize),
		quit:  make(chan struct{}),
	}
}

// Run launches N workers, calling doJob for each job pulled from Queue.
// Note the parallelism is controlled by the caller, so you can run multiple
// SyncEngines with different parallelism levels if needed.
// However, I don't handle the resource aspect of this yet... so not recommended.
// TODO: Analyze impact of parallelism
func (se *SyncEngine) Run(ctx context.Context, parallelism int, doJob func(SyncJob)) {
	for i := 0; i < parallelism; i++ {
		go func() {
			for {
				select {
				case job := <-se.Queue:
					doJob(job)
				case <-ctx.Done():
					return
				case <-se.quit:
					return
				}
			}
		}()
	}
}

// Enqueue adds a job to the queue.
// Returns true if the job was enqueued, false if the queue was full.
func (se *SyncEngine) Enqueue(job SyncJob) bool {
	select {
	case se.Queue <- job:
		return true
	default:
		log.Printf("[syncengine] queue full, dropping sync job: %+v", job)
		return false
	}
}

// Stop stops the SyncEngine gracefully
func (se *SyncEngine) Stop() {
	close(se.quit)
}

// The DefaultSyncJobRunner fetches bin from peer and writes to store.
// I was going to have different runners for different protocols, but
// for now, this is sufficient...
// Note this returns a function
func DefaultSyncJobRunner(localStore store.Driver, connMgr *connections.ConnManager) func(SyncJob) {

	return func(job SyncJob) {

		// Get connection to the peer
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()
		addr, ok := connMgr.AddrForPeer(job.PeerID)
		if !ok || addr == "" {
			log.Printf("[syncengine] no address for peer %q", job.PeerID)
			return
		}
		conn, err := connMgr.Get(ctx, job.PeerID, addr, connections.TrafficSync)
		if err != nil {
			log.Printf("[syncengine] connMgr.Get: %v", err)
			return
		}
		client := pb.NewClusterClient(conn)

		// Fetch the metric points for the given bucket
		resp, err := client.GetMetricPoints(ctx, &pb.GetMetricRequest{Bucket: job.Bucket})
		if err != nil {
			log.Printf("[syncengine] fetch from peer %q failed: %v", job.PeerID, err)
			return
		}
		// Parse the bucket key to get nodeID, name, and labels
		nodeID, name, labels, err := utils.ParseBucketKey(job.Bucket)
		if err != nil {
			log.Printf("[syncengine] bad bucket key: %v", err)
			return
		}

		// Filter points by bin
		binStart := parseBinID(job.Bin)
		binEnd := binStart.Add(BinInterval)
		var points []*pb.MetricPoint
		for _, pt := range resp.Points {
			ts := time.UnixMilli(pt.TsMs)
			if !ts.Before(binStart) && ts.Before(binEnd) {
				points = append(points, pt)
			}
		}
		if len(points) == 0 {
			log.Printf("[syncengine] no points for bin %q from %q", job.Bin, job.PeerID)
			return
		}

		// Create the metric and store it
		metric := &pb.Metric{
			NodeId: nodeID, Name: name, Labels: labels, Points: points, Source: job.PeerID, Tier: "hot",
		}
		if err := localStore.Put(metric); err != nil {
			log.Printf("[syncengine] store.Put failed: %v", err)
		} else {
			log.Printf("[syncengine] synced bin %q (%d points) from %q", job.Bin, len(points), job.PeerID)
		}
	}
}

// helper function to parse BinID into time.Time.
func parseBinID(binID BinID) time.Time {
	t, err := time.Parse("20060102T1504", string(binID))
	if err != nil {
		return time.Time{} // fallback: zero time
	}
	return t
}
