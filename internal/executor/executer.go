// Package executor provides functionality to execute external containers (Docker)
// and capture their output. The package intentionally keeps a minimal surface
// area: a single exported RunContainer function used by higher-level logic.
package executor

import (
	"os/exec"
)

// RunContainer runs the specified Docker image using the local Docker daemon.
//
//	imageName: The name (and optional tag) of the Docker image to run. The image must be available locally
//				or pullable from a configured registry.
//
// Returns the combined output (stdout and stderr) from the container, and any error encountered during execution.
//
// Example usage:
//		output, err := executor.RunContainer("alpine:latest")
//		if err != nil {
//			// handle error
//		}
//		fmt.Println(output)
func RunContainer(imageName string) (string, error) {
	if imageName == "" {
		return "", &exec.Error{Name: "docker", Err: exec.ErrNotFound}
	}
	cmd := exec.Command("docker", "run", "--rm", imageName)
	output, err := cmd.CombinedOutput()
	return string(output), err
}