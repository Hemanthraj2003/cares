package ui

import (
	"fmt"
	"net"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// getLocalIP returns the local IP address of this machine
func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "localhost"
	}
	defer conn.Close()
	
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// renderMainContainerWithHelp renders content with border and custom help text at bottom
func (m *Model) renderMainContainerWithHelp(content []string, helpText string) string {
	// Join all content lines
	contentStr := strings.Join(content, "\n")
	
	// CARES title bar at top
	titleBar := lipgloss.NewStyle().
		Bold(true).
		Reverse(true).
		Width(m.WinW).
		Align(lipgloss.Center).
		Render(" CARES ")
	
	// Main content area with border - left align for complex layouts like sidebar
	mainContent := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(2).
		Width(m.WinW - 4).
		Height(m.WinH - 6). // Account for title and help text
		Align(lipgloss.Left).
		Render(contentStr)
	
	// Help text at bottom - mode-specific
	helpBar := lipgloss.NewStyle().
		Faint(true).
		Width(m.WinW).
		Align(lipgloss.Center).
		Render(helpText)
	
	// Combine everything vertically
	layout := lipgloss.JoinVertical(
		lipgloss.Left,
		titleBar,
		mainContent,
		helpBar,
	)
	
	return layout
}

// overlayConfirmModal overlays a confirmation dialog OVER the existing content
func (m *Model) overlayConfirmModal(screenContent string) string {
	// Create a simple modal box
	modalContent := "Do you really want to quit?\n\n[Y]es / [N]o"
	
	modalWidth := min(m.WinW/3, 35)
	if modalWidth < 25 {
		modalWidth = 25
	}
	
	modal := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		Padding(1, 2).
		Width(modalWidth).
		Bold(true).
		Align(lipgloss.Center).
		Render(modalContent)
	
	// Split base content into lines
	lines := strings.Split(screenContent, "\n")
	
	// Calculate center position for modal
	centerY := len(lines) / 2 - 2
	centerX := (m.WinW - modalWidth) / 2
	
	// Insert modal lines over the base content
	modalLines := strings.Split(modal, "\n")
	for i, modalLine := range modalLines {
		lineIdx := centerY + i
		if lineIdx >= 0 && lineIdx < len(lines) {
			// Create padding to center the modal horizontally
			padding := strings.Repeat(" ", centerX)
			if centerX > 0 {
				lines[lineIdx] = padding + modalLine
			} else {
				lines[lineIdx] = modalLine
			}
		}
	}
	
	return strings.Join(lines, "\n")
}

// overlayFunctionConfirmModal overlays a function registration confirmation dialog
func (m *Model) overlayFunctionConfirmModal(screenContent string) string {
	// Get local IP for API endpoint display
	localIP := getLocalIP()
	
	// Create modal content with formatted API endpoint
	modalWidth := min(m.WinW/2, 60)
	if modalWidth < 40 {
		modalWidth = 40
	}
	
	// Create title with inverse style
	title := lipgloss.NewStyle().
		Bold(true).
		Reverse(true).
		Padding(0, 1).
		Align(lipgloss.Center).
		Render("CONFIRM FUNCTION REGISTRATION")
	
	// Create function details
	details := fmt.Sprintf(`Name: %s
Image: %s
Description: %s

API Endpoint:
POST http://%s:8080/invoke/%s

This function will be available at the above endpoint.`, 
		m.FunctionConfirmName,
		m.FunctionConfirmImage,
		m.FunctionConfirmDesc,
		localIP,
		m.FunctionConfirmName)
	
	// Create side-by-side buttons
	buttons := lipgloss.JoinHorizontal(
		lipgloss.Center,
		lipgloss.NewStyle().Padding(0, 3).Render("[Y]es"),
		lipgloss.NewStyle().Padding(0, 3).Render("[N]o"),
	)
	
	// Combine all elements
	modalContent := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		details,
		"",
		buttons,
	)
	
	modal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1).
		Width(modalWidth).
		Align(lipgloss.Center).
		Render(modalContent)
	
	// Split base content into lines
	lines := strings.Split(screenContent, "\n")
	
	// Calculate center position for modal
	centerY := len(lines)/2 - strings.Count(modalContent, "\n")/2 - 2
	if centerY < 0 {
		centerY = 0
	}
	centerX := (m.WinW - modalWidth) / 2
	if centerX < 0 {
		centerX = 0
	}
	
	// Insert modal lines over the base content
	modalLines := strings.Split(modal, "\n")
	for i, modalLine := range modalLines {
		lineIdx := centerY + i
		if lineIdx >= 0 && lineIdx < len(lines) {
			// Create padding to center the modal horizontally
			padding := strings.Repeat(" ", centerX)
			if centerX > 0 {
				lines[lineIdx] = padding + modalLine
			} else {
				lines[lineIdx] = modalLine
			}
		}
	}
	
	return strings.Join(lines, "\n")
}
