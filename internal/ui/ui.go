package ui

// Package ui provides a modular terminal user interface for the CARES Phase 02 MVP.
//
// This package is now modularized for better maintainability:
//   - types.go: Core data structures and constants
//   - model.go: Main Bubble Tea model and initialization
//   - handlers.go: Key handling and business logic
//   - views.go: Content generation and rendering utilities
//   - ui.go: Program orchestration and Start() function
//
// The modular structure makes the codebase more maintainable and allows
// for easier testing and extension in future phases.

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
)

// Start launches the Bubble Tea program and blocks until it exits.
// It uses the terminal's alternate screen buffer so the TUI occupies the full
// terminal window while running, and restores the terminal on exit.
//
// Start also installs a signal handler for SIGINT/SIGTERM; when such a signal
// is received the context is cancelled which causes the Bubble Tea program to
// exit cleanly.
func Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Forward OS signals to the context cancel function so the TUI exits cleanly.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	p := tea.NewProgram(NewModel(), tea.WithAltScreen(), tea.WithContext(ctx))
	// Program.Run is the preferred, non-deprecated entrypoint.
	_, err := p.Run()
	return err
}
