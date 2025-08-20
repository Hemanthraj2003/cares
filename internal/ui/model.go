package ui

import (
	"fmt"
	"time"

	"cares/internal/functions"
	"cares/internal/logging"
	"cares/internal/metrics"

	tea "github.com/charmbracelet/bubbletea"
)

// NewModel returns an initialized model starting in mode selection.
func NewModel() *Model {
	return &Model{
		// Phase 01 defaults
		CPU:      "N/A",
		Mem:      "N/A",
		interval: 2 * time.Second,
		WinW:     0,
		WinH:     0,
		
		// Phase 02 defaults - start in mode selection
		Mode:             ModeSelection,
		SelectedOption:   0,
		OrchestratorAddr: "", // Will be filled with local IP when needed
		InputMode:        false,
		
		// Phase 03 defaults
		SidebarSelected:  0,
		SidebarView:      "cluster",
		ShowFunctionForm: false,
		FunctionFormField: 0,
	}
}

// NewModelWithInterval allows creating a model with a custom sampling interval.
func NewModelWithInterval(d time.Duration) *Model {
	m := NewModel()
	m.interval = d
	return m
}

// Init is called when the program starts. Kick off the first metric sampling tick.
func (m *Model) Init() tea.Cmd {
	return m.tickCmd()
}

// tickCmd returns a tea.Cmd that samples metrics after the configured interval
// and sends a MetricsMsg to the Update loop.
func (m *Model) tickCmd() tea.Cmd {
	interval := m.interval
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		cpuVal, err1 := metrics.GetCPUUsage()
		memVal, err2 := metrics.GetMemoryUsage()
		var err error
		if err1 != nil {
			err = err1
		} else if err2 != nil {
			err = err2
		}
		return MetricsMsg{CPU: cpuVal, Mem: memVal, Err: err}
	})
}

// Update handles incoming messages. It processes different behavior based on current mode.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global quit confirmation (works in all modes)
		if m.ShowConfirm {
			switch msg.String() {
			case "y", "Y":
				// Save function registry before quitting if it exists
				if m.FunctionRegistry != nil {
					// Save synchronously to ensure it completes before exiting
					if err := m.FunctionRegistry.SaveToFile(functions.DefaultStoragePath); err != nil {
						// Log error but still quit
						logging.Warn("Failed to save function registry: %v", err)
					} else {
						logging.Info("Function registry saved to %s", functions.DefaultStoragePath)
					}
				}
				return m, tea.Quit
			case "n", "N", "esc":
				m.ShowConfirm = false
			}
			return m, nil
		}
		
		// Global quit trigger (works in all modes) - including when screen is too small
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			// If screen is too small, quit directly without confirmation
			if m.WinW < DesiredBoxW || m.WinH < DesiredBoxH {
				return m, tea.Quit
			}
			m.ShowConfirm = true
			return m, nil
		}
		
		// Mode-specific key handling
		switch m.Mode {
		case ModeSelection:
			return m.handleSelectionKeys(msg)
		case ModeWorkerInput:
			return m.handleInputKeys(msg)
		case ModeOrchestrator:
			return m.handleOrchestratorKeys(msg)
		case ModeOrchestratorSidebar:
			return m.handleOrchestratorSidebarKeys(msg)
		case ModeWorker:
			return m.handleWorkerKeys(msg)
		}
	case tea.WindowSizeMsg:
		m.WinW = msg.Width
		m.WinH = msg.Height
		return m, nil
	case MetricsMsg:
		if msg.Err != nil {
			// On error, display N/A and schedule next tick
			m.CPU = "N/A"
			m.Mem = "N/A"
		} else {
			m.CPU = fmt.Sprintf("%.2f%%", msg.CPU)
			m.Mem = fmt.Sprintf("%.2f%%", msg.Mem)
			
			// Update graph history for worker mode
			if m.Mode == ModeWorker {
				// Add new data points
				m.CPUHistory = append(m.CPUHistory, msg.CPU)
				m.MemoryHistory = append(m.MemoryHistory, msg.Mem)
				
				// Keep only last 20 data points to prevent memory growth
				maxHistory := 20
				if len(m.CPUHistory) > maxHistory {
					m.CPUHistory = m.CPUHistory[len(m.CPUHistory)-maxHistory:]
				}
				if len(m.MemoryHistory) > maxHistory {
					m.MemoryHistory = m.MemoryHistory[len(m.MemoryHistory)-maxHistory:]
				}
			}
		}
		
		// Continue ticking for all modes (orchestrator mode needs regular updates to show node changes)
		return m, m.tickCmd()
	}
	return m, nil
}
