package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/guptarohit/asciigraph"

	pb "github.com/drakeor/jmgr/api"
	"google.golang.org/grpc"
)

// Parameters
const (
	historyLen  = 30
	graphHeight = 6
)

// Hardcoded node names for the demo
// Also because this UI sucks and I'm done trying to make it dynamic
var (
	nodeNames = []string{"node1", "node2", "node3", "node4"}
)

// nodeMetrics holds the history of metrics for a node
type nodeMetrics struct {
	cpuHistory  []float64
	memHistory  []float64
	diskHistory []float64
	points      int
}

// model holds the state of the TUI application
type model struct {
	nodes      map[string]*nodeMetrics
	grpcConn   *grpc.ClientConn
	grpcClient pb.ClusterClient
}

func initialModel(client pb.ClusterClient, conn *grpc.ClientConn) model {
	m := model{
		nodes:      make(map[string]*nodeMetrics),
		grpcConn:   conn,
		grpcClient: client,
	}
	for _, n := range nodeNames {
		m.nodes[n] = &nodeMetrics{
			cpuHistory:  make([]float64, 0, historyLen),
			memHistory:  make([]float64, 0, historyLen),
			diskHistory: make([]float64, 0, historyLen),
			points:      0,
		}
	}
	return m
}

// tickMsg is used to trigger periodic updates in the TUI
type tickMsg struct{}

// Init initializes the TUI model and starts the periodic tick updates
func (m model) Init() tea.Cmd {
	return tea.Tick(time.Second/2, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

// lastNValues extracts the last N values from a slice of MetricPoints
// Useful for limiting the history displayed in the graphs
func lastNValues(pts []*pb.MetricPoint, n int) []float64 {
	var vals []float64
	for i := len(pts) - n; i < len(pts); i++ {
		if i >= 0 && i < len(pts) {
			vals = append(vals, pts[i].Value)
		}
	}
	return vals
}

// Core update function that handles messages and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		for _, n := range nodeNames {
			node := m.nodes[n]

			// TODO: This is terribly hacky coded and a lot of copy paste.
			// But again, this is a demo and I don't want to spend more time with this crappy UI.

			// CPU metrics
			cpuReq := &pb.GetMetricRequest{Bucket: fmt.Sprintf("%s|cpu_pct|", n)}
			cpuResp, err := m.grpcClient.GetMetricPoints(ctx, cpuReq)
			if err == nil {
				node.cpuHistory = lastNValues(cpuResp.Points, historyLen)
				node.points = len(cpuResp.Points)
			} else {
				node.cpuHistory = nil
				node.points = 0
			}

			// Memory metrics
			memReq := &pb.GetMetricRequest{Bucket: fmt.Sprintf("%s|mem_free_bytes|", n)}
			memResp, err := m.grpcClient.GetMetricPoints(ctx, memReq)
			if err == nil {
				raw := lastNValues(memResp.Points, historyLen)
				node.memHistory = make([]float64, len(raw))
				for i, v := range raw {
					node.memHistory[i] = v / (1024 * 1024 * 1024) // bytes to GB
				}
			} else {
				node.memHistory = nil
			}

			// Disk metrics
			diskReq := &pb.GetMetricRequest{Bucket: fmt.Sprintf("%s|disk_free_bytes|mount=/", n)}
			diskResp, err := m.grpcClient.GetMetricPoints(ctx, diskReq)
			if err == nil {
				raw := lastNValues(diskResp.Points, historyLen)
				node.diskHistory = make([]float64, len(raw))
				for i, v := range raw {
					node.diskHistory[i] = v / (1024 * 1024 * 1024 * 1024) // bytes to TB
				}
			} else {
				node.diskHistory = nil
			}

		}

		// Trigger a redraw
		return m, tea.Tick(time.Second/2, func(t time.Time) tea.Msg { return tickMsg{} })

	// Handle key messages (only used for quitting right now	)
	case tea.KeyMsg:
		if msg.String() == "q" {
			return m, tea.Quit
		}
	}
	return m, nil
}

// Entry point for the ascii graph rendering
// Manually set dimensions because I hate the default auto-sizing
func graphOrWait(data []float64, caption string) string {
	if len(data) == 0 {
		return "(waiting for data...)"
	}
	return asciigraph.Plot(data, asciigraph.Height(graphHeight), asciigraph.Width(16), asciigraph.Caption(caption))
}

// Okay, setting up the main view function
func (m model) View() string {
	var rows []string

	// Header
	header := lipgloss.NewStyle().Bold(true).Render
	rows = append(rows, header(fmt.Sprintf("%-8s │ %-24s │ %-24s │ %-24s", "NODE", "CPU (%)", "MEM (GB)", "DISK (TB)")))

	// Print details for each node
	for _, n := range nodeNames {
		node := m.nodes[n]

		// Thanks to Gemini for getting this to -actually- work.
		cpuGraph := graphOrWait(node.cpuHistory, "")
		memGraph := graphOrWait(node.memHistory, "")
		diskGraph := graphOrWait(node.diskHistory, "")

		cpuLines := strings.Split(cpuGraph, "\n")
		memLines := strings.Split(memGraph, "\n")
		diskLines := strings.Split(diskGraph, "\n")

		maxLines := max(len(cpuLines), len(memLines), len(diskLines))

		cpuLines = padLines(cpuLines, maxLines)
		memLines = padLines(memLines, maxLines)
		diskLines = padLines(diskLines, maxLines)

		for i := 0; i < maxLines; i++ {
			nodeLabel := ""
			if i == 0 {
				nodeLabel = n
			}
			rows = append(rows, fmt.Sprintf(
				"%-8s │ %-24s │ %-24s │ %-24s",
				nodeLabel,
				cpuLines[i],
				memLines[i],
				diskLines[i],
			))
		}
		rows = append(rows, fmt.Sprintf("%-8s │ %-24s │ %-24s │ %-24s", "", fmt.Sprintf("Points: %d", node.points), "", ""))
		rows = append(rows, "")
	}

	rows = append(rows, "\nPress 'q' to quit.")
	return strings.Join(rows, "\n")
}

// Helper function to pad lines with empty strings
func padLines(lines []string, n int) []string {
	for len(lines) < n {
		lines = append(lines, "")
	}
	return lines
}

// max returns the maximum of three integers
func max(a, b, c int) int {
	if a > b {
		if a > c {
			return a
		}
		return c
	}
	if b > c {
		return b
	}
	return c
}

// Finally, the main function to set up the gRPC connection and start the TUI
func main() {
	rand.Seed(time.Now().UnixNano())

	addr := "127.0.0.1:4000" // node1's gRPC address
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("Failed to dial node1 gRPC: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	client := pb.NewClusterClient(conn)

	p := tea.NewProgram(initialModel(client, conn))
	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
	}
}
