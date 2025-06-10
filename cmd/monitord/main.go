package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	pb "github.com/drakeor/jmgr/api"
	"github.com/drakeor/jmgr/internal/alert"
	"github.com/drakeor/jmgr/internal/antientropy"
	"github.com/drakeor/jmgr/internal/cluster"
	"github.com/drakeor/jmgr/internal/connections"
	"github.com/drakeor/jmgr/internal/core"
	"github.com/drakeor/jmgr/internal/core/collectors"
	"github.com/drakeor/jmgr/internal/membership"
	"github.com/drakeor/jmgr/internal/store"
	"google.golang.org/grpc"
)

func main() {

	// Command-line flags
	selfID := flag.String("id", hostname(), "unique node ID (defaults to host)")
	addr := flag.String("addr", ":4000", "gRPC listen address")
	join := flag.String("join", "", "comma-sep list of seed addresses")
	dataDir := flag.String("data-dir", "/var/lib/jmgr", "Badger directory")
	maxBytes := flag.Int64("disk-limit", 1<<30, "per-node DB size cap (bytes)")
	inMem := flag.Bool("in-mem", false, "use in-memory Badger (testing)")
	verbose := flag.Bool("v", false, "verbose logging")
	alertEnabled := flag.Bool("alert", true, "enable alerting")
	flag.Parse()

	// Handle termination signals
	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Start pprof server (for profiling/debugging)
	// TODO: Make this a flag option
	go func() {
		log.Println("[pprof] listening on :6060")
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// Set up pollers of interest
	poller := core.NewSysPoller(
		&collectors.CPUCollector{NodeID: *selfID},
		&collectors.DiskFreeCollector{NodeID: *selfID, Mount: "/"},
		&collectors.MemFreeCollector{NodeID: *selfID},
	)

	// Create the alert service (if enabled)
	var alertSvc *alert.Service
	if *alertEnabled {
		log.Println("[alert] alerting enabled")
		alerter := alert.NewFileAlerter("bin/alerts.log")
		thresholds := alert.Thresholds{
			CPUPercent:      1.0,                     // Set really low for testing
			MemoryUsedBytes: 16 * 1024 * 1024 * 1024, // 16 GB
			DiskFreeBytes:   10 * 1024 * 1024 * 1024, // 10 GB free
		}
		alertSvc = alert.NewService(thresholds, alerter, 1*time.Second)
	}

	// Create the Badger ring store
	ring, err := store.NewBadgerRing(
		*selfID,
		nil, // peers are auto-discovered, no initial peers
		*dataDir,
		*maxBytes,
		*inMem,
	)
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	defer ring.Close()

	// Load the ring and get the next incarnation of this process (handles restarts)
	inc, err := ring.NextIncarnation()
	if err != nil {
		log.Fatalf("could not update/load incarnation: %v", err)
	}

	// Create connections manager
	connMgr := connections.NewConnManager()

	// Create and start the membership broker
	broker := membership.NewMembershipNode(*selfID, *addr, connMgr,
		membership.WithIncarnation(inc),
		membership.WithStore(ring),
	)

	// Start the broker
	if err := broker.Start(ctx, splitNonEmpty(*join)); err != nil {
		log.Fatalf("broker: %v", err)
	}
	defer broker.Stop()

	// Create the gRPC cluster server and register handlers
	membershipHandler := &cluster.MembershipHandler{Driver: broker}
	storeHandler := &cluster.StoreHandler{Store: ring}
	antiHandler := &cluster.AntiEntropyHandler{Store: ring}

	clusterSrv := &cluster.Server{
		Membership:  membershipHandler,
		Store:       storeHandler,
		AntiEntropy: antiHandler,
	}
	grpcSrv := grpc.NewServer()
	pb.RegisterClusterServer(grpcSrv, clusterSrv)

	// Register the gRPC server with the connections manager
	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", *addr, err)
	}
	go func() {
		log.Printf("[grpc] serving cluster gRPC API on %s", *addr)
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatalf("grpc server failed: %v", err)
		}
	}()

	// Start the core loop
	loop := core.NewLoop(poller, ring, broker, alertSvc)

	// If verbose logging is enabled, set up logging for peers and metric events as goroutines
	if *verbose {
		go func() {
			snap := ""
			for range time.Tick(2 * time.Second) {
				peers := strings.Join(broker.AlivePeers(), ",")
				if peers != snap {
					log.Printf("[peers] %s", peers)
					snap = peers
				}
			}
		}()

		go func() {
			for ev := range broker.MergeEvents() {
				log.Printf("[digest] merged %d metrics from %s", ev.Count, ev.Source)
			}
		}()

		go func() {
			for m := range broker.MetricsEvents() {
				v := 0.0
				if len(m.Points) > 0 {
					v = m.Points[0].Value
				}
				log.Printf("[metrics] %s %s %v %.1f", m.NodeId, m.Name, m.Labels, v)
			}
		}()
	}

	// Antientropy setup
	syncEngine := antientropy.NewSyncEngine(64)
	go syncEngine.Run(ctx, 2, antientropy.DefaultSyncJobRunner(ring, connMgr))
	anti := antientropy.NewService(
		*selfID,
		ring,
		connMgr,
		3*time.Second,
		syncEngine,
	)
	go anti.Run(ctx)

	// Finally, start the core loop
	loop.Start(ctx)
	<-ctx.Done()

	// TODO: Add cleanup for the badger store, connections, etc.
	//anti.Close()
}

// TODO: Move to utils
func hostname() string {
	h, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return h
}

// TODO: Move to utils
func splitNonEmpty(s string) []string {
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
