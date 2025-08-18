package scheduler

import (
	"fmt"

	"cares/internal/registry"
)

// Scheduler handles worker node selection for function execution
type Scheduler struct{}

// NewScheduler creates a new scheduler instance
func NewScheduler() *Scheduler {
	return &Scheduler{}
}

// SelectNodeForExecution selects the optimal worker node based on resource usage
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
