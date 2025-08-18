// Package logging provides a centralized logging system that prevents interference
// with the TUI by directing logs to files instead of stdout/stderr when in TUI mode.
package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	// Logger is the global logger instance
	Logger *log.Logger
	logFile *os.File
	isTUIMode bool
)

// InitLogger initializes the logging system
// If tui is true, logs go to a file to avoid interfering with the TUI
// If tui is false, logs go to stderr for debugging
func InitLogger(tui bool) error {
	isTUIMode = tui
	
	if tui {
		// Create logs directory if it doesn't exist
		logDir := "logs"
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
		
		// Open log file
		var err error
		logFile, err = os.OpenFile(filepath.Join(logDir, "cares.log"), 
			os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		
		Logger = log.New(logFile, "", log.LstdFlags|log.Lshortfile)
	} else {
		// Development mode - log to stderr
		Logger = log.New(os.Stderr, "[CARES] ", log.LstdFlags|log.Lshortfile)
	}
	
	return nil
}

// Close closes the log file if it was opened
func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

// Info logs an info message
func Info(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Printf("[INFO] "+format, args...)
	}
}

// Error logs an error message
func Error(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Printf("[ERROR] "+format, args...)
	}
}

// Debug logs a debug message (only in non-TUI mode)
func Debug(format string, args ...interface{}) {
	if Logger != nil && !isTUIMode {
		Logger.Printf("[DEBUG] "+format, args...)
	}
}

// Warn logs a warning message
func Warn(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Printf("[WARN] "+format, args...)
	}
}
