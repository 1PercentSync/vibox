package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

// PersistentData represents the data structure saved to disk
type PersistentData struct {
	Workspaces map[string]*domain.Workspace `json:"workspaces"`
}

// FileRepository implements WorkspaceRepository with file-based persistence
type FileRepository struct {
	mu       sync.RWMutex
	store    map[string]*domain.Workspace
	dataFile string
}

// NewWorkspaceRepository creates a new file-based repository
// dataDir: directory where the workspaces.json file will be stored
func NewWorkspaceRepository(dataDir string) (*FileRepository, error) {
	utils.Info("Initializing file-based workspace repository", "dataDir", dataDir)

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		utils.Error("Failed to create data directory", "error", err, "dataDir", dataDir)
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	repo := &FileRepository{
		store:    make(map[string]*domain.Workspace),
		dataFile: filepath.Join(dataDir, "workspaces.json"),
	}

	// Load existing data from disk
	if err := repo.load(); err != nil {
		// If file doesn't exist or is corrupted, start with empty store
		if !os.IsNotExist(err) {
			utils.Warn("Failed to load existing data, starting fresh", "error", err)
		} else {
			utils.Info("No existing data found, starting with empty repository")
		}
		repo.store = make(map[string]*domain.Workspace)
	} else {
		utils.Info("Loaded workspaces from disk", "count", len(repo.store))
	}

	return repo, nil
}

// NewMemoryRepository creates a new file-based repository (alias for backward compatibility)
func NewMemoryRepository() *FileRepository {
	// For backward compatibility, use default data directory
	repo, err := NewWorkspaceRepository("./data")
	if err != nil {
		utils.Error("Failed to initialize repository", "error", err)
		// Return empty repository without persistence
		return &FileRepository{
			store:    make(map[string]*domain.Workspace),
			dataFile: "",
		}
	}
	return repo
}

// save writes all workspaces to disk
func (r *FileRepository) save() error {
	if r.dataFile == "" {
		// Persistence disabled
		return nil
	}

	// Prepare data for serialization
	data := PersistentData{
		Workspaces: r.store,
	}

	// Marshal to JSON with indentation for readability
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		utils.Error("Failed to marshal workspace data", "error", err)
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Atomic write: write to temp file first, then rename
	tmpFile := r.dataFile + ".tmp"
	if err := os.WriteFile(tmpFile, jsonData, 0644); err != nil {
		utils.Error("Failed to write temporary file", "error", err, "file", tmpFile)
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	// Rename temp file to actual file (atomic operation)
	if err := os.Rename(tmpFile, r.dataFile); err != nil {
		utils.Error("Failed to rename temporary file", "error", err)
		// Clean up temp file
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	utils.Debug("Workspace data saved to disk", "file", r.dataFile, "count", len(r.store))
	return nil
}

// load reads workspaces from disk
func (r *FileRepository) load() error {
	if r.dataFile == "" {
		// Persistence disabled
		return nil
	}

	// Read file
	jsonData, err := os.ReadFile(r.dataFile)
	if err != nil {
		return err
	}

	// Unmarshal JSON
	var data PersistentData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		utils.Error("Failed to unmarshal workspace data", "error", err)
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	// Load workspaces into store
	if data.Workspaces != nil {
		r.store = data.Workspaces
	} else {
		r.store = make(map[string]*domain.Workspace)
	}

	utils.Debug("Workspace data loaded from disk", "file", r.dataFile, "count", len(r.store))
	return nil
}

// Create adds a new workspace to the repository and persists to disk
func (r *FileRepository) Create(ws *domain.Workspace) error {
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

	// Persist to disk
	if err := r.save(); err != nil {
		// Rollback in-memory change
		delete(r.store, ws.ID)
		return fmt.Errorf("failed to persist workspace: %w", err)
	}

	utils.Info("Workspace created in repository", "id", ws.ID, "name", ws.Name)
	return nil
}

// Get retrieves a workspace by ID
func (r *FileRepository) Get(id string) (*domain.Workspace, error) {
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
func (r *FileRepository) List() ([]*domain.Workspace, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	workspaces := make([]*domain.Workspace, 0, len(r.store))
	for _, ws := range r.store {
		workspaces = append(workspaces, ws)
	}

	utils.Debug("Listed workspaces from repository", "count", len(workspaces))
	return workspaces, nil
}

// Update modifies an existing workspace in the repository and persists to disk
func (r *FileRepository) Update(ws *domain.Workspace) error {
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

	// Store old workspace for rollback
	oldWs := r.store[ws.ID]
	r.store[ws.ID] = ws

	// Persist to disk
	if err := r.save(); err != nil {
		// Rollback in-memory change
		r.store[ws.ID] = oldWs
		return fmt.Errorf("failed to persist workspace: %w", err)
	}

	utils.Info("Workspace updated in repository", "id", ws.ID, "status", ws.Status)
	return nil
}

// Delete removes a workspace from the repository and persists to disk
func (r *FileRepository) Delete(id string) error {
	if id == "" {
		return fmt.Errorf("workspace ID cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	ws, exists := r.store[id]
	if !exists {
		utils.Warn("Attempted to delete non-existent workspace", "id", id)
		return fmt.Errorf("workspace with ID %s not found", id)
	}

	delete(r.store, id)

	// Persist to disk
	if err := r.save(); err != nil {
		// Rollback in-memory change
		r.store[id] = ws
		return fmt.Errorf("failed to persist workspace deletion: %w", err)
	}

	utils.Info("Workspace deleted from repository", "id", id)
	return nil
}
