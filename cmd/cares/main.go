package main

import (
	"cares/internal/executor"
	"cares/internal/metrics"
	"fmt"
)

func main() {
	fmt.Println("Hello, CARES!")

	// Example: run any image, e.g., "hello-world" or "alpine"
	imageName := "hello-world" // You can change this to any local image
	output, err := executor.RunContainer(imageName)
	if err != nil {
		fmt.Printf("Error running container '%s': %v\n", imageName, err)
	}
	fmt.Printf("Container output:\n%s\n", output)

	cpu, err := metrics.GetCPUUsage()
	if err != nil {
		fmt.Println("Error getting CPU usage:", err)
	} else {
		fmt.Printf("CPU Usage: %.2f%%\n", cpu)
	}

	mem, err := metrics.GetMemoryUsage()
	if err != nil {
		fmt.Println("Error getting memory usage:", err)
	} else {
		fmt.Printf("Memory Usage: %.2f%%\n", mem)
	}
}