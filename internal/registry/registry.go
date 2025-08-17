// Package registry provides thread-safe management of cluster nodes.
// It maintains the state of all worker nodes connected to the orchestrator,
// including their resource metrics and connection status.
package registry

import (
	"sync"
	"time"
)

// NodeStatus represents the current status of a node in the cluster.
type NodeStatus string

const (
	// NodeStatusActive indicates the node is connected and responsive
	NodeStatusActive NodeStatus = "Active"
	// NodeStatusDisconnected indicates the node has lost connection
	NodeStatusDisconnected NodeStatus = "Disconnected"
	// NodeStatusJoining indicates the node is in the process of joining
	NodeStatusJoining NodeStatus = "Joining"
)

// Node represents a worker node in the cluster with its current state and metrics.
type Node struct {
	ID           string     `json:"id"`
	Address      string     `json:"address"`
	Hostname     string     `json:"hostname"`
	Status       NodeStatus `json:"status"`
	CPUUsage     float64    `json:"cpu_usage"`
	MemoryUsage  float64    `json:"memory_usage"`
	LastSeen     time.Time  `json:"last_seen"`
	JoinedAt     time.Time  `json:"joined_at"`
}

// NodeRegistry provides thread-safe management of cluster nodes.
// It maintains a registry of all nodes and their current state.
type NodeRegistry struct {
	mu    sync.RWMutex
	nodes map[string]*Node
}

// NewNodeRegistry creates a new thread-safe node registry.
func NewNodeRegistry() *NodeRegistry {
	return &NodeRegistry{
		nodes: make(map[string]*Node),
	}
}

// AddNode adds a new node to the registry or updates an existing one.
// It's thread-safe and can be called from multiple goroutines.
func (nr *NodeRegistry) AddNode(id, address, hostname string) *Node {
	nr.mu.Lock()
	defer nr.mu.Unlock()

	now := time.Now()
	node := &Node{
		ID:          id,
		Address:     address,
		Hostname:    hostname,
		Status:      NodeStatusJoining,
		CPUUsage:    0.0,
		MemoryUsage: 0.0,
		LastSeen:    now,
		JoinedAt:    now,
	}

	nr.nodes[id] = node
	return node
}

// UpdateMetrics updates the resource metrics for a specific node.
// Returns true if the node exists, false otherwise.
func (nr *NodeRegistry) UpdateMetrics(nodeID string, cpuUsage, memoryUsage float64) bool {
	nr.mu.Lock()
	defer nr.mu.Unlock()

	node, exists := nr.nodes[nodeID]
	if !exists {
		return false
	}

	node.CPUUsage = cpuUsage
	node.MemoryUsage = memoryUsage
	node.LastSeen = time.Now()
	node.Status = NodeStatusActive

	return true
}

// GetNode retrieves a node by ID. Returns nil if not found.
func (nr *NodeRegistry) GetNode(nodeID string) *Node {
	nr.mu.RLock()
	defer nr.mu.RUnlock()

	node, exists := nr.nodes[nodeID]
	if !exists {
		return nil
	}

	// Return a copy to avoid concurrent access issues
	nodeCopy := *node
	return &nodeCopy
}

// GetAllNodes returns a snapshot of all nodes in the registry.
// The returned slice contains copies of the nodes to prevent concurrent access issues.
func (nr *NodeRegistry) GetAllNodes() []*Node {
	nr.mu.RLock()
	defer nr.mu.RUnlock()

	nodes := make([]*Node, 0, len(nr.nodes))
	for _, node := range nr.nodes {
		// Return copies to avoid concurrent access issues
		nodeCopy := *node
		nodes = append(nodes, &nodeCopy)
	}

	return nodes
}

// RemoveNode removes a node from the registry.
// Returns true if the node was removed, false if it didn't exist.
func (nr *NodeRegistry) RemoveNode(nodeID string) bool {
	nr.mu.Lock()
	defer nr.mu.Unlock()

	_, exists := nr.nodes[nodeID]
	if exists {
		delete(nr.nodes, nodeID)
	}

	return exists
}

// MarkDisconnected marks a node as disconnected but keeps it in the registry.
// This allows the orchestrator to show disconnected nodes in the UI.
func (nr *NodeRegistry) MarkDisconnected(nodeID string) bool {
	nr.mu.Lock()
	defer nr.mu.Unlock()

	node, exists := nr.nodes[nodeID]
	if !exists {
		return false
	}

	node.Status = NodeStatusDisconnected
	return true
}

// GetNodeCount returns the total number of nodes in the registry.
func (nr *NodeRegistry) GetNodeCount() int {
	nr.mu.RLock()
	defer nr.mu.RUnlock()

	return len(nr.nodes)
}

// GetActiveNodeCount returns the number of active (connected) nodes.
func (nr *NodeRegistry) GetActiveNodeCount() int {
	nr.mu.RLock()
	defer nr.mu.RUnlock()

	count := 0
	for _, node := range nr.nodes {
		if node.Status == NodeStatusActive {
			count++
		}
	}

	return count
}
