package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// getTwoPanelLayout returns two equal empty rectangles
func (m Model) getTwoPanelLayout() []string {
	// Safety check for minimum terminal size
	if m.WinW < 60 || m.WinH < 20 {
		return []string{
			"Terminal too small",
			fmt.Sprintf("Current: %dx%d", m.WinW, m.WinH),
			"Need at least 60x20",
		}
	}
	
	// Calculate 1:4 ratio panel widths
	totalWidth := m.WinW - 15  // Account for main container padding
	menuWidth := totalWidth / 5      // 1 part for menu (20%)
	contentWidth := totalWidth * 4 / 5  // 4 parts for content (80%)
	
	// Calculate dynamic height based on terminal size
	// Subtract space for title bar (3 lines) + help text (3 lines) + padding (6 lines) = 12 lines
	availableHeight := m.WinH - 10
	if availableHeight < 5 {
		availableHeight = 5 // minimum height
	}
	
	// Create menu items
	menuItems := []string{
		"Logs",
		"Orchestrator", 
		"Functions",
		"Add Function",
	}
	
	var menuContent []string
	for i, item := range menuItems {
		if i == m.SidebarSelected {
			// Selected item - reverse colors
			line := lipgloss.NewStyle().
				Reverse(true).
				Bold(true).
				Width(menuWidth-2).
				Render(" " + item)
			menuContent = append(menuContent, line)
		} else {
			// Normal item
			menuContent = append(menuContent, " " + item)
		}
		// Add line spacing after each menu item
		menuContent = append(menuContent, "")
	}
	
	// Fill remaining height for menu
	for len(menuContent) < availableHeight {
		menuContent = append(menuContent, "")
	}
	
	// Left panel - menu (NO border)
	leftPanel := lipgloss.NewStyle().
		Width(menuWidth).
		Render(strings.Join(menuContent, "\n"))
	
	// Right panel - content area based on selected menu item
	var contentText string
	switch m.SidebarSelected {
	case 0: // Logs
		contentText = m.getLogsContent()
	case 1: // Orchestrator
		contentText = m.getOrchestratorContent()
	case 2: // Functions
		contentText = m.getFunctionsContent()
	case 3: // Add Function
		contentText = m.getAddFunctionContent()
	default:
		contentText = "Select a menu item to view content"
	}
	
	// Calculate content height to prevent overflow
	contentHeight := availableHeight - 4  // Account for border and padding
	if contentHeight < 3 {
		contentHeight = 3
	}
	
	rightPanel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(contentWidth).
		Height(contentHeight).
		Padding(1).
		Render(contentText)
	
	// Join the two panels horizontally
	layout := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanel,
		rightPanel,
	)
	
	return strings.Split(layout, "\n")
}













// getSimpleWorkerContent returns simple worker node info (like Phase 01)
func (m Model) getSimpleWorkerContent() []string {
	var lines []string
	
	// Header
	lines = append(lines, 
		lipgloss.NewStyle().Bold(true).Render("WORKER NODE"),
		"",
	)
	
	// Connection status
	if m.GrpcClient != nil && m.GrpcClient.IsConnected() {
		lines = append(lines, 
			fmt.Sprintf("Connected to: %s", m.OrchestratorAddr),
			"Status: ONLINE",
			"Heartbeat: Active",
		)
	} else {
		lines = append(lines,
			fmt.Sprintf("Orchestrator: %s", m.OrchestratorAddr),
			"Status: DISCONNECTED",
			"Heartbeat: Inactive",
		)
	}
	
	lines = append(lines, "")
	
	// System metrics
	lines = append(lines,
		lipgloss.NewStyle().Bold(true).Render("SYSTEM METRICS"),
		"",
		fmt.Sprintf("CPU Usage: %s", m.CPU),
		fmt.Sprintf("Memory Usage: %s", m.Mem),
		"",
		"Node ID: worker-001",
		"Uptime: Active",
	)
	
	return lines
}

// getLogsContent returns logs content for the right panel
func (m Model) getLogsContent() string {
	var lines []string
	
	lines = append(lines,
		lipgloss.NewStyle().Bold(true).Render("SYSTEM LOGS"),
		"",
		"[14:32:07] Orchestrator started",
		"[14:32:08] gRPC server on :50051",
		"[14:32:09] REST API on :8080",
		"",
	)
	
	if m.NodeRegistry != nil {
		nodes := m.NodeRegistry.GetAllNodes()
		if len(nodes) > 0 {
			lines = append(lines, fmt.Sprintf("[14:32:10] %d nodes connected", len(nodes)))
			for i, node := range nodes {
				if i >= 3 { // Limit to 3 nodes
					break
				}
				lines = append(lines, fmt.Sprintf("[14:32:1%d] %s: CPU %.1f%% MEM %.1f%%", 
					1+i, node.ID[:8], node.CPUUsage, node.MemoryUsage))
			}
		}
	}
	
	lines = append(lines, "", "System operational - monitoring active")
	return strings.Join(lines, "\n")
}

// getOrchestratorContent returns orchestrator info for the right panel
func (m Model) getOrchestratorContent() string {
	if m.NodeRegistry == nil {
		return "Registry not initialized"
	}
	
	nodes := m.NodeRegistry.GetAllNodes()
	localIP := getLocalIP()
	
	var lines []string
	lines = append(lines,
		lipgloss.NewStyle().Bold(true).Render("ORCHESTRATOR STATUS"),
		"",
		fmt.Sprintf("Address: %s:50051", localIP),
		"API Port: 8080",
		"Status: ONLINE",
		fmt.Sprintf("Connected Nodes: %d", len(nodes)),
		"",
		lipgloss.NewStyle().Bold(true).Render("WORKER NODES"),
	)
	
	if len(nodes) == 0 {
		lines = append(lines, "", "No worker nodes connected")
		lines = append(lines, "", "Workers can join using:")
		lines = append(lines, localIP+":50051")
	} else {
		lines = append(lines, "")
		lines = append(lines, "ID       │ CPU   │ MEM   │ STATUS")
		lines = append(lines, "─────────┼───────┼───────┼────────")
		
		for _, node := range nodes {
			nodeID := node.ID
			if len(nodeID) > 8 {
				nodeID = nodeID[:8]
			}
			
			status := "OFFLINE"
			if string(node.Status) == "Active" {
				status = "ONLINE"
			}
			
			lines = append(lines, fmt.Sprintf("%-8s │ %5.1f │ %5.1f │ %s",
				nodeID, node.CPUUsage, node.MemoryUsage, status))
		}
	}
	
	return strings.Join(lines, "\n")
}

// getFunctionsContent returns functions list for the right panel
func (m Model) getFunctionsContent() string {
	if m.FunctionRegistry == nil {
		return "Function registry not initialized"
	}
	
	functions := m.FunctionRegistry.GetAllFunctions()
	
	var lines []string
	lines = append(lines,
		lipgloss.NewStyle().Bold(true).Render("REGISTERED FUNCTIONS"),
		"",
		fmt.Sprintf("Total: %d functions", len(functions)),
		"",
	)
	
	if len(functions) == 0 {
		lines = append(lines, "No functions registered yet.")
		lines = append(lines, "")
		lines = append(lines, "Use 'Add Function' from menu to register.")
	} else {
		lines = append(lines, "NAME         │ IMAGE            │ STATUS")
		lines = append(lines, "─────────────┼──────────────────┼────────")
		
		for _, fn := range functions {
			name := fn.Name
			if len(name) > 12 {
				name = name[:9] + "..."
			}
			
			image := fn.Image
			if len(image) > 16 {
				image = image[:13] + "..."
			}
			
			status := "INACTIVE"
			if fn.Status == "active" {
				status = "ACTIVE"
			}
			
			lines = append(lines, fmt.Sprintf("%-12s │ %-16s │ %s", name, image, status))
		}
		
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().Bold(true).Render("API ENDPOINTS"))
		for _, fn := range functions {
			lines = append(lines, fmt.Sprintf("POST /invoke/%s", fn.Name))
		}
	}
	
	return strings.Join(lines, "\n")
}

// getAddFunctionContent returns the add function form for the right panel
func (m Model) getAddFunctionContent() string {
	var lines []string
	
	lines = append(lines,
		lipgloss.NewStyle().Bold(true).Render("ADD NEW FUNCTION"),
		"",
	)
	
	// Form fields
	fields := []struct {
		label string
		value string
		active bool
	}{
		{"Name", m.FunctionFormName, m.FunctionFormField == 0},
		{"Image", m.FunctionFormImage, m.FunctionFormField == 1},
		{"Description", m.FunctionFormDesc, m.FunctionFormField == 2},
	}
	
	for _, field := range fields {
		if field.active {
			lines = append(lines, lipgloss.NewStyle().Bold(true).Render("> "+field.label+":"))
		} else {
			lines = append(lines, "  "+field.label+":")
		}
		
		value := field.value
		if field.active {
			value += "│"
		}
		
		if field.active {
			lines = append(lines, lipgloss.NewStyle().Reverse(true).Render("  "+value))
		} else {
			lines = append(lines, "  "+value)
		}
		lines = append(lines, "")
	}
	
	lines = append(lines,
		"Tab: Next field",
		"Enter: Submit",
		"Esc: Cancel")
	
	return strings.Join(lines, "\n")
}


