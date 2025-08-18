// Package executor provides functionality to execute external containers (Docker)
// and capture their output. It handles Docker daemon management and supports
// both local images and URL-based image references.
package executor

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"cares/internal/logging"
)

// ensureDockerRunning checks if Docker daemon is running and starts it if needed.
// This function attempts to start Docker using systemctl on Linux systems.
func ensureDockerRunning() error {
	// First check if Docker is already running
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err == nil {
		logging.Debug("Docker daemon is already running")
		return nil
	}

	logging.Info("Docker daemon not running, attempting to start...")
	
	// Try to start Docker daemon using systemctl
	startCmd := exec.Command("sudo", "systemctl", "start", "docker")
	if err := startCmd.Run(); err != nil {
		return fmt.Errorf("failed to start Docker daemon: %w", err)
	}

	// Wait a moment for Docker to fully start
	time.Sleep(3 * time.Second)

	// Verify Docker is now running
	verifyCmd := exec.Command("docker", "info")
	if err := verifyCmd.Run(); err != nil {
		return fmt.Errorf("Docker daemon failed to start properly: %w", err)
	}

	logging.Info("Docker daemon started successfully")
	return nil
}

// pullImageIfNeeded checks if an image exists locally and pulls it if not.
// Supports both standard image names and URL-based registry paths.
func pullImageIfNeeded(imageName string) error {
	// Check if image exists locally
	cmd := exec.Command("docker", "image", "inspect", imageName)
	if err := cmd.Run(); err == nil {
		logging.Debug("Image '%s' found locally", imageName)
		return nil
	}

	logging.Info("Pulling image '%s'...", imageName)
	
	// Pull the image
	pullCmd := exec.Command("docker", "pull", imageName)
	output, err := pullCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to pull image '%s': %w\nOutput: %s", imageName, err, string(output))
	}

	logging.Info("Successfully pulled image '%s'", imageName)
	return nil
}

// normalizeImageName handles URL-based image names and converts them to proper Docker format.
// Examples:
//   - "https://registry.hub.docker.com/nginx:latest" -> "nginx:latest"
//   - "ghcr.io/user/repo:tag" -> "ghcr.io/user/repo:tag" (unchanged)
//   - "nginx" -> "nginx:latest" (add latest tag)
func normalizeImageName(imageName string) string {
	// Handle HTTP/HTTPS URLs by extracting the path
	if strings.HasPrefix(imageName, "http://") || strings.HasPrefix(imageName, "https://") {
		// Extract the path after the domain
		parts := strings.Split(imageName, "/")
		if len(parts) >= 4 {
			// Skip protocol and domain, join the rest
			imageName = strings.Join(parts[3:], "/")
		}
	}

	// Add :latest tag if no tag is specified
	if !strings.Contains(imageName, ":") {
		imageName += ":latest"
	}

	return imageName
}

// RunContainer runs the specified Docker image using the local Docker daemon.
// It automatically ensures Docker is running, normalizes image names, and pulls
// images if they're not available locally.
//
// Parameters:
//   - imageName: The name, tag, or URL of the Docker image to run
//
// Returns the combined output (stdout and stderr) from the container, and any error encountered during execution.
//
// The function supports multiple image formats:
//   - Standard names: "alpine", "nginx:1.21"
//   - Registry URLs: "ghcr.io/user/repo:tag"
//   - HTTP URLs: "https://registry.hub.docker.com/nginx:latest"
//
// Example usage:
//
//	output, err := executor.RunContainer("alpine:latest")
//	if err != nil {
//	    // handle error
//	}
//	fmt.Println(output)
func RunContainer(imageName string) (string, error) {
	if imageName == "" {
		return "", fmt.Errorf("image name cannot be empty")
	}

	// Ensure Docker daemon is running
	if err := ensureDockerRunning(); err != nil {
		return "", fmt.Errorf("Docker daemon error: %w", err)
	}

	// Normalize the image name
	normalizedImage := normalizeImageName(imageName)
	logging.Debug("Normalized image name: %s -> %s", imageName, normalizedImage)

	// Pull image if not available locally
	if err := pullImageIfNeeded(normalizedImage); err != nil {
		return "", fmt.Errorf("image pull error: %w", err)
	}

	// Run the container
	logging.Debug("Running container with image: %s", normalizedImage)
	cmd := exec.Command("docker", "run", "--rm", normalizedImage)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return string(output), fmt.Errorf("container execution failed: %w", err)
	}

	logging.Debug("Container executed successfully, output length: %d bytes", len(output))
	return string(output), nil
}