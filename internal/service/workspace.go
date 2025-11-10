package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/1PercentSync/vibox/internal/config"
	"github.com/1PercentSync/vibox/internal/domain"
	"github.com/1PercentSync/vibox/internal/repository"
	"github.com/1PercentSync/vibox/pkg/utils"
)

// CreateWorkspaceRequest represents a request to create a new workspace
type CreateWorkspaceRequest struct {
	Name    string          `json:"name" binding:"required"`
	Image   string          `json:"image"`
	Scripts []domain.Script `json:"scripts,omitempty"`
}

// WorkspaceService handles workspace management operations
type WorkspaceService struct {
	dockerSvc *DockerService
	repo      repository.WorkspaceRepository
	config    *config.Config
}

// NewWorkspaceService creates a new workspace service instance
func NewWorkspaceService(dockerSvc *DockerService, repo repository.WorkspaceRepository, cfg *config.Config) *WorkspaceService {
	utils.Info("Initializing workspace service")
	return &WorkspaceService{
		dockerSvc: dockerSvc,
		repo:      repo,
		config:    cfg,
	}
}

// CreateWorkspace creates a new workspace with a Docker container
func (s *WorkspaceService) CreateWorkspace(ctx context.Context, req CreateWorkspaceRequest) (*domain.Workspace, error) {
	utils.Info("Creating workspace", "name", req.Name)

	// Generate workspace ID
	workspaceID := utils.GenerateID()
	utils.Debug("Generated workspace ID", "id", workspaceID)

	// Use default image if not specified
	image := req.Image
	if image == "" {
		image = s.config.DefaultImage
	}

	// Create workspace object with initial status
	now := time.Now()
	workspace := &domain.Workspace{
		ID:        workspaceID,
		Name:      req.Name,
		Status:    domain.StatusCreating,
		CreatedAt: now,
		UpdatedAt: now,
		Config: domain.WorkspaceConfig{
			Image:        image,
			Scripts:      req.Scripts,
			ExposedPorts: []int{}, // Initialize as empty array
		},
	}

	// Save workspace to repository with "creating" status
	err := s.repo.Create(workspace)
	if err != nil {
		utils.Error("Failed to save workspace to repository", "error", err)
		return nil, fmt.Errorf("failed to save workspace: %w", err)
	}

	// Create and start container in background
	go func() {
		// Create new context for background operation
		bgCtx := context.Background()

		// Create Docker container
		containerCfg := ContainerConfig{
			Image: image,
			Name:  fmt.Sprintf("vibox-%s", workspaceID),
		}

		containerID, err := s.dockerSvc.CreateContainer(bgCtx, containerCfg)
		if err != nil {
			utils.Error("Failed to create container", "workspaceID", workspaceID, "error", err)
			s.updateWorkspaceStatus(workspaceID, domain.StatusError, fmt.Sprintf("Failed to create container: %v", err))
			return
		}

		// Update workspace with container ID
		workspace.ContainerID = containerID
		workspace.UpdatedAt = time.Now()
		if err := s.repo.Update(workspace); err != nil {
			utils.Error("Failed to update workspace with container ID", "workspaceID", workspaceID, "error", err)
			// Try to clean up the container
			_ = s.dockerSvc.RemoveContainer(bgCtx, containerID)
			s.updateWorkspaceStatus(workspaceID, domain.StatusError, fmt.Sprintf("Failed to update workspace: %v", err))
			return
		}

		// Start container
		err = s.dockerSvc.StartContainer(bgCtx, containerID)
		if err != nil {
			utils.Error("Failed to start container", "workspaceID", workspaceID, "containerID", containerID[:12], "error", err)
			s.updateWorkspaceStatus(workspaceID, domain.StatusError, fmt.Sprintf("Failed to start container: %v", err))
			return
		}

		// Execute initialization scripts if any
		if len(req.Scripts) > 0 {
			utils.Info("Executing initialization scripts", "workspaceID", workspaceID, "scriptCount", len(req.Scripts))
			err = s.executeScripts(bgCtx, containerID, req.Scripts)
			if err != nil {
				utils.Error("Script execution failed", "workspaceID", workspaceID, "error", err)
				s.updateWorkspaceStatus(workspaceID, domain.StatusError, fmt.Sprintf("Script execution failed: %v", err))
				// Keep container running for debugging
				return
			}
		}

		// Update status to running
		utils.Info("Workspace created successfully", "workspaceID", workspaceID)
		s.updateWorkspaceStatus(workspaceID, domain.StatusRunning, "")
	}()

	// Return workspace immediately with "creating" status
	return workspace, nil
}

// GetWorkspace retrieves a workspace by ID
func (s *WorkspaceService) GetWorkspace(id string) (*domain.Workspace, error) {
	utils.Debug("Getting workspace", "id", id)

	workspace, err := s.repo.Get(id)
	if err != nil {
		utils.Error("Failed to get workspace", "id", id, "error", err)
		return nil, fmt.Errorf("workspace not found: %w", err)
	}

	return workspace, nil
}

// ListWorkspaces returns all workspaces
func (s *WorkspaceService) ListWorkspaces() ([]*domain.Workspace, error) {
	utils.Debug("Listing all workspaces")

	workspaces, err := s.repo.List()
	if err != nil {
		utils.Error("Failed to list workspaces", "error", err)
		return nil, fmt.Errorf("failed to list workspaces: %w", err)
	}

	utils.Debug("Listed workspaces", "count", len(workspaces))
	return workspaces, nil
}

// DeleteWorkspace deletes a workspace and its container
func (s *WorkspaceService) DeleteWorkspace(ctx context.Context, id string) error {
	utils.Info("Deleting workspace", "id", id)

	// Get workspace from repository
	workspace, err := s.repo.Get(id)
	if err != nil {
		utils.Error("Failed to get workspace for deletion", "id", id, "error", err)
		return fmt.Errorf("workspace not found: %w", err)
	}

	// Delete container if it exists
	if workspace.ContainerID != "" {
		utils.Info("Deleting container", "workspaceID", id, "containerID", workspace.ContainerID[:12])
		err = s.dockerSvc.RemoveContainer(ctx, workspace.ContainerID)
		if err != nil {
			utils.Warn("Failed to delete container (continuing with workspace deletion)", "containerID", workspace.ContainerID[:12], "error", err)
			// Continue with workspace deletion even if container deletion fails
		}
	}

	// Delete workspace from repository
	err = s.repo.Delete(id)
	if err != nil {
		utils.Error("Failed to delete workspace from repository", "id", id, "error", err)
		return fmt.Errorf("failed to delete workspace: %w", err)
	}

	utils.Info("Workspace deleted successfully", "id", id)
	return nil
}

// sanitizeScriptName removes dangerous characters from script names to prevent path traversal
func sanitizeScriptName(name string) string {
	// Only allow alphanumeric, underscore, and hyphen characters
	return regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(name, "_")
}

// executeScripts executes initialization scripts in order
func (s *WorkspaceService) executeScripts(ctx context.Context, containerID string, scripts []domain.Script) error {
	if len(scripts) == 0 {
		return nil
	}

	// Sort scripts by order
	sortedScripts := make([]domain.Script, len(scripts))
	copy(sortedScripts, scripts)
	sort.Slice(sortedScripts, func(i, j int) bool {
		return sortedScripts[i].Order < sortedScripts[j].Order
	})

	utils.Info("Starting script execution", "containerID", containerID[:12], "scriptCount", len(sortedScripts))

	// Create log directory in container
	logDir := "/var/log/vibox"
	_, err := s.dockerSvc.ExecCommand(ctx, containerID, []string{"mkdir", "-p", logDir})
	if err != nil {
		utils.Warn("Failed to create log directory", "containerID", containerID[:12], "error", err)
		// Continue anyway, scripts might still work
	}

	// Execute each script in order
	for i, script := range sortedScripts {
		utils.Info("Executing script", "containerID", containerID[:12], "scriptName", script.Name, "order", script.Order, "progress", fmt.Sprintf("%d/%d", i+1, len(sortedScripts)))

		// Sanitize script name to prevent path traversal
		safeScriptName := sanitizeScriptName(script.Name)
		if safeScriptName != script.Name {
			utils.Warn("Script name sanitized", "original", script.Name, "sanitized", safeScriptName)
		}

		// Create script file path
		scriptPath := fmt.Sprintf("/tmp/vibox-script-%d-%s.sh", script.Order, safeScriptName)

		// Copy script to container
		err := s.dockerSvc.CopyToContainer(ctx, containerID, scriptPath, []byte(script.Content))
		if err != nil {
			utils.Error("Failed to copy script to container", "scriptName", script.Name, "error", err)
			return fmt.Errorf("failed to copy script %s: %w", script.Name, err)
		}

		// Make script executable
		_, err = s.dockerSvc.ExecCommand(ctx, containerID, []string{"chmod", "+x", scriptPath})
		if err != nil {
			utils.Error("Failed to make script executable", "scriptName", script.Name, "error", err)
			return fmt.Errorf("failed to make script %s executable: %w", script.Name, err)
		}

		// Execute script and redirect output to log file
		logFile := fmt.Sprintf("%s/%s.log", logDir, script.Name)
		cmd := []string{
			"/bin/bash", "-c",
			fmt.Sprintf("%s > %s 2>&1; echo $? > %s.exit", scriptPath, logFile, logFile),
		}
		_, err = s.dockerSvc.ExecCommand(ctx, containerID, cmd)
		if err != nil {
			utils.Error("Failed to execute script", "scriptName", script.Name, "error", err)
			return fmt.Errorf("failed to execute script %s: %w", script.Name, err)
		}

		// Check exit code
		exitCodeOutput, err := s.dockerSvc.ExecCommand(ctx, containerID, []string{"cat", fmt.Sprintf("%s.exit", logFile)})
		if err != nil {
			utils.Error("Failed to read script exit code", "scriptName", script.Name, "error", err)
			return fmt.Errorf("failed to read exit code for script %s: %w", script.Name, err)
		}

		// Parse exit code (trim newline)
		exitCode := exitCodeOutput
		if len(exitCode) > 0 && exitCode[len(exitCode)-1] == '\n' {
			exitCode = exitCode[:len(exitCode)-1]
		}

		if exitCode != "0" {
			// Get script output for error message
			scriptOutput, _ := s.dockerSvc.ExecCommand(ctx, containerID, []string{"cat", logFile})
			utils.Error("Script failed", "scriptName", script.Name, "exitCode", exitCode, "output", scriptOutput)
			return fmt.Errorf("script %s failed with exit code %s. Check logs at %s in container", script.Name, exitCode, logFile)
		}

		utils.Info("Script executed successfully", "scriptName", script.Name, "logFile", logFile)
	}

	utils.Info("All scripts executed successfully", "containerID", containerID[:12])
	return nil
}

// updateWorkspaceStatus updates the status of a workspace
func (s *WorkspaceService) updateWorkspaceStatus(workspaceID string, status domain.WorkspaceStatus, errorMsg string) {
	workspace, err := s.repo.Get(workspaceID)
	if err != nil {
		utils.Error("Failed to get workspace for status update", "workspaceID", workspaceID, "error", err)
		return
	}

	workspace.Status = status
	workspace.Error = errorMsg
	workspace.UpdatedAt = time.Now()

	err = s.repo.Update(workspace)
	if err != nil {
		utils.Error("Failed to update workspace status", "workspaceID", workspaceID, "status", status, "error", err)
	} else {
		utils.Info("Workspace status updated", "workspaceID", workspaceID, "status", status)
	}
}
