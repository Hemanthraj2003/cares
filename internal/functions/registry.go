package functions

import (
	"fmt"
	"sync"
	"time"

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

// NewRegistry creates a new function registry
func NewRegistry() *Registry {
	return &Registry{
		functions: make(map[string]*Function),
	}
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

// GetAllFunctions returns a snapshot of all functions in the registry
func (r *Registry) GetAllFunctions() []*Function {
	r.mu.RLock()
	defer r.mu.RUnlock()

	functions := make([]*Function, 0, len(r.functions))
	for _, fn := range r.functions {
		// Return copies to prevent concurrent access issues
		fnCopy := *fn
		functions = append(functions, &fnCopy)
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
	return true
}
