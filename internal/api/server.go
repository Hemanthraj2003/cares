package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"cares/internal/functions"
)

// Server represents the REST API server for function management
type Server struct {
	registry *functions.Registry
	server   *http.Server
}

// NewServer creates a new REST API server
func NewServer(registry *functions.Registry) *Server {
	return &Server{
		registry: registry,
	}
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

	s.server = &http.Server{
		Addr:    ":" + port,
		Handler: s.corsMiddleware(mux),
	}

	log.Printf("REST API server starting on port %s", port)
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
