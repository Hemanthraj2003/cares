// Package cluster provides gRPC client implementation for worker nodes.
package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"cares/internal/metrics"
)

// Client represents a gRPC client for worker nodes to communicate with the orchestrator.
type Client struct {
	conn        *grpc.ClientConn
	client      ClusterServiceClient
	nodeID      string
	address     string
	hostname    string
	isConnected bool
}

// NewClient creates a new gRPC client instance.
func NewClient(hostname string) *Client {
	return &Client{
		nodeID:   uuid.New().String(),
		hostname: hostname,
	}
}

// Connect establishes a connection to the orchestrator at the given address.
func (c *Client) Connect(orchestratorAddr string) error {
	// Establish gRPC connection
	conn, err := grpc.Dial(orchestratorAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to orchestrator: %v", err)
	}

	c.conn = conn
	c.client = NewClusterServiceClient(conn)
	c.address = orchestratorAddr

	// Join the cluster
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	joinReq := &NodeInfo{
		NodeId:    c.nodeID,
		Address:   c.getLocalAddress(),
		Hostname:  c.hostname,
		Timestamp: time.Now().Unix(),
	}

	ack, err := c.client.JoinCluster(ctx, joinReq)
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to join cluster: %v", err)
	}

	if !ack.Success {
		c.conn.Close()
		return fmt.Errorf("cluster rejected join request: %s", ack.Message)
	}

	c.isConnected = true
	return nil
}

// StartHeartbeat begins sending periodic heartbeat messages with metrics.
// This function runs in a loop and should be called in a separate goroutine.
func (c *Client) StartHeartbeat(ctx context.Context) error {
	if !c.isConnected {
		return fmt.Errorf("not connected to orchestrator")
	}

	stream, err := c.client.Heartbeat(ctx)
	if err != nil {
		return fmt.Errorf("failed to establish heartbeat stream: %v", err)
	}

	// Goroutine to receive commands from orchestrator
	go func() {
		for {
			cmd, err := stream.Recv()
			if err != nil {
				return
			}
			
			// TODO: Handle commands in Phase 03
			_ = cmd // Suppress unused variable warning
		}
	}()

	// Send heartbeat messages every 2 seconds
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return stream.CloseSend()
		case <-ticker.C:
			// Collect current metrics
			cpu, err1 := metrics.GetCPUUsage()
			memory, err2 := metrics.GetMemoryUsage()

			status := "active"
			if err1 != nil || err2 != nil {
				status = "error"
			}

			// Send metrics to orchestrator
			metricsMsg := &NodeMetrics{
				NodeId:      c.nodeID,
				CpuUsage:    cpu,
				MemoryUsage: memory,
				Timestamp:   time.Now().Unix(),
				Status:      status,
			}

			if err := stream.Send(metricsMsg); err != nil {
				return err
			}
		}
	}
}

// Disconnect closes the connection to the orchestrator.
func (c *Client) Disconnect() error {
	if c.conn != nil {
		err := c.conn.Close()
		c.isConnected = false
		return err
	}
	return nil
}

// GetNodeID returns the unique identifier for this worker node.
func (c *Client) GetNodeID() string {
	return c.nodeID
}

// IsConnected returns true if the client is connected to an orchestrator.
func (c *Client) IsConnected() bool {
	return c.isConnected
}

// getLocalAddress returns a string representation of the local address.
// Returns localhost with a port for gRPC connections back to this worker
func (c *Client) getLocalAddress() string {
	// For Phase 4: Return localhost with default port so orchestrator can connect back
	// In production, this would be the actual network interface IP
	return "localhost:50052" // Different port from orchestrator (50051)
}
