package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// getModeSelectionContent returns enhanced role selection screen with side-by-side options
func (m Model) getModeSelectionContent() []string {
	// Main title with full form - no colors, just bold
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Align(lipgloss.Center).
		MarginBottom(1)
	
	title := titleStyle.Render("CARES - Cost-Aware Resource Allocation and Execution Scheduler")
	
	// Description about CARES and technology - plain text
	descStyle := lipgloss.NewStyle().
		Width(90).
		Align(lipgloss.Center).
		MarginBottom(2)
	
	description := descStyle.Render(
		"A distributed computing platform for efficient containerized function execution.\n" +
		"Built with Go, gRPC for communication, Docker for containerization, and Bubble Tea for the TUI.\n" +
		"Enables cost-aware resource allocation across multiple nodes in a cluster.")
	
	// Role selection header - no colors, just bold
	roleHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Align(lipgloss.Center).
		MarginBottom(2)
	
	roleHeader := roleHeaderStyle.Render("SELECT YOUR ROLE:")
	
	// Option styles
	optionWidth := 40
	optionHeight := 12
	
	baseOptionStyle := lipgloss.NewStyle().
		Width(optionWidth).
		Height(optionHeight).
		Border(lipgloss.NormalBorder()).
		Padding(1, 2).
		Align(lipgloss.Left)
	
	selectedOptionStyle := baseOptionStyle.
		Reverse(true)
	
	// Option 1 - Orchestrator content
	option1Content := "1. CLUSTER ORCHESTRATOR\n\n" +
		"Start as central coordinator\n\n" +
		"What happens when selected:\n" +
		"• Initializes gRPC server on :50051\n" +
		"• Creates cluster node registry\n" +
		"• Launches web dashboard\n" +
		"• Accepts worker connections\n" +
		"• Manages function distribution\n" +
		"• Provides cluster monitoring"
	
	// Option 2 - Worker content
	option2Content := "2. WORKER NODE\n\n" +
		"Join existing cluster\n\n" +
		"What happens when selected:\n" +
		"• Prompts for orchestrator address\n" +
		"• Establishes gRPC connection\n" +
		"• Registers node capabilities\n" +
		"• Starts local execution server\n" +
		"• Reports system metrics\n" +
		"• Waits for task assignments"
	
	// Apply selection styling to entire box
	var orchestratorBox, workerBox string
	if m.SelectedOption == 0 {
		orchestratorBox = selectedOptionStyle.Render(option1Content)
		workerBox = baseOptionStyle.Render(option2Content)
	} else {
		orchestratorBox = baseOptionStyle.Render(option1Content)
		workerBox = selectedOptionStyle.Render(option2Content)
	}
	
	// Arrange options side by side
	optionsRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		orchestratorBox,
		strings.Repeat(" ", 6), // Spacing between options
		workerBox,
	)
	
	// Navigation instructions - plain text
	navStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		MarginTop(2)
	
	navigation := navStyle.Render("Use [←] [→] arrow keys to navigate • Press Enter to select your role")
	
	// System status footer - plain text
	footerStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		MarginTop(1)
	
	footer := footerStyle.Render("Current System Resources: CPU " + m.CPU + " | Memory " + m.Mem)
	
	// Create the complete layout
	content := lipgloss.NewStyle().
		Width(m.WinW - 8).
		Height(m.WinH - 10).
		Align(lipgloss.Center, lipgloss.Center).
		Render(lipgloss.JoinVertical(
			lipgloss.Center,
			title,
			"",
			description,
			"",
			roleHeader,
			optionsRow,
			navigation,
			footer,
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
