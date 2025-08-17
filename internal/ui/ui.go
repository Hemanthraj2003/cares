// Package ui provides a terminal user interface for the CARES Phase 01 MVP.
//
// overview:
//
// This package implements a small, dependency-light Bubble Tea (github.com/charmbracelet/bubbletea)
// TUI responsible for displaying sampled system metrics (CPU and memory) in a
// centered bordered box and presenting a modal confirmation dialog when the
// user requests to quit (Ctrl+C or 'q'). It is intentionally minimal so it can
// be embedded into larger applications and extended without heavy refactors.
//
// Public API:
//   - NewModel(), NewModelWithInterval(time.Duration) -> model
//     Create an initialized model with default or custom sampling intervals.
//   - Start() error
//     Bootstraps and runs the TUI program using the alternate screen buffer. It
//     returns any error from the Bubble Tea program (propagated to callers).
//
// Behavior and responsibilities:
//   - Polls metrics via the internal/metrics package by scheduling a recurring
//     tick (m.interval). Metric values are rendered inside the centered box.
//   - Respects terminal resize events and re-centers the UI accordingly.
//   - Provides a lightweight quit-confirmation flow: when the user presses
//     Ctrl+C or 'q' the UI overlays a modal asking for Y/N confirmation; 'Y'
//     quits the program, 'N' or Escape dismisses the modal.
//
// Concurrency and signals:
//   - Start() installs a signal handler (SIGINT, SIGTERM) and uses a context to
//     cancel the Bubble Tea program so shutdown is graceful. Metric sampling runs
//     on Bubble Tea's command goroutines and sends typed messages to the Update
//     method; the model itself is not concurrently mutated outside Bubble Tea's
//     single-threaded update loop.
//
// Error handling and robustness:
//   - When metric sampling fails, the UI displays "N/A" for the corresponding
//     value and continues scheduling further samples. Any unexpected errors from
//     the Bubble Tea runtime are returned from Start() so callers can act on
//     them (logging, restart, etc.). The UI avoids panics and tries to preserve
//     terminal state by using Bubble Tea's alternate screen behavior.
//
// Styling and portability:
//   - The package uses only portable Unicode box-drawing characters and avoids
//     terminal-specific libraries except Bubble Tea. When rendering the modal
//     an ANSI inverse-video sequence is used for contrast; callers should run in
//     terminals with basic ANSI support for correct appearance.
//
// Configuration and constants:
//   - desiredBoxW and desiredBoxH define the target size for the centered box.
//     The TUI shows a helpful message when the terminal is smaller than these
//     dimensions. These constants are intentionally conservative and can be
//     adjusted for different deployments.
//
// Testing and maintenance notes:
//   - View and renderBox produce plain strings and are straightforward to unit
//     test by simulating window sizes and model states. Keep renderBox logic
//     rune-aware (use go-runewidth) to handle wide characters correctly.
//   - Keep UI logic (presentation) and metrics sampling (data) separated to
//     simplify unit tests and allow headless testing of the sampling logic.
//
// Example (simplified):
//
//	if err := ui.Start(); err != nil {
//	    log.Fatalf("ui failed: %v", err)
//	}
//
// This file is intended to be production-ready for a small terminal agent and
// to serve as a clear, maintainable starting point for future Phase 02 work.
package ui

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"cares/internal/metrics"

	"github.com/mattn/go-runewidth"
)

// model is the Bubble Tea model for the minimal CARES TUI.
// Fields are exported as simple strings to make later integration straightforward.
type model struct {
	CPU      string
	Mem      string
	interval time.Duration
	// terminal window size
	winW int
	winH int
	// confirmation modal state
	showConfirm bool
}

// metricsMsg is sent by the sampler to the UI update loop.
type metricsMsg struct {
	CPU float64
	Mem float64
	Err error
}

// Desired box size for the centered UI. Chosen to fit most modern laptop terminals
// while remaining reasonable on smaller screens. If the terminal is smaller than
// this, the TUI will display a helpful message instead of the box.
const (
	desiredBoxW = 160
	desiredBoxH = 40
)

// NewModel returns an initialized model with placeholder values and a default sample interval.
func NewModel() model {
	return model{
		CPU:      "N/A",
		Mem:      "N/A",
		interval: 2 * time.Second,
		winW:     0,
		winH:     0,
	}
}

// NewModelWithInterval allows creating a model with a custom sampling interval.
func NewModelWithInterval(d time.Duration) model {
	m := NewModel()
	m.interval = d
	return m
}

// Init is called when the program starts. Kick off the first metric sampling tick.
func (m model) Init() tea.Cmd {
	return m.tickCmd()
}

// tickCmd returns a tea.Cmd that samples metrics after the configured interval
// and sends a metricsMsg to the Update loop.
func (m model) tickCmd() tea.Cmd {
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
		return metricsMsg{CPU: cpuVal, Mem: memVal, Err: err}
	})
}

// Update handles incoming messages. It processes metric updates and basic key events.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If confirmation modal is showing, handle Y/N
		if m.showConfirm {
			switch msg.String() {
			case "y", "Y":
				return m, tea.Quit
			case "n", "N", "esc":
				m.showConfirm = false
			}
			return m, nil
		}
		
		// Quit on Ctrl+C or 'q' - show confirmation modal
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			m.showConfirm = true
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.winW = msg.Width
		m.winH = msg.Height
		return m, nil
	case metricsMsg:
		if msg.Err != nil {
			// On error, display N/A and schedule next tick
			m.CPU = "N/A"
			m.Mem = "N/A"
			return m, m.tickCmd()
		}
		m.CPU = fmt.Sprintf("%.2f%%", msg.CPU)
		m.Mem = fmt.Sprintf("%.2f%%", msg.Mem)
		return m, m.tickCmd()
	}
	return m, nil
}

// renderBox builds a bordered rectangle of given width/height with the provided content lines
// centered inside. Returns the full box string without any surrounding padding.
// Uses Unicode box-drawing characters for a solid border.
func renderBox(boxW, boxH int, title, cpuLine, memLine string) string {
	if boxW < 4 || boxH < 3 {
		return ""
	}
	// Unicode box drawing
	tl := '┌'
	tr := '┐'
	bl := '└'
	br := '┘'
	horz := '─'
	vert := '│'

	hBorder := string(tl) + strings.Repeat(string(horz), boxW-2) + string(tr) + "\n"
	emptyLine := string(vert) + strings.Repeat(" ", boxW-2) + string(vert) + "\n"
	// Prepare content lines centered
	pad := func(s string) string {
		// Truncate based on display width
		if runewidth.StringWidth(s) > boxW-2 {
			s = runewidth.Truncate(s, boxW-2, "...")
		}
		w := runewidth.StringWidth(s)
		left := (boxW-2-w)/2
		if left < 0 {
			left = 0
		}
		return string(vert) + strings.Repeat(" ", left) + s + strings.Repeat(" ", boxW-2-left-w) + string(vert) + "\n"
	}

	var b strings.Builder
	b.WriteString(hBorder)
	// compute positions: place title near top, cpu and mem a few lines below
	b.WriteString(pad(title))
	b.WriteString(emptyLine)
	b.WriteString(pad(cpuLine))
	b.WriteString(pad(memLine))
	// fill remaining lines
	currentLines := 4
	for i := currentLines; i < boxH-1; i++ {
		b.WriteString(emptyLine)
	}
	b.WriteString(string(bl) + strings.Repeat(string(horz), boxW-2) + string(br) + "\n")
	return b.String()
}

// View renders the TUI. If the terminal is smaller than the desired box size, show a warning.
// Otherwise center a bordered box with the metrics inside.
func (m model) View() string {
	// If window size not yet known, just show a loading message
	if m.winW == 0 || m.winH == 0 {
		return "CARES — Phase 01\n\nDetermining terminal size...\n"
	}
	if m.winW < desiredBoxW || m.winH < desiredBoxH {
		return fmt.Sprintf("Terminal too small: need at least %dx%d (current %dx%d).\nPlease resize your terminal window or increase the terminal font/zoom (or use fullscreen) and try again.\n",
			desiredBoxW, desiredBoxH, m.winW, m.winH)
	}
	// Center the box
	leftPad := (m.winW - desiredBoxW) / 2
	topPad := (m.winH - desiredBoxH) / 2
	box := renderBox(desiredBoxW, desiredBoxH, "CARES — Phase 01", "CPU: "+m.CPU, "Memory: "+m.Mem)
	// Build the base screen first
	screenLines := make([]string, m.winH)
	for i := range screenLines {
		screenLines[i] = strings.Repeat(" ", m.winW)
	}
	
	// Place the box
	box = strings.TrimSuffix(box, "\n")
	boxLines := strings.Split(box, "\n")
	
	for i, line := range boxLines {
		if topPad+i < m.winH {
			// Add the "Quit - Ctrl+C" message on the last line (bottom border) in the bottom-left corner
			if i == len(boxLines)-1 && !m.showConfirm {
				quitMsg := " Quit - Ctrl+C "
				runes := []rune(line)
				
				// Make sure we have enough space and maintain border structure
				if len(runes) > len(quitMsg)+2 { // +2 for corner characters
					// Keep the left corner character, insert message, then continue with border
					msgRunes := []rune(quitMsg)
					// Start after the left corner character (position 1)
					copy(runes[1:1+len(msgRunes)], msgRunes)
					line = string(runes)
				}
			}
			
			// Place the line with left padding
			lineRunes := []rune(screenLines[topPad+i])
			boxRunes := []rune(line)
			if leftPad+len(boxRunes) <= len(lineRunes) {
				copy(lineRunes[leftPad:], boxRunes)
				screenLines[topPad+i] = string(lineRunes)
			}
		}
	}
	
	// Overlay confirmation modal if needed
	if m.showConfirm {
		m.overlayConfirmModal(screenLines)
	}
	
	// Build final output
	var sb strings.Builder
	for _, line := range screenLines {
		sb.WriteString(strings.TrimRight(line, " "))
		sb.WriteString("\n")
	}
	
	return sb.String()
}

// overlayConfirmModal overlays a confirmation dialog on the screen buffer with proper inverse colors
func (m model) overlayConfirmModal(screenLines []string) {
	// Modal content
	modalLines := []string{
		"┌──────────────────────────────────────┐",
		"│                                      │",
		"│        Do you really want to quit?   │",
		"│                                      │",
		"│              [Y]es / [N]o            │",
		"│                                      │",
		"└──────────────────────────────────────┘",
	}
	
	modalW := 40
	modalH := len(modalLines)
	
	// Center the modal on screen
	leftStart := (m.winW - modalW) / 2
	topStart := (m.winH - modalH) / 2
	
	// Apply inverse colors using ANSI escape codes for better visibility
	inverseOn := "\033[7m"  // Inverse video on
	inverseOff := "\033[0m" // Reset all attributes
	
	// Overlay modal onto screen buffer
	for i, modalLine := range modalLines {
		screenRow := topStart + i
		if screenRow >= 0 && screenRow < len(screenLines) {
			// Get the current line as runes
			lineRunes := []rune(screenLines[screenRow])
			
			// Create the styled modal line
			styledLine := inverseOn + modalLine + inverseOff
			
			// Make sure we have enough space
			if leftStart >= 0 && leftStart+modalW <= len(lineRunes) {
				// Replace the section with spaces first to clear it
				for j := leftStart; j < leftStart+modalW && j < len(lineRunes); j++ {
					lineRunes[j] = ' '
				}
				
				// Convert back to string and insert styled content
				screenLines[screenRow] = string(lineRunes[:leftStart]) + styledLine + string(lineRunes[leftStart+modalW:])
			}
		}
	}
}

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
