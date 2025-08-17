package ui

import (
	"context"
	"log"

	"cares/internal/cluster"

	tea "github.com/charmbracelet/bubbletea"
)

// handleSelectionKeys processes key input during mode selection screen
func (m Model) handleSelectionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.SelectedOption > 0 {
			m.SelectedOption--
		}
	case "down", "j":
		if m.SelectedOption < 1 { // 0=orchestrator, 1=worker
			m.SelectedOption++
		}
	case "enter":
		if m.SelectedOption == 0 {
			// Start orchestrator mode
			return m.startOrchestratorMode()
		} else {
			// Go to worker input mode
			m.Mode = ModeWorkerInput
			m.InputMode = true
		}
	}
	return m, nil
}

// handleInputKeys processes key input during orchestrator address entry
func (m Model) handleInputKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Connect to orchestrator and switch to worker mode
		return m.startWorkerMode()
	case "esc":
		// Go back to mode selection
		m.Mode = ModeSelection
		m.InputMode = false
	case "backspace":
		if len(m.OrchestratorAddr) > 0 {
			m.OrchestratorAddr = m.OrchestratorAddr[:len(m.OrchestratorAddr)-1]
		}
	default:
		// Add character to address input
		if len(msg.String()) == 1 && len(m.OrchestratorAddr) < 50 {
			m.OrchestratorAddr += msg.String()
		}
	}
	return m, nil
}

// handleOrchestratorKeys processes key input in orchestrator dashboard mode
func (m Model) handleOrchestratorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// For Phase 02, orchestrator mode just shows the dashboard
	// Additional key handling can be added in Phase 03
	return m, nil
}

// handleWorkerKeys processes key input in worker mode (same as Phase 01)
func (m Model) handleWorkerKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Worker mode behaves like Phase 01 - no additional keys needed
	return m, nil
}

// startOrchestratorMode initializes the gRPC server and switches to orchestrator mode
func (m Model) startOrchestratorMode() (tea.Model, tea.Cmd) {
	// Create gRPC server
	m.GrpcServer = cluster.NewServer()
	m.NodeRegistry = m.GrpcServer.GetRegistry()
	m.Mode = ModeOrchestrator
	
	// Start gRPC server in background goroutine
	go func() {
		if err := m.GrpcServer.StartServer("50051"); err != nil {
			// TODO: In Phase 03, send error message to TUI
			log.Printf("gRPC server error: %v", err)
		}
	}()
	
	return m, nil
}

// startWorkerMode initializes the gRPC client and switches to worker mode
func (m Model) startWorkerMode() (tea.Model, tea.Cmd) {
	// Create gRPC client
	m.GrpcClient = cluster.NewClient("worker-node")
	
	// Connect to orchestrator
	if err := m.GrpcClient.Connect(m.OrchestratorAddr); err != nil {
		// TODO: In Phase 03, show error to user
		log.Printf("Failed to connect to orchestrator: %v", err)
		// Go back to input mode
		m.Mode = ModeWorkerInput
		return m, nil
	}
	
	// Switch to worker mode and start metrics collection
	m.Mode = ModeWorker
	
	// Start heartbeat in background
	go func() {
		ctx := context.Background()
		if err := m.GrpcClient.StartHeartbeat(ctx); err != nil {
			log.Printf("Heartbeat error: %v", err)
		}
	}()
	
	// Start local metrics collection (same as Phase 01)
	return m, m.tickCmd()
}
