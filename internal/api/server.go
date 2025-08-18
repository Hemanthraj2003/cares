// Package api provides a REST API server for the CARES function execution platform.
// It exposes HTTP endpoints for function registration, listing, and invocation,
// integrating with the scheduler to execute functions on optimal worker nodes.
//
// Available endpoints:
//   - GET /functions - List all registered functions
//   - POST /functions - Register a new function
//   - GET /functions/{id} - Get function details by ID
//   - POST /invoke/{name} - Execute a function by name
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"cares/internal/cluster"
	"cares/internal/functions"
	"cares/internal/logging"
	"cares/internal/registry"
	"cares/internal/scheduler"
)

// Server represents the REST API server for function management and execution.
// It provides HTTP endpoints for function lifecycle management and coordinates
// with the scheduler to execute functions on worker nodes via gRPC.
//
// The server supports CORS for browser compatibility and provides JSON responses
// for all endpoints. It integrates with the function registry for persistence
// and the node registry for worker node management.
type Server struct {
	registry     *functions.Registry  // Function registry for storage and retrieval
	nodeRegistry *registry.NodeRegistry // Node registry for worker management
	scheduler    *scheduler.Scheduler    // Scheduler for optimal node selection
	server       *http.Server           // HTTP server instance
}

// NewServer creates a new REST API server with the provided function registry.
//
// The server is initialized with a function registry for persistence and
// a scheduler for worker node selection. The node registry can be set later
// using SetNodeRegistry method.
//
// Parameters:
//   - registry: Function registry for storing and retrieving function definitions
//
// Returns:
//   - *Server: Configured API server ready to handle HTTP requests
//
// Example usage:
//
//	funcRegistry := functions.NewRegistry()
//	apiServer := NewServer(funcRegistry)
//	apiServer.SetNodeRegistry(nodeRegistry)
//	err := apiServer.StartServer("8080")
func NewServer(registry *functions.Registry) *Server {
	return &Server{
		registry:  registry,
		scheduler: scheduler.NewScheduler(),
	}
}

// SetNodeRegistry sets the node registry for function execution
func (s *Server) SetNodeRegistry(nodeRegistry *registry.NodeRegistry) {
	s.nodeRegistry = nodeRegistry
}

// FunctionRequest represents the JSON payload for function registration
type FunctionRequest struct {
	Name        string `json:"name"`
	Image       string `json:"image"`
	Description string `json:"description,omitempty"`
}

// FunctionResponse represents the JSON response for function operations
type FunctionResponse struct {
	Status     string                `json:"status"`
	Message    string                `json:"message,omitempty"`
	Function   *functions.Function   `json:"function,omitempty"`
	Functions  []*functions.Function `json:"functions,omitempty"`
	InvokePath string                `json:"invoke_path,omitempty"`
}

// ErrorResponse represents error responses
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// StartServer starts the REST API server on the specified port
func (s *Server) StartServer(port string) error {
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/functions", s.handleFunctions)
	mux.HandleFunc("/functions/", s.handleFunctionByID)
	mux.HandleFunc("/invoke/", s.handleInvokeFunction)

	s.server = &http.Server{
		Addr:    ":" + port,
		Handler: s.corsMiddleware(mux),
	}

	logging.Info("REST API server starting on port %s", port)
	return s.server.ListenAndServe()
}

// corsMiddleware adds CORS headers for browser compatibility
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleFunctions handles /functions endpoint
func (s *Server) handleFunctions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.listFunctions(w, r)
	case "POST":
		s.createFunction(w, r)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleFunctionByID handles /functions/{id} endpoint
func (s *Server) handleFunctionByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.getFunction(w, r)
	case "DELETE":
		s.deleteFunction(w, r)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// createFunction handles POST /functions
func (s *Server) createFunction(w http.ResponseWriter, r *http.Request) {
	var req FunctionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Validate required fields
	if req.Name == "" {
		s.writeError(w, http.StatusBadRequest, "Function name is required")
		return
	}
	if req.Image == "" {
		s.writeError(w, http.StatusBadRequest, "Docker image is required")
		return
	}

	// Add function to registry
	function, err := s.registry.AddFunction(req.Name, req.Image, req.Description)
	if err != nil {
		s.writeError(w, http.StatusConflict, err.Error())
		return
	}

	// Success response
	response := FunctionResponse{
		Status:     "success",
		Message:    fmt.Sprintf("Function '%s' registered successfully", req.Name),
		Function:   function,
		InvokePath: fmt.Sprintf("/invoke/%s", req.Name),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// listFunctions handles GET /functions
func (s *Server) listFunctions(w http.ResponseWriter, r *http.Request) {
	functions := s.registry.GetAllFunctions()

	response := FunctionResponse{
		Status:    "success",
		Functions: functions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getFunction handles GET /functions/{id}
func (s *Server) getFunction(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	path := r.URL.Path
	if len(path) < 11 { // "/functions/" = 11 chars
		s.writeError(w, http.StatusBadRequest, "Function ID required")
		return
	}
	id := path[11:] // Get everything after "/functions/"

	function, exists := s.registry.GetFunction(id)
	if !exists {
		s.writeError(w, http.StatusNotFound, "Function not found")
		return
	}

	response := FunctionResponse{
		Status:   "success",
		Function: function,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// deleteFunction handles DELETE /functions/{id}
func (s *Server) deleteFunction(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	path := r.URL.Path
	if len(path) < 11 { // "/functions/" = 11 chars
		s.writeError(w, http.StatusBadRequest, "Function ID required")
		return
	}
	id := path[11:] // Get everything after "/functions/"

	if !s.registry.RemoveFunction(id) {
		s.writeError(w, http.StatusNotFound, "Function not found")
		return
	}

	response := FunctionResponse{
		Status:  "success",
		Message: "Function deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// writeError writes an error response
func (s *Server) writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Status:  "error",
		Message: message,
	}

	json.NewEncoder(w).Encode(response)
}

// handleInvokeFunction handles POST /invoke/{function_name} endpoint
func (s *Server) handleInvokeFunction(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract function name from URL path
	path := r.URL.Path
	if len(path) < 9 { // "/invoke/" = 8 chars
		s.writeError(w, http.StatusBadRequest, "Function name required")
		return
	}
	functionName := path[8:] // Get everything after "/invoke/"

	// Step 1: Lookup function in registry
	function, exists := s.registry.GetFunctionByName(functionName)
	if !exists {
		s.writeError(w, http.StatusNotFound, fmt.Sprintf("Function '%s' not found", functionName))
		return
	}

	// Step 2: Schedule execution (select optimal worker)
	if s.nodeRegistry == nil {
		s.writeError(w, http.StatusServiceUnavailable, "No worker nodes available")
		return
	}

	selectedNode, err := s.scheduler.SelectNodeForExecution(s.nodeRegistry)
	if err != nil {
		s.writeError(w, http.StatusServiceUnavailable, fmt.Sprintf("Failed to select worker: %v", err))
		return
	}

	logging.Info("Selected node '%s' for function '%s' execution", selectedNode.ID, functionName)

	// Step 3: Execute function on selected worker via gRPC
	result, err := s.executeOnWorker(selectedNode, function)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Execution failed: %v", err))
		return
	}

	// Step 4: Return result
	if !result.Success {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Function execution failed: %s", result.Error))
		return
	}

	// Return successful result
	response := map[string]interface{}{
		"status": "success",
		"output": result.Output,
		"node":   selectedNode.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// executeOnWorker executes a function on a specific worker node via gRPC
func (s *Server) executeOnWorker(node *registry.Node, function *functions.Function) (*cluster.FunctionResult, error) {
	// Connect to worker's gRPC server
	conn, err := grpc.Dial(node.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to worker %s: %v", node.ID, err)
	}
	defer conn.Close()

	// Create gRPC client
	client := cluster.NewClusterServiceClient(conn)

	// Call ExecuteFunction
	ctx := context.Background()
	req := &cluster.FunctionRequest{
		DockerImage:  function.Image,
		FunctionName: function.Name,
	}

	logging.Info("Executing function '%s' with image '%s' on worker '%s'", 
		function.Name, function.Image, node.ID)

	result, err := client.ExecuteFunction(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("gRPC call failed: %v", err)
	}

	return result, nil
}
