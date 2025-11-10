package repository

import (
	"fmt"
	"sync"

	"github.com/1PercentSync/vibox/internal/domain"
	"github.com/1PercentSync/vibox/pkg/utils"
)

// WorkspaceRepository defines the interface for workspace storage operations
type WorkspaceRepository interface {
	Create(ws *domain.Workspace) error
	Get(id string) (*domain.Workspace, error)
	List() ([]*domain.Workspace, error)
	Update(ws *domain.Workspace) error
	Delete(id string) error
}

// MemoryRepository implements WorkspaceRepository using in-memory storage
type MemoryRepository struct {
	mu    sync.RWMutex
	store map[string]*domain.Workspace
}

// NewMemoryRepository creates a new in-memory repository
func NewMemoryRepository() *MemoryRepository {
	utils.Info("Initializing memory repository")
	return &MemoryRepository{
		store: make(map[string]*domain.Workspace),
	}
}

// Create adds a new workspace to the repository
func (r *MemoryRepository) Create(ws *domain.Workspace) error {
	if ws == nil {
		return fmt.Errorf("workspace cannot be nil")
	}
	if ws.ID == "" {
		return fmt.Errorf("workspace ID cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.store[ws.ID]; exists {
		utils.Warn("Attempted to create duplicate workspace", "id", ws.ID)
		return fmt.Errorf("workspace with ID %s already exists", ws.ID)
	}

	r.store[ws.ID] = ws
	utils.Info("Workspace created in repository", "id", ws.ID, "name", ws.Name)
	return nil
}

// Get retrieves a workspace by ID
func (r *MemoryRepository) Get(id string) (*domain.Workspace, error) {
	if id == "" {
		return nil, fmt.Errorf("workspace ID cannot be empty")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	ws, exists := r.store[id]
	if !exists {
		utils.Debug("Workspace not found in repository", "id", id)
		return nil, fmt.Errorf("workspace with ID %s not found", id)
	}

	utils.Debug("Workspace retrieved from repository", "id", id)
	return ws, nil
}

// List returns all workspaces in the repository
func (r *MemoryRepository) List() ([]*domain.Workspace, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	workspaces := make([]*domain.Workspace, 0, len(r.store))
	for _, ws := range r.store {
		workspaces = append(workspaces, ws)
	}

	utils.Debug("Listed workspaces from repository", "count", len(workspaces))
	return workspaces, nil
}

// Update modifies an existing workspace in the repository
func (r *MemoryRepository) Update(ws *domain.Workspace) error {
	if ws == nil {
		return fmt.Errorf("workspace cannot be nil")
	}
	if ws.ID == "" {
		return fmt.Errorf("workspace ID cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.store[ws.ID]; !exists {
		utils.Warn("Attempted to update non-existent workspace", "id", ws.ID)
		return fmt.Errorf("workspace with ID %s not found", ws.ID)
	}

	r.store[ws.ID] = ws
	utils.Info("Workspace updated in repository", "id", ws.ID, "status", ws.Status)
	return nil
}

// Delete removes a workspace from the repository
func (r *MemoryRepository) Delete(id string) error {
	if id == "" {
		return fmt.Errorf("workspace ID cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.store[id]; !exists {
		utils.Warn("Attempted to delete non-existent workspace", "id", id)
		return fmt.Errorf("workspace with ID %s not found", id)
	}

	delete(r.store, id)
	utils.Info("Workspace deleted from repository", "id", id)
	return nil
}
