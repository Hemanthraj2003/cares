package ui

import (
	"context"
	"log"

	"cares/internal/api"
	"cares/internal/cluster"
	"cares/internal/functions"

	tea "github.com/charmbracelet/bubbletea"
)

// handleSelectionKeys processes key input during mode selection screen
func (m *Model) handleSelectionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
func (m *Model) handleInputKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
func (m *Model) handleOrchestratorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.NodeRegistry == nil {
		return m, nil
	}
	
	nodes := m.NodeRegistry.GetAllNodes()
	maxVisibleNodes := 4 // Same as in views.go
	
	switch msg.String() {
	case "up", "k":
		// Scroll up in nodes list
		if m.NodeScrollOffset > 0 {
			m.NodeScrollOffset--
		}
	case "down", "j":
		// Scroll down in nodes list
		if len(nodes) > maxVisibleNodes && m.NodeScrollOffset < len(nodes)-maxVisibleNodes {
			m.NodeScrollOffset++
		}
	case "esc":
		// Return to mode selection menu
		// Cleanup orchestrator mode
		if m.GrpcServer != nil {
			// TODO: Properly stop the gRPC server in Phase 03
		}
		m.Mode = ModeSelection
		m.GrpcServer = nil
		m.NodeRegistry = nil
		m.NodeScrollOffset = 0
	}
	
	return m, nil
}

// handleWorkerKeys processes key input in worker mode (same as Phase 01)
func (m *Model) handleWorkerKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Disconnect from orchestrator and return to menu
		if m.GrpcClient != nil {
			m.GrpcClient.Disconnect()
		}
		m.Mode = ModeSelection
		m.GrpcClient = nil
		m.OrchestratorAddr = ""
		m.InputMode = false
	}
	return m, nil
}

// startOrchestratorMode initializes the gRPC server and switches to orchestrator mode
func (m *Model) startOrchestratorMode() (tea.Model, tea.Cmd) {
	// Create gRPC server
	m.GrpcServer = cluster.NewServer()
	m.NodeRegistry = m.GrpcServer.GetRegistry()
	
	// Create function registry and API server
	m.FunctionRegistry = functions.NewRegistry()
	m.ApiServer = api.NewServer(m.FunctionRegistry)
	
	// Connect API server to node registry for function execution
	m.ApiServer.SetNodeRegistry(m.NodeRegistry)
	
	// Switch to sidebar mode for Phase 3
	m.Mode = ModeOrchestratorSidebar
	m.SidebarSelected = 0  // Start with "Logs" selected
	
	// Start gRPC server in background goroutine
	go func() {
		if err := m.GrpcServer.StartServer("50051"); err != nil {
			// TODO: In Phase 03, send error message to TUI
			log.Printf("gRPC server error: %v", err)
		}
	}()
	
	// Start REST API server in background goroutine
	go func() {
		if err := m.ApiServer.StartServer("8080"); err != nil {
			log.Printf("REST API server error: %v", err)
		}
	}()
	
	// Start the tick command to refresh UI regularly (this will show node updates)
	return m, m.tickCmd()
}

// startWorkerMode initializes the gRPC client and switches to worker mode
func (m *Model) startWorkerMode() (tea.Model, tea.Cmd) {
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
	
	// Create and start worker's own gRPC server for receiving function execution requests
	m.WorkerGrpcServer = cluster.NewServer()
	go func() {
		if err := m.WorkerGrpcServer.StartServer("50052"); err != nil {
			log.Printf("Worker gRPC server error: %v", err)
		}
	}()
	
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

// handleOrchestratorSidebarKeys processes key input in orchestrator sidebar mode
func (m *Model) handleOrchestratorSidebarKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle function confirmation modal if open
	if m.ShowFunctionConfirmModal {
		return m.handleFunctionConfirmModalKeys(msg)
	}
	
	// Handle function form input if form is open
	if m.ShowFunctionForm {
		return m.handleFunctionFormKeys(msg)
	}
	
	switch msg.String() {
	case "up", "k":
		if m.SidebarSelected > 0 {
			m.SidebarSelected--
		}
	case "down", "j":
		maxItems := 4 // logs, orchestrator, functions, add-function
		if m.SidebarSelected < maxItems-1 {
			m.SidebarSelected++
		}
	case "enter", " ":
		switch m.SidebarSelected {
		case 0: // Logs
			// Just selection change, content will update automatically
		case 1: // Orchestrator
			// Just selection change, content will update automatically
		case 2: // Functions
			// Just selection change, content will update automatically
		case 3: // Add Function
			// Open function form
			m.ShowFunctionForm = true
			m.FunctionFormName = ""
			m.FunctionFormImage = ""
			m.FunctionFormDesc = ""
			m.FunctionFormField = 0
		}
	case "esc":
		// Return to mode selection menu
		// Cleanup orchestrator mode
		if m.GrpcServer != nil {
			// TODO: Properly stop the servers in Phase 03+
		}
		m.Mode = ModeSelection
		m.GrpcServer = nil
		m.NodeRegistry = nil
		m.ApiServer = nil
		m.FunctionRegistry = nil
		m.NodeScrollOffset = 0
		m.SidebarSelected = 0
		m.ShowFunctionForm = false
	}
	
	return m, nil
}

// handleFunctionFormKeys processes key input in function registration form
func (m *Model) handleFunctionFormKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Close form
		m.ShowFunctionForm = false
		m.FunctionFormName = ""
		m.FunctionFormImage = ""
		m.FunctionFormDesc = ""
		m.FunctionFormField = 0
	case "tab", "down":
		// Move to next field
		if m.FunctionFormField < 2 {
			m.FunctionFormField++
		}
	case "shift+tab", "up":
		// Move to previous field
		if m.FunctionFormField > 0 {
			m.FunctionFormField--
		}
	case "enter":
		// Show confirmation modal if all required fields are filled
		if m.FunctionFormName != "" && m.FunctionFormImage != "" {
			return m.validateAndShowConfirmModal()
		}
	case "backspace":
		// Delete character from current field
		switch m.FunctionFormField {
		case 0:
			if len(m.FunctionFormName) > 0 {
				m.FunctionFormName = m.FunctionFormName[:len(m.FunctionFormName)-1]
			}
		case 1:
			if len(m.FunctionFormImage) > 0 {
				m.FunctionFormImage = m.FunctionFormImage[:len(m.FunctionFormImage)-1]
			}
		case 2:
			if len(m.FunctionFormDesc) > 0 {
				m.FunctionFormDesc = m.FunctionFormDesc[:len(m.FunctionFormDesc)-1]
			}
		}
	default:
		// Add character to current field
		if len(msg.String()) == 1 {
			switch m.FunctionFormField {
			case 0:
				if len(m.FunctionFormName) < 50 {
					m.FunctionFormName += msg.String()
				}
			case 1:
				if len(m.FunctionFormImage) < 100 {
					m.FunctionFormImage += msg.String()
				}
			case 2:
				if len(m.FunctionFormDesc) < 200 {
					m.FunctionFormDesc += msg.String()
				}
			}
		}
	}
	
	return m, nil
}

// validateAndShowConfirmModal validates the function form before showing confirmation
func (m *Model) validateAndShowConfirmModal() (tea.Model, tea.Cmd) {
	// First set the confirm fields so they can be displayed in the modal
	m.FunctionConfirmName = m.FunctionFormName
	m.FunctionConfirmImage = m.FunctionFormImage
	m.FunctionConfirmDesc = m.FunctionFormDesc
	
	// Show the confirmation modal
	m.ShowFunctionConfirmModal = true
	
	return m, nil
}

// handleFunctionConfirmModalKeys processes key input in function confirmation modal
func (m *Model) handleFunctionConfirmModalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "n", "N", "esc":
		// No - just close the modal
		m.ShowFunctionConfirmModal = false
		
	case "y", "Y":
		// Yes - close modal and add function
		m.ShowFunctionConfirmModal = false
		
		// Add function directly to registry
		if m.FunctionRegistry != nil {
			_, err := m.FunctionRegistry.AddFunction(m.FunctionConfirmName, m.FunctionFormImage, m.FunctionFormDesc)
			if err != nil {
				// TODO: Show error message in UI
				log.Printf("Failed to add function: %v", err)
			} else {
				// Success - close form and reset fields
				m.ShowFunctionForm = false
				m.FunctionFormName = ""
				m.FunctionFormImage = ""
				m.FunctionFormDesc = ""
				m.FunctionFormField = 0
				m.FunctionConfirmName = ""
				m.FunctionConfirmImage = ""
				m.FunctionConfirmDesc = ""
				m.SidebarSelected = 2 // Switch to Functions view
			}
		}
	}
	
	return m, nil
}
