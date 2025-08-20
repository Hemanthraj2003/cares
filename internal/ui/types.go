package ui

import (
	"time"

	"cares/internal/api"
	"cares/internal/cluster"
	"cares/internal/functions"
	"cares/internal/registry"
)

// AppMode represents the current mode of the application
type AppMode int

const (
	// ModeSelection - Initial screen where user chooses orchestrator or worker
	ModeSelection AppMode = iota
	// ModeOrchestrator - Running as cluster orchestrator (shows node dashboard)
	ModeOrchestrator
	// ModeWorker - Running as worker node (shows local metrics like Phase 01)
	ModeWorker
	// ModeWorkerInput - Getting orchestrator address input from user
	ModeWorkerInput
	// ModeOrchestratorSidebar - New sidebar navigation mode for orchestrator
	ModeOrchestratorSidebar
)

// Desired box size for the centered UI. Chosen to fit most modern laptop terminals
// while remaining reasonable on smaller screens. If the terminal is smaller than
// this, the TUI will display a helpful message instead of the box.
const (
	DesiredBoxW = 160
	DesiredBoxH = 40
)

// Model is the Bubble Tea model for the CARES Phase 02 TUI.
// Now supports multiple modes: mode selection, orchestrator dashboard, and worker view.
type Model struct {
	// Phase 01 fields (worker mode)
	CPU      string
	Mem      string
	interval time.Duration
	
	// Terminal window size
	WinW int
	WinH int
	
	// Application state
	Mode        AppMode
	ShowConfirm bool
	
	// Mode selection
	SelectedOption int // 0 = orchestrator, 1 = worker
	
	// Worker mode - orchestrator address input
	OrchestratorAddr string
	InputMode        bool
	
	// Orchestrator mode - cluster state
	GrpcServer      *cluster.Server
	NodeRegistry    *registry.NodeRegistry
	NodeScrollOffset int // For scrolling through nodes list
	
	// Worker mode - connection to orchestrator
	GrpcClient *cluster.Client
	WorkerGrpcServer *cluster.Server // Worker's own gRPC server for receiving function execution requests
	
	// Phase 3 - Function management
	FunctionRegistry *functions.Registry
	ApiServer        *api.Server
	
	// Sidebar navigation state
	SidebarSelected  int
	SidebarView      string // "cluster", "functions", "logs"
	
	// Function form state
	ShowFunctionForm bool
	FunctionFormName string
	FunctionFormImage string
	FunctionFormDesc string
	FunctionFormField int // 0=name, 1=image, 2=desc
	
	// Function navigation state
	FunctionTableFocused bool // True when user is navigating functions table
	FunctionSelectedIndex int // Currently selected function in table
	
	// Node navigation state
	NodeTableFocused bool // True when user is navigating nodes table
	NodeSelectedIndex int // Currently selected node in table
	
	// Function confirmation modal state
	ShowFunctionConfirmModal bool
	FunctionConfirmName string
	FunctionConfirmImage string
	FunctionConfirmDesc string
}

// MetricsMsg is sent by the sampler to the UI update loop.
type MetricsMsg struct {
	CPU float64
	Mem float64
	Err error
}
