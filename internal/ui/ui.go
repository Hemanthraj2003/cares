// Package ui provides a minimal Bubble Tea TUI used by the Phase 01 MVP.
//
// This package offers a lightweight scaffold: a Model with placeholders for
// CPU and Memory values, and a Start() helper that boots the Bubble Tea program.
// The UI is intentionally minimal so it can be integrated quickly and extended
// later to display live metrics.
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
	desiredBoxW = 140
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
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
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
	// Build final view with vertical and horizontal padding
	var sb strings.Builder
	for i := 0; i < topPad; i++ {
		sb.WriteString("\n")
	}
	padSpaces := strings.Repeat(" ", leftPad)
	box = strings.TrimSuffix(box, "\n")
	for _, line := range strings.Split(box, "\n") {
		// render even empty lines so borders are not truncated
		sb.WriteString(padSpaces)
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	return sb.String()
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
