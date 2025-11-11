package domain

import "time"

// WorkspaceStatus represents the current status of a workspace
type WorkspaceStatus string

const (
	StatusCreating WorkspaceStatus = "creating"
	StatusRunning  WorkspaceStatus = "running"
	StatusStopped  WorkspaceStatus = "stopped"
	StatusError    WorkspaceStatus = "error"
)

// Workspace represents a development workspace with a Docker container
type Workspace struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	ContainerID string            `json:"container_id,omitempty"` // Runtime field, not persisted
	Status      WorkspaceStatus   `json:"status,omitempty"`       // Runtime field, not persisted
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at,omitempty"` // Runtime field, for API responses
	Config      WorkspaceConfig   `json:"config"`
	Ports       map[string]string `json:"ports,omitempty"` // Port label mappings (port number -> service name)
	Error       string            `json:"error,omitempty"` // Runtime field, not persisted
}

// WorkspaceConfig holds configuration for a workspace
type WorkspaceConfig struct {
	Image   string   `json:"image"`
	Scripts []Script `json:"scripts,omitempty"`
}

// Script represents an initialization script to be executed in the workspace
type Script struct {
	Name    string `json:"name"`
	Content string `json:"content"`
	Order   int    `json:"order"`
}
