// Package main contains the CARES Phase 01 CLI entrypoint.
//
// This small binary bootstraps the terminal UI (internal/ui) and exits with a
// non-zero status code if the UI returns an error. Keep this file minimal — it
// delegates real work to internal packages so it remains easy to test and to
// replace in deployments.
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