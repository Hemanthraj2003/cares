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
	
	// Simple centered layout - no instructions needed (they're at bottom)
	layout := lipgloss.JoinVertical(
		lipgloss.Center,
		header,
		"",
		"",
		opt1,
		opt2,
	)
	
	return strings.Split(layout, "\n")
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
	if m.OrchestratorAddr == "" && !m.InputMode {
		// Get actual local IP for placeholder
		localIP := getLocalIP()
		placeholder := fmt.Sprintf("%s:50051", localIP)
		inputContent = lipgloss.NewStyle().
			Faint(true).
			Render(placeholder)
	}
	
	inputBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(0, 1).
		Width(30).
		Align(lipgloss.Center).
		Render(inputContent)
	
	// Simple centered layout - help text is at bottom border
	layout := lipgloss.JoinVertical(
		lipgloss.Center,
		header,
		"",
		"",
		label,
		"",
		inputBox,
	)
	
	return strings.Split(layout, "\n")
}
