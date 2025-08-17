// Package metrics provides utilities for collecting system resource usage
// statistics such as CPU and memory utilization. The functions in this package
// are safe to call from other packages and return values in percentage units
// (0.0 - 100.0).
package metrics

import (
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// GetCPUUsage returns the current CPU usage percentage of the system.
//
// It samples the CPU usage over a 1-second interval and returns the average usage
// as a float64 value (0.0 - 100.0). If an error occurs or no data is available, it returns 0 and the error.
//
// Example usage:
//     cpu, err := metrics.GetCPUUsage()
//     if err != nil {
//         // handle error
//     }
//     fmt.Printf("CPU Usage: %.2f%%\n", cpu)
func GetCPUUsage() (float64, error) {
	percentage, err := cpu.Percent(time.Second, false)
	if err != nil || len(percentage) == 0 {
		return 0, err
	}
	return percentage[0], nil
}

// GetMemoryUsage returns the current memory usage percentage of the system.
//
// It queries the system's virtual memory statistics and returns the percentage of memory used
// as a float64 value (0.0 - 100.0). If an error occurs, it returns 0 and the error.
//
// Example usage:
//     mem, err := metrics.GetMemoryUsage()
//     if err != nil {
//         // handle error
//     }
//     fmt.Printf("Memory Usage: %.2f%%\n", mem)
func GetMemoryUsage() (float64, error) {
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return 0, err
	}
	return vmStat.UsedPercent, nil
}