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
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	ContainerID string          `json:"container_id"`
	Status      WorkspaceStatus `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	Config      WorkspaceConfig `json:"config"`
	Error       string          `json:"error,omitempty"`
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
