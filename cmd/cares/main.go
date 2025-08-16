package main

import (
	"cares/internal/ui"
	"fmt"
	"os"
)

func main() {
	// Start the minimal TUI (blocks until exit)
	if err := ui.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "TUI exited with error:", err)
		os.Exit(1)
	}
}