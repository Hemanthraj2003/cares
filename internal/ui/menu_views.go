package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// getModeSelectionContent returns simple mode selection - clean terminal style
func (m Model) getModeSelectionContent() []string {
	// Simple centered header
	header := lipgloss.NewStyle().
		Bold(true).
		Align(lipgloss.Center).
		Render("Phase 02 Cluster Setup")
	
	// Simple option styles - no fancy borders
	baseStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Align(lipgloss.Center)
	
	selectedStyle := baseStyle.
		Reverse(true)
	
	// Simple options - just the essentials
	option1 := "[ 1 ] Start Orchestrator"
	option2 := "[ 2 ] Join Worker Pool"
	
	var opt1, opt2 string
	if m.SelectedOption == 0 {
		opt1 = selectedStyle.Render(option1)
		opt2 = baseStyle.Render(option2)
	} else {
		opt1 = baseStyle.Render(option1)
		opt2 = selectedStyle.Render(option2)
	}
	
	// Create a centered layout that will work within the container
	content := lipgloss.NewStyle().
		Width(m.WinW - 8). // Account for container padding and borders
		Height(m.WinH - 10). // Account for title, help, borders, padding
		Align(lipgloss.Center, lipgloss.Center). // Both horizontal and vertical center
		Render(lipgloss.JoinVertical(
			lipgloss.Center,
			header,
			"",
			"",
			opt1,
			opt2,
		))
	
	return strings.Split(content, "\n")
}

// getWorkerInputContent returns simple worker input screen
func (m Model) getWorkerInputContent() []string {
	// Simple centered header
	header := lipgloss.NewStyle().
		Bold(true).
		Align(lipgloss.Center).
		Render("Join Existing Cluster")
	
	// Simple label
	label := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Render("Enter orchestrator address (host:port):")
	
	// Input with cursor
	cursor := ""
	if m.InputMode {
		cursor = "|"
	}
	
	inputContent := m.OrchestratorAddr + cursor
	if m.OrchestratorAddr == "localhost:50051" || (m.OrchestratorAddr == "" && !m.InputMode) {
		// Get actual local IP for placeholder instead of localhost
		localIP := getLocalIP()
		placeholder := fmt.Sprintf("%s:50051", localIP)
		if m.OrchestratorAddr == "" && !m.InputMode {
			inputContent = lipgloss.NewStyle().
				Faint(true).
				Render(placeholder)
		} else {
			// Replace localhost with actual IP for display
			inputContent = strings.Replace(inputContent, "localhost", localIP, 1)
		}
	}
	
	inputWidth := m.WinW / 3
	if inputWidth < 25 {
		inputWidth = 25
	} else if inputWidth > 50 {
		inputWidth = 50
	}
	
	inputBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(0, 1).
		Width(inputWidth).
		Align(lipgloss.Center).
		Render(inputContent)
	
	// Create a centered layout that will work within the container
	content := lipgloss.NewStyle().
		Width(m.WinW - 8). // Account for container padding and borders
		Height(m.WinH - 10). // Account for title, help, borders, padding
		Align(lipgloss.Center, lipgloss.Center). // Both horizontal and vertical center
		Render(lipgloss.JoinVertical(
			lipgloss.Center,
			header,
			"",
			"",
			label,
			"",
			inputBox,
		))
	
	return strings.Split(content, "\n")
}
