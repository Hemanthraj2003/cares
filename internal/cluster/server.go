// Package cluster provides gRPC server and client implementations for
// CARES cluster communication. It handles node registration, heartbeat
// streams, and orchestrator-worker communication.
package cluster

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"

	"google.golang.org/grpc"

	"cares/internal/registry"
)

// Server implements the gRPC ClusterService for the orchestrator.
type Server struct {
	UnimplementedClusterServiceServer
	registry *registry.NodeRegistry
	listeners map[string]chan *OrchestratorCommand // nodeID -> command channel
	mu       sync.RWMutex
}

// NewServer creates a new gRPC server instance with an empty node registry.
func NewServer() *Server {
	return &Server{
		registry:  registry.NewNodeRegistry(),
		listeners: make(map[string]chan *OrchestratorCommand),
	}
}

// GetRegistry returns the node registry for access by the UI layer.
func (s *Server) GetRegistry() *registry.NodeRegistry {
	return s.registry
}

// JoinCluster handles worker node registration requests.
func (s *Server) JoinCluster(ctx context.Context, nodeInfo *NodeInfo) (*Acknowledgement, error) {
	// Add node to registry
	s.registry.AddNode(nodeInfo.NodeId, nodeInfo.Address, nodeInfo.Hostname)
	
	// Create command channel for this node
	s.mu.Lock()
	s.listeners[nodeInfo.NodeId] = make(chan *OrchestratorCommand, 10)
	s.mu.Unlock()

	return &Acknowledgement{
		Success: true,
		Message: fmt.Sprintf("Welcome to cluster, node %s", nodeInfo.NodeId),
	}, nil
}

// Heartbeat handles bidirectional streaming for worker heartbeats.
func (s *Server) Heartbeat(stream grpc.BidiStreamingServer[NodeMetrics, OrchestratorCommand]) error {
	var nodeID string

	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	// Handle incoming metrics from worker
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Receive metrics from worker (blocking call)
		metrics, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		nodeID = metrics.NodeId

		// Update node metrics in registry
		s.registry.UpdateMetrics(nodeID, float64(metrics.CpuUsage), float64(metrics.MemoryUsage))

		// Send commands to worker (if any)
		s.mu.RLock()
		commandChan, exists := s.listeners[nodeID]
		s.mu.RUnlock()

		if exists {
			select {
			case cmd := <-commandChan:
				if err := stream.Send(cmd); err != nil {
					break
				}
			default:
				// No commands to send, continue
			}
		}

		// Check if node exists in registry, add if missing
		if s.registry.GetNode(nodeID) == nil {
			// This shouldn't happen normally, but handle gracefully
		}
	}

	// Cleanup when stream ends
	if nodeID != "" {
		s.mu.Lock()
		delete(s.listeners, nodeID)
		s.mu.Unlock()
		
		s.registry.MarkDisconnected(nodeID)
	}

	return nil
}// StartServer starts the gRPC server on the specified port.
// This function blocks until the server is stopped.
func (s *Server) StartServer(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	RegisterClusterServiceServer(grpcServer, s)

	return grpcServer.Serve(lis)
}
