package ui

import (
	"fmt"
	"strings"

	"cares/internal/registry"

	"github.com/mattn/go-runewidth"
)

// View renders the TUI based on current mode.
func (m Model) View() string {
	// If window size not yet known, just show a loading message
	if m.WinW == 0 || m.WinH == 0 {
		return "CARES — Phase 02\n\nDetermining terminal size...\n"
	}
	if m.WinW < DesiredBoxW || m.WinH < DesiredBoxH {
		return fmt.Sprintf("Terminal too small: need at least %dx%d (current %dx%d).\nPlease resize your terminal window or increase the terminal font/zoom (or use fullscreen) and try again.\n\nPress Ctrl+C to quit.\n",
			DesiredBoxW, DesiredBoxH, m.WinW, m.WinH)
	}

	// All modes now render inside the Phase 01 rectangle UI
	return m.renderInBox()
}

// renderInBox renders all modes inside the Phase 01 rectangle UI
func (m Model) renderInBox() string {
	// Center the box (reusing Phase 01 positioning logic)
	leftPad := (m.WinW - DesiredBoxW) / 2
	topPad := (m.WinH - DesiredBoxH) / 2
	
	// Get content based on current mode
	var content []string
	switch m.Mode {
	case ModeSelection:
		content = m.getModeSelectionContent()
	case ModeWorkerInput:
		content = m.getWorkerInputContent()
	case ModeOrchestrator:
		content = m.getOrchestratorContent()
	case ModeWorker:
		content = m.getWorkerContent()
	default:
		content = []string{"Unknown mode"}
	}
	
	// Build the box with content
	box := m.renderBoxWithContent(DesiredBoxW, DesiredBoxH, content)
	
	// Build the base screen first
	screenLines := make([]string, m.WinH)
	for i := range screenLines {
		screenLines[i] = strings.Repeat(" ", m.WinW)
	}
	
	// Place the box
	box = strings.TrimSuffix(box, "\n")
	boxLines := strings.Split(box, "\n")
	
	for i, line := range boxLines {
		if topPad+i < m.WinH {
			// Add the "Quit - Ctrl+C" message on the last line (bottom border) in the bottom-left corner
			if i == len(boxLines)-1 && !m.ShowConfirm {
				quitMsg := " Quit - Ctrl+C "
				runes := []rune(line)
				
				// Make sure we have enough space and maintain border structure
				if len(runes) > len(quitMsg)+2 { // +2 for corner characters
					// Keep the left corner character, insert message, then continue with border
					msgRunes := []rune(quitMsg)
					// Start after the left corner character (position 1)
					copy(runes[1:1+len(msgRunes)], msgRunes)
					line = string(runes)
				}
			}
			
			// Place the line with left padding
			lineRunes := []rune(screenLines[topPad+i])
			boxRunes := []rune(line)
			if leftPad+len(boxRunes) <= len(lineRunes) {
				copy(lineRunes[leftPad:], boxRunes)
				screenLines[topPad+i] = string(lineRunes)
			}
		}
	}
	
	// Overlay confirmation modal if needed
	if m.ShowConfirm {
		m.overlayConfirmModal(screenLines)
	}
	
	// Build final output
	var sb strings.Builder
	for _, line := range screenLines {
		sb.WriteString(strings.TrimRight(line, " "))
		sb.WriteString("\n")
	}
	
	return sb.String()
}

// getModeSelectionContent returns content for mode selection screen
func (m Model) getModeSelectionContent() []string {
	content := []string{
		"CARES — Phase 02 Cluster Setup",
		"",
		"",
		"",
		"",
		"",
	}
	
	// Create properly aligned options with inversion for selected
	option1 := "[ 1 ] Start Orchestrator"
	option2 := "[ 2 ] Join as Worker Node"
	
	if m.SelectedOption == 0 {
		// Highlight selected option with inverse colors
		option1 = "\033[7m " + option1 + " \033[0m"
	} else {
		// Add padding to match highlighted option width
		option1 = " " + option1 + " "
	}
	
	if m.SelectedOption == 1 {
		// Highlight selected option with inverse colors
		option2 = "\033[7m " + option2 + " \033[0m"
	} else {
		// Add padding to match highlighted option width
		option2 = " " + option2 + " "
	}
	
	content = append(content, option1)
	content = append(content, "")
	content = append(content, option2)
	content = append(content, "")
	content = append(content, "")
	content = append(content, "")
	content = append(content, "")
	content = append(content, "Use ↑↓ to select, Enter to confirm")
	
	return content
}

// getWorkerInputContent returns content for worker input screen
func (m Model) getWorkerInputContent() []string {
	content := []string{
		"CARES — Join Existing Cluster",
		"",
		"Enter orchestrator address (host:port):",
		"",
	}
	
	// Simple cursor indicator
	cursor := ""
	if m.InputMode {
		cursor = "█"
	}
	
	inputLine := m.OrchestratorAddr + cursor
	content = append(content, inputLine)
	content = append(content, "")
	content = append(content, "")
	content = append(content, "Press Enter to connect, Esc to go back")
	
	return content
}

// getOrchestratorContent returns content for orchestrator dashboard
func (m Model) getOrchestratorContent() []string {
	content := []string{
		"CARES — Orchestrator Dashboard",
		"Listening on: localhost:50051",
		"",
	}
	
	if m.NodeRegistry == nil {
		content = append(content, "Registry not initialized")
		return content
	}
	
	nodes := m.NodeRegistry.GetAllNodes()
	if len(nodes) == 0 {
		content = append(content, "No worker nodes connected.")
		content = append(content, "Workers can join using: localhost:50051")
		content = append(content, "")
		content = append(content, "┌─── System Logs ───────────────────────────┐")
		content = append(content, "│ Orchestrator started successfully         │")
		content = append(content, "│ Waiting for worker connections...         │")
		content = append(content, "│                                           │")
		content = append(content, "│                                           │")
		content = append(content, "│                                           │")
		content = append(content, "└───────────────────────────────────────────┘")
	} else {
		content = append(content, fmt.Sprintf("Connected Nodes: %d", len(nodes)))
		content = append(content, "")
		
		// Show up to 3 nodes to save space for logs
		maxNodes := 3
		for i, node := range nodes {
			if i >= maxNodes {
				break
			}
			
			status := "🟢"
			if node.Status != registry.NodeStatusActive {
				status = "🔴"
			}
			
			content = append(content, fmt.Sprintf("%s %s", status, node.ID[:8])) // Short ID
			content = append(content, fmt.Sprintf("  CPU: %.1f%% | Memory: %.1f%%", 
				node.CPUUsage, node.MemoryUsage))
			content = append(content, "")
		}
		
		if len(nodes) > maxNodes {
			content = append(content, fmt.Sprintf("... and %d more nodes", len(nodes)-maxNodes))
			content = append(content, "")
		}
		
		// Add logs section
		content = append(content, "┌─── System Logs ───────────────────────────┐")
		content = append(content, "│ Orchestrator started successfully         │")
		content = append(content, fmt.Sprintf("│ %d worker node(s) connected               │", len(nodes)))
		content = append(content, "│ Heartbeat streams active                  │")
		content = append(content, "│ Cluster operating normally                │")
		content = append(content, "└───────────────────────────────────────────┘")
	}
	
	return content
}

// getWorkerContent returns content for worker mode (Phase 01 style)
func (m Model) getWorkerContent() []string {
	// Connection status
	connectionStatus := "Disconnected"
	if m.GrpcClient != nil && m.GrpcClient.IsConnected() {
		connectionStatus = "Connected to " + m.OrchestratorAddr
	}
	
	content := []string{
		"CARES — Worker Node",
		connectionStatus,
		"",
		"System Metrics:",
		"  CPU: " + m.CPU,
		"  Memory: " + m.Mem,
		"",
		"┌─── System Logs ───────────────────────────┐",
	}
	
	if m.GrpcClient != nil && m.GrpcClient.IsConnected() {
		content = append(content, "│ Connected to orchestrator successfully    │")
		content = append(content, "│ Heartbeat stream active                   │")
		content = append(content, "│ Sending metrics every 2 seconds          │")
		content = append(content, "│ Node operating normally                   │")
	} else {
		content = append(content, "│ Connection failed                         │")
		content = append(content, "│ Retrying connection...                    │")
		content = append(content, "│ Check orchestrator address               │")
		content = append(content, "│                                           │")
	}
	
	content = append(content, "└───────────────────────────────────────────┘")
	
	return content
}

// renderBoxWithContent builds a bordered rectangle with the provided content lines
func (m Model) renderBoxWithContent(boxW, boxH int, contentLines []string) string {
	if boxW < 4 || boxH < 3 {
		return ""
	}
	// Unicode box drawing
	tl := '┌'
	tr := '┐'
	bl := '└'
	br := '┘'
	horz := '─'
	vert := '│'

	hBorder := string(tl) + strings.Repeat(string(horz), boxW-2) + string(tr) + "\n"
	emptyLine := string(vert) + strings.Repeat(" ", boxW-2) + string(vert) + "\n"
	
	// Prepare content lines centered
	pad := func(s string) string {
		// Handle ANSI escape sequences for color inversion
		displayWidth := runewidth.StringWidth(s)
		// If string contains ANSI codes, calculate actual display width
		if strings.Contains(s, "\033[") {
			// Remove ANSI codes for width calculation
			cleaned := strings.ReplaceAll(s, "\033[7m", "")
			cleaned = strings.ReplaceAll(cleaned, "\033[0m", "")
			displayWidth = runewidth.StringWidth(cleaned)
		}
		
		// Truncate based on display width
		if displayWidth > boxW-2 {
			s = runewidth.Truncate(s, boxW-2, "...")
			displayWidth = boxW-2
		}
		
		left := (boxW-2-displayWidth)/2
		if left < 0 {
			left = 0
		}
		right := boxW-2-left-displayWidth
		if right < 0 {
			right = 0
		}
		return string(vert) + strings.Repeat(" ", left) + s + strings.Repeat(" ", right) + string(vert) + "\n"
	}

	var b strings.Builder
	b.WriteString(hBorder)
	
	// Add content lines
	lineCount := 1 // Start after top border
	for _, line := range contentLines {
		if lineCount < boxH-1 { // Leave space for bottom border
			b.WriteString(pad(line))
			lineCount++
		}
	}
	
	// Fill remaining lines
	for lineCount < boxH-1 {
		b.WriteString(emptyLine)
		lineCount++
	}
	
	b.WriteString(string(bl) + strings.Repeat(string(horz), boxW-2) + string(br) + "\n")
	return b.String()
}

// overlayConfirmModal overlays a confirmation dialog on the screen buffer with proper inverse colors
func (m Model) overlayConfirmModal(screenLines []string) {
	// Modal content
	modalLines := []string{
		"┌──────────────────────────────────────┐",
		"│                                      │",
		"│        Do you really want to quit?   │",
		"│                                      │",
		"│              [Y]es / [N]o            │",
		"│                                      │",
		"└──────────────────────────────────────┘",
	}
	
	modalW := 40
	modalH := len(modalLines)
	
	// Center the modal on screen
	leftStart := (m.WinW - modalW) / 2
	topStart := (m.WinH - modalH) / 2
	
	// Apply inverse colors using ANSI escape codes for better visibility
	inverseOn := "\033[7m"  // Inverse video on
	inverseOff := "\033[0m" // Reset all attributes
	
	// Overlay modal onto screen buffer
	for i, modalLine := range modalLines {
		screenRow := topStart + i
		if screenRow >= 0 && screenRow < len(screenLines) {
			// Get the current line as runes
			lineRunes := []rune(screenLines[screenRow])
			
			// Create the styled modal line
			styledLine := inverseOn + modalLine + inverseOff
			
			// Make sure we have enough space
			if leftStart >= 0 && leftStart+modalW <= len(lineRunes) {
				// Replace the section with spaces first to clear it
				for j := leftStart; j < leftStart+modalW && j < len(lineRunes); j++ {
					lineRunes[j] = ' '
				}
				
				// Convert back to string and insert styled content
				screenLines[screenRow] = string(lineRunes[:leftStart]) + styledLine + string(lineRunes[leftStart+modalW:])
			}
		}
	}
}
