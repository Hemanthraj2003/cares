package functions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"cares/internal/logging"

	"github.com/google/uuid"
)

// Function represents a registered function in the system
type Function struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Image       string    `json:"image"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	Status      string    `json:"status"` // "active", "inactive"
}

// Registry provides thread-safe management of registered functions
type Registry struct {
	mu        sync.RWMutex
	functions map[string]*Function
}

// The default storage file path
const DefaultStoragePath = "data/functions.json"

// NewRegistry creates a new function registry
func NewRegistry() *Registry {
	registry := &Registry{
		functions: make(map[string]*Function),
	}
	
	// Try to load from default storage file
	err := registry.LoadFromFile(DefaultStoragePath)
	if err != nil {
		// Just log the error, don't fail
		logging.Warn("Could not load function registry: %v", err)
	}
	
	return registry
}

// AddFunction adds a new function to the registry
func (r *Registry) AddFunction(name, image, description string) (*Function, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if function name already exists
	for _, fn := range r.functions {
		if fn.Name == name {
			return nil, fmt.Errorf("function with name '%s' already exists", name)
		}
	}

	// Create new function
	function := &Function{
		ID:          uuid.New().String(),
		Name:        name,
		Image:       image,
		Description: description,
		CreatedAt:   time.Now(),
		Status:      "active",
	}

	r.functions[function.ID] = function
	
	// Save changes to file
	go r.SaveToFile(DefaultStoragePath) // Run in background to avoid blocking
	
	return function, nil
}

// GetFunction retrieves a function by ID
func (r *Registry) GetFunction(id string) (*Function, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fn, exists := r.functions[id]
	if !exists {
		return nil, false
	}

	// Return a copy to prevent concurrent access issues
	fnCopy := *fn
	return &fnCopy, true
}

// GetFunctionByName retrieves a function by name
func (r *Registry) GetFunctionByName(name string) (*Function, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, fn := range r.functions {
		if fn.Name == name {
			// Return a copy to prevent concurrent access issues
			fnCopy := *fn
			return &fnCopy, true
		}
	}

	return nil, false
}

// GetAllFunctions returns a snapshot of all functions in the registry sorted by creation time
func (r *Registry) GetAllFunctions() []*Function {
	r.mu.RLock()
	defer r.mu.RUnlock()

	functions := make([]*Function, 0, len(r.functions))
	for _, fn := range r.functions {
		// Return copies to prevent concurrent access issues
		fnCopy := *fn
		functions = append(functions, &fnCopy)
	}

	// Sort functions by creation time to ensure consistent order
	for i := 0; i < len(functions)-1; i++ {
		for j := i + 1; j < len(functions); j++ {
			if functions[i].CreatedAt.After(functions[j].CreatedAt) {
				functions[i], functions[j] = functions[j], functions[i]
			}
		}
	}

	return functions
}

// RemoveFunction removes a function from the registry
func (r *Registry) RemoveFunction(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.functions[id]
	if exists {
		delete(r.functions, id)
		
		// Save changes to file
		go r.SaveToFile(DefaultStoragePath) // Run in background to avoid blocking
	}

	return exists
}

// GetFunctionCount returns the total number of functions
func (r *Registry) GetFunctionCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.functions)
}

// UpdateFunctionStatus updates the status of a function
func (r *Registry) UpdateFunctionStatus(id, status string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	fn, exists := r.functions[id]
	if !exists {
		return false
	}

	fn.Status = status
	
	// Save changes to file
	go r.SaveToFile(DefaultStoragePath) // Run in background to avoid blocking
	
	return true
}

// SaveToFile saves the registry to a JSON file
func (r *Registry) SaveToFile(filePath string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}
	
	// Convert registry to a slice for serialization
	functions := r.GetAllFunctions()
	
	// Marshal to JSON
	data, err := json.MarshalIndent(functions, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %v", err)
	}
	
	// Write to file
	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write registry file: %v", err)
	}
	
	return nil
}

// LoadFromFile loads the registry from a JSON file
func (r *Registry) LoadFromFile(filePath string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File doesn't exist, but that's not an error
		return nil
	}
	
	// Read file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read registry file: %v", err)
	}
	
	// Unmarshal JSON
	var functions []*Function
	if err := json.Unmarshal(data, &functions); err != nil {
		return fmt.Errorf("failed to unmarshal registry: %v", err)
	}
	
	// Lock and update registry
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Clear existing functions
	r.functions = make(map[string]*Function)
	
	// Add loaded functions
	for _, fn := range functions {
		r.functions[fn.ID] = fn
	}
	
	return nil
}
