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
	
	// Create menu items - Orchestrator Details as default first
	menuItems := []string{
		"Orchestrator", 
		"Logs",
		"Functions",
		"Add Function",
	}
	
	var menuContent []string
	for i, item := range menuItems {
		if i == m.SidebarSelected {
			// Selected item - reverse colors - center all titles
			line := lipgloss.NewStyle().
				Reverse(true).
				Bold(true).
				Width(menuWidth-2).
				Align(lipgloss.Center).
				Render(item)
			menuContent = append(menuContent, line)
		} else {
			// Normal item - center all titles
			centeredItem := lipgloss.NewStyle().
				Width(menuWidth-2).
				Align(lipgloss.Center).
				Render(item)
			menuContent = append(menuContent, centeredItem)
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
	case 0: // Orchestrator - now first/default
		contentText = m.getOrchestratorContent()
	case 1: // Logs
		contentText = m.getLogsContent()
	case 2: // Functions
		contentText = m.getFunctionsContent(contentWidth)
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
	
	// Enhanced styling to match orchestrator UI
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Reverse(true).  // Inverted for headings like orchestrator
		Padding(0, 1)
	
	descriptionStyle := lipgloss.NewStyle().
		Faint(true).      // Grey/dull color for descriptions
		Italic(true)
	
	labelStyle := lipgloss.NewStyle().Bold(true)
	
	// Header with inverse highlighting
	lines = append(lines, 
		headerStyle.Render("  WORKER NODE  "),
		"",
		descriptionStyle.Render("DISTRIBUTED COMPUTE WORKER - EXECUTING CONTAINERIZED FUNCTIONS"),
		"",
	)
	
	// Connection status section
	lines = append(lines,
		headerStyle.Render("  CONNECTION STATUS  "),
		"",
	)
	
	if m.GrpcClient != nil && m.GrpcClient.IsConnected() {
		lines = append(lines, 
			fmt.Sprintf("%s %s", labelStyle.Render("ORCHESTRATOR:"), m.OrchestratorAddr),
			fmt.Sprintf("%s %s", labelStyle.Render("STATUS:"), "ONLINE"),
			fmt.Sprintf("%s %s", labelStyle.Render("HEARTBEAT:"), "ACTIVE"),
			"",
			descriptionStyle.Render("Connected to cluster orchestrator - ready for task assignments"),
		)
	} else {
		lines = append(lines,
			fmt.Sprintf("%s %s", labelStyle.Render("ORCHESTRATOR:"), m.OrchestratorAddr),
			fmt.Sprintf("%s %s", labelStyle.Render("STATUS:"), "DISCONNECTED"),
			fmt.Sprintf("%s %s", labelStyle.Render("HEARTBEAT:"), "INACTIVE"),
			"",
			descriptionStyle.Render("Worker node isolated - attempting reconnection to cluster"),
		)
	}
	
	lines = append(lines, "")
	
	// System metrics section
	lines = append(lines,
		headerStyle.Render("  SYSTEM METRICS  "),
		"",
		fmt.Sprintf("%s %s", labelStyle.Render("CPU USAGE:"), m.CPU),
		fmt.Sprintf("%s %s", labelStyle.Render("MEMORY USAGE:"), m.Mem),
		"",
		fmt.Sprintf("%s %s", labelStyle.Render("NODE ID:"), "WORKER-001"),
		fmt.Sprintf("%s %s", labelStyle.Render("UPTIME:"), "ACTIVE"),
		"",
		descriptionStyle.Render("Real-time resource monitoring for optimal task distribution"),
		"",
		"",
		descriptionStyle.Render("■ DOCKER CONTAINERIZATION FOR SECURE EXECUTION"),
		descriptionStyle.Render("■ AUTOMATIC LOAD BALANCING AND FAILOVER"),
		descriptionStyle.Render("■ COST-AWARE RESOURCE ALLOCATION"),
	)
	
	return lines
}

// getLogsContent returns logs content for the right panel
func (m Model) getLogsContent() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Underline(true).
		MarginBottom(2)
	
	timestampStyle := lipgloss.NewStyle().
		Faint(true)
	
	logStyle := lipgloss.NewStyle()
	
	watermarkStyle := lipgloss.NewStyle().
		Faint(true).
		Italic(true)
	
	var lines []string
	
	lines = append(lines,
		titleStyle.Render("SYSTEM ACTIVITY LOGS"),
		"",
		timestampStyle.Render("[14:32:07]") + " ORCHESTRATOR INITIALIZED SUCCESSFULLY",
		timestampStyle.Render("[14:32:08]") + " GRPC SERVER LISTENING ON PORT :50051",
		timestampStyle.Render("[14:32:09]") + " REST API SERVER RUNNING ON PORT :8080",
		timestampStyle.Render("[14:32:10]") + " FUNCTION REGISTRY INITIALIZED",
		"",
	)
	
	if m.NodeRegistry != nil {
		nodes := m.NodeRegistry.GetAllNodes()
		if len(nodes) > 0 {
			lines = append(lines, 
				timestampStyle.Render("[14:32:11]") + 
				logStyle.Render(fmt.Sprintf(" %d WORKER NODE(S) CONNECTED TO CLUSTER", len(nodes))))
			
			for i, node := range nodes {
				if i >= 3 { // Limit to 3 nodes for readability
					break
				}
				
				nodeID := node.ID
				if len(nodeID) > 12 {
					nodeID = nodeID[:9] + "..."
				}
				
				lines = append(lines, 
					timestampStyle.Render(fmt.Sprintf("[14:32:1%d]", 2+i)) + 
					logStyle.Render(fmt.Sprintf(" %s: CPU %.1f%% | MEM %.1f%% | STATUS: ACTIVE", 
						nodeID, node.CPUUsage, node.MemoryUsage)))
			}
			
			lines = append(lines, 
				timestampStyle.Render("[14:32:15]") + " CLUSTER LOAD BALANCING ACTIVE")
		} else {
			lines = append(lines,
				timestampStyle.Render("[14:32:11]") + " WAITING FOR WORKER NODES TO JOIN...")
		}
	}
	
	lines = append(lines, 
		"",
		timestampStyle.Render("[14:32:16]") + " SYSTEM OPERATIONAL - MONITORING ACTIVE",
		"",
		"",
		watermarkStyle.Render("REAL-TIME LOGS | AUTO-ROTATION | EVENTS HIGHLIGHTED"),
	)
	
	return strings.Join(lines, "\n")
}

// getOrchestratorContent returns orchestrator info for the right panel
func (m Model) getOrchestratorContent() string {
	if m.NodeRegistry == nil {
		return "Registry not initialized"
	}
	
	nodes := m.NodeRegistry.GetAllNodes()
	localIP := getLocalIP()
	
	// Enhanced styling with inverted colors for highlights
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Reverse(true).  // Inverted for headings
		Padding(0, 1)
	
	// Inverted style for IP addresses and ports
	highlightStyle := lipgloss.NewStyle().
		Reverse(true).
		Bold(true).
		Padding(0, 1)
	
	watermarkStyle := lipgloss.NewStyle().
		Faint(true).
		Italic(true)

	// Custom grey color for tooltips - more subdued
	tooltipStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).  // Dark grey color
		Italic(true)
	
	var lines []string
	
	// Centered title on same line
	centerStyle := lipgloss.NewStyle().
		Bold(true).
		Reverse(true).
		Align(lipgloss.Center).
		Width(60)
	
	lines = append(lines,
		centerStyle.Render("ORCHESTRATOR DASHBOARD - DISTRIBUTED COMPUTE CONTROL"),
		"",
	)
	
	// System information with inverted highlights
	lines = append(lines,
		headerStyle.Render("  SYSTEM INFORMATION  "),
		"",
	)
	
	// Compact layout to prevent line wrapping
	lines = append(lines,
		fmt.Sprintf("ORCHESTRATOR ID: ORCH-%s", localIP[strings.LastIndex(localIP, ".")+1:]),
		fmt.Sprintf("NETWORK ADDRESS: %s", highlightStyle.Render(fmt.Sprintf("%s:50051", localIP))),
		tooltipStyle.Render("  → gRPC communication endpoint for worker nodes"),
		fmt.Sprintf("HTTP SERVER: %s", highlightStyle.Render(fmt.Sprintf("%s:8080", localIP))),
		tooltipStyle.Render("  → REST API server for function invocation"),
		fmt.Sprintf("STATUS: %s", highlightStyle.Render("ONLINE")),
	)
	
	lines = append(lines, "")
	
	// Cluster metrics with inverted header
	lines = append(lines,
		headerStyle.Render("  CLUSTER METRICS  "),
		"",
	)
	
	metricsLeft := []string{
		fmt.Sprintf("%-15s %d", "ACTIVE NODES:", len(nodes)),
		fmt.Sprintf("%-15s %d CORES", "TOTAL CAPACITY:", len(nodes)*4),
	}
	
	metricsRight := []string{
		fmt.Sprintf("%-15s ACTIVE", "LOAD BALANCER:"),
		fmt.Sprintf("%-15s ENABLED", "FAILOVER MODE:"),
	}
	
	// Add tooltips for cluster metrics
	metricsTooltips := []string{
		"→ Connected worker nodes ready for task execution",
		"→ Automatic failover ensures high availability",
	}
	
	for i := 0; i < len(metricsLeft) || i < len(metricsRight); i++ {
		var left, right string
		if i < len(metricsLeft) {
			left = metricsLeft[i]
		}
		if i < len(metricsRight) {
			right = metricsRight[i]
		}
		
		leftPart := lipgloss.NewStyle().Width(35).Render(left)
		line := leftPart + " " + right
		lines = append(lines, line)
		
		// Add tooltip for this metric line
		if i < len(metricsTooltips) {
			lines = append(lines, tooltipStyle.Render("  "+metricsTooltips[i]))
		}
	}
	
	lines = append(lines, "")
	
	// Worker nodes table (removed the header since we removed it per request)
	if len(nodes) == 0 {
		lines = append(lines, 
			"NO WORKER NODES CONNECTED",
			tooltipStyle.Render("  → Start worker nodes to begin distributed processing"),
			"",
			fmt.Sprintf("JOIN ADDRESS: %s", highlightStyle.Render(fmt.Sprintf("%s:50051", localIP))),
			tooltipStyle.Render("  → Use this address when starting new worker nodes"),
		)
	} else {
		// Compact table that fits (removed header per request)
		lines = append(lines, 
			"┌───────────┬─────────┬─────────┬─────────┐",
			"│ NODE ID   │ CPU     │ MEMORY  │ STATUS  │",
			"├───────────┼─────────┼─────────┼─────────┤",
		)
		
		// Limit nodes to prevent overflow
		maxNodes := 3
		if len(nodes) > maxNodes {
			nodes = nodes[:maxNodes]
		}
		
		for _, node := range nodes {
			nodeID := node.ID
			if len(nodeID) > 9 {
				nodeID = nodeID[:6] + "..."
			}
			
			status := "OFFLINE"
			if string(node.Status) == "Active" {
				status = "ONLINE"
			}
			
			lines = append(lines, fmt.Sprintf("│ %-9s │ %5.1f%% │ %5.1f%% │ %-7s │",
				nodeID, node.CPUUsage, node.MemoryUsage, status))
		}
		
		lines = append(lines, 
			"└───────────┴─────────┴─────────┴─────────┘",
		)
		
		if len(m.NodeRegistry.GetAllNodes()) > maxNodes {
			lines = append(lines, 
				watermarkStyle.Render(fmt.Sprintf("... and %d more nodes", len(m.NodeRegistry.GetAllNodes())-maxNodes)))
		}
		
		lines = append(lines, 
			"",
			"CLUSTER STATUS: OPERATIONAL",
			tooltipStyle.Render("  → All systems running, ready for function execution"))
	}
	
	return strings.Join(lines, "\n")
}

// getFunctionsContent returns functions list for the right panel
func (m Model) getFunctionsContent(contentWidth int) string {
	if m.FunctionRegistry == nil {
		return "Function registry not initialized"
	}
	
	functions := m.FunctionRegistry.GetAllFunctions()
	
	// Styling
	titleStyle := lipgloss.NewStyle().Bold(true).Underline(true)
	labelStyle := lipgloss.NewStyle().Bold(true)
	highlightStyle := lipgloss.NewStyle().Reverse(true).Bold(true).Padding(0, 1)
	tooltipStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	selectedRowStyle := lipgloss.NewStyle().Reverse(true)
	
	var lines []string
	
	if len(functions) == 0 {
		return strings.Join([]string{
			titleStyle.Render("FUNCTION REGISTRY"),
			"",
			"NO FUNCTIONS REGISTERED YET",
			"",
			tooltipStyle.Render("→ Use 'Add Function' to register new functions"),
			tooltipStyle.Render("→ Functions will appear here once registered"),
		}, "\n")
	}
	
	// Calculate selected function index from model state
	selectedIndex := m.FunctionSelectedIndex
	if selectedIndex >= len(functions) {
		selectedIndex = 0
	}
	
	selectedFunction := functions[selectedIndex]
	
	// 30% area - Selected function details
	lines = append(lines,
		titleStyle.Render("FUNCTION REGISTRY"),
		"",
		fmt.Sprintf("%s %s", labelStyle.Render("SELECTED:"), selectedFunction.Name),
		fmt.Sprintf("%s %s", labelStyle.Render("IMAGE:"), selectedFunction.Image),
		fmt.Sprintf("%s %s", labelStyle.Render("STATUS:"), highlightStyle.Render(strings.ToUpper(selectedFunction.Status))),
		fmt.Sprintf("%s %s", labelStyle.Render("ENDPOINT:"), fmt.Sprintf("POST /invoke/%s", strings.ToLower(selectedFunction.Name))),
		tooltipStyle.Render(fmt.Sprintf("→ Description: %s", getOrDefault(selectedFunction.Description, "No description provided"))),
		"",
		"",
	)
	
	// 70% area - Navigable table
	lines = append(lines,
		labelStyle.Render("FUNCTION INVENTORY - PRESS ENTER TO NAVIGATE"),
		"",
	)
	
	// Calculate table width to use full available space
	tableWidth := contentWidth - 4 // Account for border padding
	if tableWidth < 40 {
		tableWidth = 40 // Minimum width
	}
	
	// Calculate column widths dynamically (4 columns: Function, Image, Status, Endpoint)
	// Function: 20%, Docker Image: 30%, Status: 15%, Endpoint: 35%
	functionWidth := tableWidth * 20 / 100
	imageWidth := tableWidth * 30 / 100  
	statusWidth := tableWidth * 15 / 100
	endpointWidth := tableWidth - functionWidth - imageWidth - statusWidth - 6 // Account for separators
	
	// Ensure minimum widths
	if functionWidth < 8 {
		functionWidth = 8
	}
	if imageWidth < 12 {
		imageWidth = 12
	}
	if statusWidth < 6 {
		statusWidth = 6
	}
	if endpointWidth < 15 {
		endpointWidth = 15
	}
	
	// Build dynamic table header
	topBorder := "┌" + strings.Repeat("─", functionWidth) + "┬" + strings.Repeat("─", imageWidth) + "┬" + strings.Repeat("─", statusWidth) + "┬" + strings.Repeat("─", endpointWidth) + "┐"
	headerRow := fmt.Sprintf("│ %-*s │ %-*s │ %-*s │ %-*s │", functionWidth-2, "FUNCTION", imageWidth-2, "IMAGE", statusWidth-2, "STATUS", endpointWidth-2, "ENDPOINT")
	midBorder := "├" + strings.Repeat("─", functionWidth) + "┼" + strings.Repeat("─", imageWidth) + "┼" + strings.Repeat("─", statusWidth) + "┼" + strings.Repeat("─", endpointWidth) + "┤"
	
	lines = append(lines, topBorder, headerRow, midBorder)
	
	// Fixed number of table rows (7 rows)
	maxRows := 7
	for i := 0; i < maxRows; i++ {
		var row string
		
		if i < len(functions) {
			// Display actual function data
			fn := functions[i]
			
			// Truncate text to fit column widths
			name := fn.Name
			if len(name) > functionWidth-3 {
				name = name[:functionWidth-6] + "..."
			}
			
			image := fn.Image
			if len(image) > imageWidth-3 {
				image = image[:imageWidth-6] + "..."
			}
			
			status := "READY"
			if fn.Status == "active" {
				status = "ACTIVE"
			}
			
			// Generate endpoint
			endpoint := fmt.Sprintf("/invoke/%s", strings.ToLower(fn.Name))
			if len(endpoint) > endpointWidth-3 {
				endpoint = endpoint[:endpointWidth-6] + "..."
			}
			
			row = fmt.Sprintf("│ %-*s │ %-*s │ %-*s │ %-*s │",
				functionWidth-2, name, imageWidth-2, image, statusWidth-2, status, endpointWidth-2, endpoint)
			
			// Highlight selected row only when table is focused
			if i == selectedIndex && m.FunctionTableFocused {
				row = selectedRowStyle.Render(row)
			}
		} else {
			// Empty row with dynamic spacing
			row = fmt.Sprintf("│%*s│%*s│%*s│%*s│", 
				functionWidth, "", imageWidth, "", statusWidth, "", endpointWidth, "")
		}
		
		lines = append(lines, row)
	}
	
	// Table footer with dynamic width
	bottomBorder := "└" + strings.Repeat("─", functionWidth) + "┴" + strings.Repeat("─", imageWidth) + "┴" + strings.Repeat("─", statusWidth) + "┴" + strings.Repeat("─", endpointWidth) + "┘"
	lines = append(lines, bottomBorder,
		"",
		tooltipStyle.Render(func() string {
			if m.FunctionTableFocused {
				return fmt.Sprintf("→ Function %d of %d | ↑↓: Navigate | ESC: Exit table", selectedIndex+1, len(functions))
			}
			return fmt.Sprintf("→ %d of %d functions | ENTER: Navigate table", len(functions), maxRows)
		}()),
	)
	
	return strings.Join(lines, "\n")
}

// Helper function to get value or default
func getOrDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// getAddFunctionContent returns the add function form for the right panel
func (m Model) getAddFunctionContent() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Underline(true).
		MarginBottom(2)
	
	labelStyle := lipgloss.NewStyle().Bold(true)
	activeFieldStyle := lipgloss.NewStyle().Bold(true).Reverse(true)
	
	inputActiveStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Bold(true)
	
	inputInactiveStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(0, 1)
	
	watermarkStyle := lipgloss.NewStyle().
		Faint(true).
		Italic(true)
	
	var lines []string
	
	lines = append(lines,
		titleStyle.Render("ADD NEW FUNCTION"),
		"",
		watermarkStyle.Render("REGISTER A NEW SERVERLESS FUNCTION FOR EXECUTION"),
		"",
		"",
	)
	
	// Form fields with enhanced styling
	fields := []struct {
		label       string
		value       string
		active      bool
		placeholder string
		required    bool
	}{
		{"FUNCTION NAME", m.FunctionFormName, m.FunctionFormField == 0, "E.G., HELLO-WORLD", true},
		{"DOCKER IMAGE", m.FunctionFormImage, m.FunctionFormField == 1, "E.G., NODE:16-ALPINE", true},
		{"DESCRIPTION", m.FunctionFormDesc, m.FunctionFormField == 2, "BRIEF DESCRIPTION (OPTIONAL)", false},
	}
	
	for _, field := range fields {
		// Field label with indicator
		var labelText string
		if field.active {
			labelText = activeFieldStyle.Render(fmt.Sprintf(" %s ", field.label))
		} else {
			labelText = labelStyle.Render(field.label)
		}
		
		if field.required {
			labelText += " *"
		}
		
		lines = append(lines, labelText)
		
		// Field value with cursor
		value := field.value
		if field.active {
			value += "|"
		}
		
		// Show placeholder if empty and not active
		if value == "" && !field.active {
			value = field.placeholder
		}
		
		// Apply appropriate styling
		var styledValue string
		if field.active {
			styledValue = inputActiveStyle.Width(50).Render(value)
		} else {
			if field.value == "" {
				styledValue = inputInactiveStyle.Width(50).Render(watermarkStyle.Render(value))
			} else {
				styledValue = inputInactiveStyle.Width(50).Render(value)
			}
		}
		
		lines = append(lines, styledValue, "")
	}
	
	// Validation status
	if m.FunctionFormName != "" && m.FunctionFormImage != "" {
		lines = append(lines, 
			labelStyle.Render("STATUS: READY TO SUBMIT"),
			"")
	} else {
		lines = append(lines, 
			watermarkStyle.Render("STATUS: NAME AND IMAGE ARE REQUIRED"),
			"")
	}
	
	// Add navigation instructions using tooltip style
	tooltipStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).  // Dark grey color
		Italic(true)
	
	lines = append(lines,
		"",
		tooltipStyle.Render("→ TAB/UP/DOWN: Navigate fields"),
		tooltipStyle.Render("→ ENTER: Submit function"),
		tooltipStyle.Render("→ ESC: Cancel and return"),
	)
	
	return strings.Join(lines, "\n")
}


