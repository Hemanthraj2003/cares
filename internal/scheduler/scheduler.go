// Package scheduler provides intelligent worker node selection for function execution.
// It implements a cost-based scheduling algorithm that considers CPU and memory
// usage to distribute workload optimally across available worker nodes.
package scheduler

import (
	"fmt"

	"cares/internal/registry"
)

// Scheduler handles worker node selection for function execution.
// It uses a weighted cost model to select the most suitable worker node
// based on current resource utilization metrics.
type Scheduler struct{}

// NewScheduler creates a new scheduler instance.
//
// Returns a configured Scheduler ready to select worker nodes for task execution.
//
// Example usage:
//
//	scheduler := NewScheduler()
//	node, err := scheduler.SelectNodeForExecution(nodeRegistry)
func NewScheduler() *Scheduler {
	return &Scheduler{}
}

// SelectNodeForExecution selects the optimal worker node for function execution
// based on a cost model that considers CPU and memory usage.
//
// The selection algorithm:
//  1. Filters for only active worker nodes
//  2. Calculates cost score: (cpu_usage * 0.5) + (memory_usage * 0.5)
//  3. Selects the node with the lowest cost score (least utilized)
//
// Parameters:
//   - nodeRegistry: Registry containing available worker nodes with their metrics
//
// Returns:
//   - *registry.Node: The selected worker node for execution
//   - error: Error if no nodes are available or registry is invalid
//
// Example usage:
//
//	selectedNode, err := scheduler.SelectNodeForExecution(nodeRegistry)
//	if err != nil {
//	    return fmt.Errorf("no workers available: %w", err)
//	}
//	// Execute function on selectedNode
func (s *Scheduler) SelectNodeForExecution(nodeRegistry *registry.NodeRegistry) (*registry.Node, error) {
	if nodeRegistry == nil {
		return nil, fmt.Errorf("node registry is nil")
	}
	
	nodes := nodeRegistry.GetAllNodes()
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no worker nodes available")
	}
	
	// Filter for active nodes only
	var activeNodes []*registry.Node
	for _, node := range nodes {
		if string(node.Status) == "Active" {
			activeNodes = append(activeNodes, node)
		}
	}
	
	if len(activeNodes) == 0 {
		return nil, fmt.Errorf("no active worker nodes available")
	}
	
	// Find the node with the lowest cost score
	var bestNode *registry.Node
	var lowestScore float64 = -1
	
	for _, node := range activeNodes {
		// Cost model: (cpu_usage * 0.5) + (memory_usage * 0.5)
		score := (node.CPUUsage * 0.5) + (node.MemoryUsage * 0.5)
		
		if lowestScore == -1 || score < lowestScore {
			lowestScore = score
			bestNode = node
		}
	}
	
	if bestNode == nil {
		return nil, fmt.Errorf("failed to select optimal node")
	}
	
	return bestNode, nil
}
