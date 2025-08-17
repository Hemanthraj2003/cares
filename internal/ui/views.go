package ui

import (
	"fmt"
)

// View renders the TUI based on current mode using pure Lipgloss
func (m Model) View() string {
	// If window size not yet known, show loading
	if m.WinW == 0 || m.WinH == 0 {
		return "CARES — Phase 02\n\nDetermining terminal size...\n"
	}
	
	// Check minimum size
	if m.WinW < 80 || m.WinH < 24 {
		return fmt.Sprintf("Terminal too small: need at least 80x24 (current %dx%d).\nPlease resize your terminal window.\n\nPress Ctrl+C to quit.\n",
			m.WinW, m.WinH)
	}

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
	
	// Get mode-specific help text
	var helpText string
	switch m.Mode {
	case ModeSelection:
		helpText = "Use ↑↓ arrow keys | Enter to select | Ctrl+C to quit"
	case ModeWorkerInput:
		helpText = "Type address | Enter to connect | Esc to go back | Ctrl+C to quit"
	case ModeOrchestrator:
		helpText = "↑↓ scroll nodes | Esc return to menu | Ctrl+C to quit"
	case ModeWorker:
		helpText = "Esc disconnect | Ctrl+C to quit"
	default:
		helpText = "Ctrl+C to quit"
	}
	
	// Render using pure Lipgloss container
	if m.ShowConfirm {
		return m.overlayConfirmModal(m.renderMainContainerWithHelp(content, helpText))
	}
	
	return m.renderMainContainerWithHelp(content, helpText)
}
