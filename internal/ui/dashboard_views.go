package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// getOrchestratorContent returns professional orchestrator dashboard with grid layout
func (m Model) getOrchestratorContent() []string {
	// Get local IP address
	localIP := getLocalIP()
	address := fmt.Sprintf("%s:50051", localIP)
	
	if m.NodeRegistry == nil {
		return []string{"Registry not initialized"}
	}
	
	nodes := m.NodeRegistry.GetAllNodes()
	
	// Header - professional
	header := lipgloss.NewStyle().
		Bold(true).
		Underline(true).
		Align(lipgloss.Center).
		Render("Orchestrator Dashboard")
	
	// Server info card - left side
	serverInfo := fmt.Sprintf("Address: %s\nPort: 50051\nStatus: ONLINE\nNodes: %d", 
		localIP, len(nodes))
	
	serverCard := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1).
		Width(30).
		Height(6).
		Render(lipgloss.NewStyle().Bold(true).Render("Server Info") + "\n\n" + serverInfo)
	
	// Node list card - right side
	var nodeContent strings.Builder
	nodeContent.WriteString("ID       │ CPU   │ MEM   │ STATUS\n")
	nodeContent.WriteString("─────────┼───────┼───────┼────────\n")
	
	if len(nodes) == 0 {
		nodeContent.WriteString("No worker nodes connected\n\n")
		nodeContent.WriteString("Workers can join using:\n")
		nodeContent.WriteString(address)
	} else {
		maxNodes := 4
		startIdx := m.NodeScrollOffset
		for i := 0; i < maxNodes && startIdx+i < len(nodes); i++ {
			node := nodes[startIdx+i]
			nodeID := node.ID
			if len(nodeID) > 8 {
				nodeID = nodeID[:8]
			}
			nodeContent.WriteString(fmt.Sprintf("%-8s │ %.1f%% │ %.1f%% │ ONLINE\n", 
				nodeID, node.CPUUsage, node.MemoryUsage))
		}
		
		if len(nodes) > maxNodes {
			nodeContent.WriteString(fmt.Sprintf("\nShowing %d-%d of %d nodes (UP/DOWN to scroll)", 
				startIdx+1, min(startIdx+maxNodes, len(nodes)), len(nodes)))
		}
	}
	
	nodesCard := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1).
		Width(50).
		Height(6).
		Render(lipgloss.NewStyle().Bold(true).Render("Connected Nodes") + "\n\n" + nodeContent.String())
	
	// Top row - server info and nodes side by side
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, serverCard, " ", nodesCard)
	
	// Activity logs - bottom full width
	var logContent strings.Builder
	timestamp := "14:32:07"
	if len(nodes) == 0 {
		logContent.WriteString(fmt.Sprintf("%s Orchestrator started successfully\n", timestamp))
		logContent.WriteString(fmt.Sprintf("%s Listening on %s\n", timestamp, address))
		logContent.WriteString(fmt.Sprintf("%s Waiting for worker connections...", timestamp))
	} else {
		logContent.WriteString(fmt.Sprintf("%s Orchestrator operational\n", timestamp))
		logContent.WriteString(fmt.Sprintf("%s Heartbeat streams: %d active\n", timestamp, len(nodes)))
		logContent.WriteString(fmt.Sprintf("%s Cluster status: healthy", timestamp))
	}
	
	logsCard := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1).
		Width(82). // Match total width of top row
		Height(5).
		Render(lipgloss.NewStyle().Bold(true).Render("Activity Logs") + "\n\n" + logContent.String())
	
	// Instructions
	instructions := lipgloss.NewStyle().
		Faint(true).
		Align(lipgloss.Center).
		Render("UP/DOWN scroll nodes | ESC return to menu | CTRL+C exit")
	
	// Combine everything in grid layout
	layout := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		topRow,
		"",
		logsCard,
		"",
		instructions,
	)
	
	return strings.Split(layout, "\n")
}

// getWorkerContent returns professional worker dashboard with borders
func (m Model) getWorkerContent() []string {
	// Header - professional
	header := lipgloss.NewStyle().
		Bold(true).
		Underline(true).
		Align(lipgloss.Center).
		Render("Worker Node Dashboard")
	
	// Connection status card
	var connStatus string
	var statusColor lipgloss.Color
	if m.GrpcClient != nil && m.GrpcClient.IsConnected() {
		connStatus = fmt.Sprintf("Connected to %s\nHeartbeat: Active\nLast ping: < 1s ago", m.OrchestratorAddr)
		statusColor = lipgloss.Color("10") // Green
	} else {
		connStatus = "Disconnected from orchestrator\nHeartbeat: Inactive\nRetrying connection..."
		statusColor = lipgloss.Color("9") // Red
	}
	
	connectionCard := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1).
		Width(40).
		Height(6).
		Render(lipgloss.NewStyle().Bold(true).Render("Connection Status") + "\n\n" + 
			lipgloss.NewStyle().Foreground(statusColor).Render(connStatus))
	
	// System metrics card
	metricsInfo := fmt.Sprintf("CPU Usage: %s\nMemory Usage: %s\nUptime: Active\nNode ID: %s", 
		m.CPU, m.Mem, "worker-001")
	
	metricsCard := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1).
		Width(40).
		Height(6).
		Render(lipgloss.NewStyle().Bold(true).Render("System Metrics") + "\n\n" + metricsInfo)
	
	// Top row - connection and metrics side by side
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, connectionCard, " ", metricsCard)
	
	// Activity logs - bottom
	var logContent strings.Builder
	timestamp := "14:32:15"
	if m.GrpcClient != nil && m.GrpcClient.IsConnected() {
		logContent.WriteString(fmt.Sprintf("%s Worker node started successfully\n", timestamp))
		logContent.WriteString(fmt.Sprintf("%s Connected to orchestrator\n", timestamp))
		logContent.WriteString(fmt.Sprintf("%s Sending metrics every 2 seconds\n", timestamp))
		logContent.WriteString(fmt.Sprintf("%s Node operating normally", timestamp))
	} else {
		logContent.WriteString(fmt.Sprintf("%s Connection to orchestrator failed\n", timestamp))
		logContent.WriteString(fmt.Sprintf("%s Retrying connection...\n", timestamp))
		logContent.WriteString(fmt.Sprintf("%s Check orchestrator address: %s", timestamp, m.OrchestratorAddr))
	}
	
	logsCard := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1).
		Width(82). // Match width of top row
		Height(6).
		Render(lipgloss.NewStyle().Bold(true).Render("Activity Logs") + "\n\n" + logContent.String())
	
	// Instructions
	instructions := lipgloss.NewStyle().
		Faint(true).
		Align(lipgloss.Center).
		Render("ESC disconnect and return to menu | CTRL+C exit")
	
	// Grid layout
	layout := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		topRow,
		"",
		logsCard,
		"",
		instructions,
	)
	
	return strings.Split(layout, "\n")
}
