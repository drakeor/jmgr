package alert

import (
	"fmt"
	"log"
	"sync"
	"time"

	pb "github.com/drakeor/jmgr/api"
)

// High level structs for alerting system
type Service struct {
	thresholds Thresholds
	alerter    Alerter

	mu            sync.Mutex
	lastAlertTime time.Time
	minInterval   time.Duration
}

// Creates a new alerting service with given thresholds, alerter, and minimum interval between alerts.
func NewService(thresholds Thresholds, alerter Alerter, minInterval time.Duration) *Service {
	return &Service{
		thresholds:    thresholds,
		alerter:       alerter,
		minInterval:   minInterval,
		lastAlertTime: time.Time{},
	}
}

// Main processing function that checks metrics against thresholds and sends alerts if exceeded.
func (s *Service) Process(m *pb.Metric) {
	exceeded := false
	var reasons []string

	// TODO: This is super hard-coded. Ideally, we want strings like "cpu_pct" to be defined
	// in a config or constants file... Mispelling would result in alerts not happening.

	log.Printf("[Alert] Processing metric: %s from node %s", m.Name, m.NodeId)
	for _, pt := range m.Points {
		switch m.Name {
		case "cpu_pct":
			if pt.Value > float64(s.thresholds.CPUPercent) {
				exceeded = true
				reasons = append(reasons,
					fmt.Sprintf("CPU %.1f%% > %.1f%%", pt.Value, s.thresholds.CPUPercent))
			}
		case "mem_free_bytes":
			if pt.Value > float64(s.thresholds.MemoryUsedBytes) {
				exceeded = true
				reasons = append(reasons,
					fmt.Sprintf("Memory used %.0f > %.0f", pt.Value, s.thresholds.MemoryUsedBytes))
			}
		case "disk_free_bytes":
			if pt.Value < float64(s.thresholds.DiskFreeBytes) {
				exceeded = true
				reasons = append(reasons,
					fmt.Sprintf("Disk free %.0f < %.0f", pt.Value, s.thresholds.DiskFreeBytes))
			}
		}
	}

	// If any threshold was exceeded, send an alert
	if exceeded {
		s.mu.Lock()
		defer s.mu.Unlock()

		if time.Since(s.lastAlertTime) >= s.minInterval {
			s.lastAlertTime = time.Now()
			subject := fmt.Sprintf("[ALERT] Node %s exceeded thresholds", m.NodeId)
			body := fmt.Sprintf("Node: %s\nMetric: %s\nReasons:\n- %s\nTime: %s\n",
				m.NodeId, m.Name, reasons, s.lastAlertTime.Format(time.RFC3339))
			s.alerter.Send(subject, body)
		}
	}
}
