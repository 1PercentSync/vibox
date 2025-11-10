package service

import (
	"context"
	"testing"
	"time"

	"github.com/1PercentSync/vibox/internal/config"
	"github.com/1PercentSync/vibox/internal/domain"
	"github.com/1PercentSync/vibox/internal/repository"
	"github.com/1PercentSync/vibox/pkg/utils"
)

func TestNewWorkspaceService(t *testing.T) {
	cfg := &config.Config{
		DockerHost:   "unix:///var/run/docker.sock",
		DefaultImage: "alpine:latest",
	}

	dockerSvc, err := NewDockerService(cfg)
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer dockerSvc.Close()

	repo := repository.NewMemoryRepository()
	workspaceSvc := NewWorkspaceService(dockerSvc, repo, cfg)

	if workspaceSvc == nil {
		t.Fatal("Expected workspace service to be created")
	}
	if workspaceSvc.dockerSvc != dockerSvc {
		t.Error("Expected docker service to be set")
	}
	if workspaceSvc.repo != repo {
		t.Error("Expected repository to be set")
	}
	if workspaceSvc.config != cfg {
		t.Error("Expected config to be set")
	}
}

func TestCreateWorkspace(t *testing.T) {
	cfg := &config.Config{
		DockerHost:   "unix:///var/run/docker.sock",
		DefaultImage: "alpine:latest",
	}

	dockerSvc, err := NewDockerService(cfg)
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer dockerSvc.Close()

	repo := repository.NewMemoryRepository()
	workspaceSvc := NewWorkspaceService(dockerSvc, repo, cfg)

	ctx := context.Background()
	req := CreateWorkspaceRequest{
		Name:  "test-workspace",
		Image: "alpine:latest",
	}

	// Create workspace
	workspace, err := workspaceSvc.CreateWorkspace(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}

	// Verify workspace was created
	if workspace.ID == "" {
		t.Error("Expected workspace ID to be generated")
	}
	if workspace.Name != req.Name {
		t.Errorf("Expected name %s, got %s", req.Name, workspace.Name)
	}
	if workspace.Status != domain.StatusCreating {
		t.Errorf("Expected status %s, got %s", domain.StatusCreating, workspace.Status)
	}
	if workspace.Config.Image != "alpine:latest" {
		t.Errorf("Expected image alpine:latest, got %s", workspace.Config.Image)
	}

	// Wait a bit for background goroutine to complete
	time.Sleep(3 * time.Second)

	// Verify workspace was saved to repository
	savedWorkspace, err := repo.Get(workspace.ID)
	if err != nil {
		t.Fatalf("Failed to get workspace from repository: %v", err)
	}

	// Check if status changed from creating
	if savedWorkspace.Status == domain.StatusCreating {
		t.Log("Workspace still in creating status (background operation may still be running)")
	}

	// Cleanup
	if savedWorkspace.ContainerID != "" {
		_ = dockerSvc.RemoveContainer(ctx, savedWorkspace.ContainerID)
	}
	_ = repo.Delete(workspace.ID)
}

func TestCreateWorkspaceWithScripts(t *testing.T) {
	cfg := &config.Config{
		DockerHost:   "unix:///var/run/docker.sock",
		DefaultImage: "alpine:latest",
	}

	dockerSvc, err := NewDockerService(cfg)
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer dockerSvc.Close()

	repo := repository.NewMemoryRepository()
	workspaceSvc := NewWorkspaceService(dockerSvc, repo, cfg)

	ctx := context.Background()
	req := CreateWorkspaceRequest{
		Name:  "test-workspace-with-scripts",
		Image: "alpine:latest",
		Scripts: []domain.Script{
			{
				Name:    "test-script",
				Content: "#!/bin/sh\necho 'Hello from script'\n",
				Order:   1,
			},
		},
	}

	// Create workspace
	workspace, err := workspaceSvc.CreateWorkspace(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}

	// Wait for background operation to complete
	time.Sleep(5 * time.Second)

	// Get workspace to check final status
	savedWorkspace, err := repo.Get(workspace.ID)
	if err != nil {
		t.Fatalf("Failed to get workspace: %v", err)
	}

	// Verify scripts were configured
	if len(savedWorkspace.Config.Scripts) != 1 {
		t.Errorf("Expected 1 script, got %d", len(savedWorkspace.Config.Scripts))
	}

	// Check if workspace reached running status or has error
	t.Logf("Final workspace status: %s", savedWorkspace.Status)
	if savedWorkspace.Status == domain.StatusError {
		t.Logf("Workspace error: %s", savedWorkspace.Error)
	}

	// Cleanup
	if savedWorkspace.ContainerID != "" {
		_ = dockerSvc.RemoveContainer(ctx, savedWorkspace.ContainerID)
	}
	_ = repo.Delete(workspace.ID)
}

func TestCreateWorkspaceWithFailingScript(t *testing.T) {
	cfg := &config.Config{
		DockerHost:   "unix:///var/run/docker.sock",
		DefaultImage: "alpine:latest",
	}

	dockerSvc, err := NewDockerService(cfg)
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer dockerSvc.Close()

	repo := repository.NewMemoryRepository()
	workspaceSvc := NewWorkspaceService(dockerSvc, repo, cfg)

	ctx := context.Background()
	req := CreateWorkspaceRequest{
		Name:  "test-workspace-failing",
		Image: "alpine:latest",
		Scripts: []domain.Script{
			{
				Name:    "failing-script",
				Content: "#!/bin/sh\nexit 1\n",
				Order:   1,
			},
		},
	}

	// Create workspace
	workspace, err := workspaceSvc.CreateWorkspace(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}

	// Wait for background operation to complete
	time.Sleep(5 * time.Second)

	// Get workspace to check final status
	savedWorkspace, err := repo.Get(workspace.ID)
	if err != nil {
		t.Fatalf("Failed to get workspace: %v", err)
	}

	// Verify workspace is in error status
	if savedWorkspace.Status != domain.StatusError {
		t.Errorf("Expected status %s, got %s", domain.StatusError, savedWorkspace.Status)
	}

	// Verify error message is set
	if savedWorkspace.Error == "" {
		t.Error("Expected error message to be set")
	} else {
		t.Logf("Error message: %s", savedWorkspace.Error)
	}

	// Cleanup
	if savedWorkspace.ContainerID != "" {
		_ = dockerSvc.RemoveContainer(ctx, savedWorkspace.ContainerID)
	}
	_ = repo.Delete(workspace.ID)
}

func TestGetWorkspace(t *testing.T) {
	cfg := &config.Config{
		DockerHost:   "unix:///var/run/docker.sock",
		DefaultImage: "alpine:latest",
	}

	dockerSvc, err := NewDockerService(cfg)
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer dockerSvc.Close()

	repo := repository.NewMemoryRepository()
	workspaceSvc := NewWorkspaceService(dockerSvc, repo, cfg)

	// Create a test workspace directly in repository
	testWorkspace := &domain.Workspace{
		ID:        "ws-test123",
		Name:      "test-get",
		Status:    domain.StatusRunning,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Config: domain.WorkspaceConfig{
			Image: "alpine:latest",
		},
	}
	err = repo.Create(testWorkspace)
	if err != nil {
		t.Fatalf("Failed to create test workspace: %v", err)
	}

	// Get workspace
	workspace, err := workspaceSvc.GetWorkspace(testWorkspace.ID)
	if err != nil {
		t.Fatalf("Failed to get workspace: %v", err)
	}

	// Verify workspace data
	if workspace.ID != testWorkspace.ID {
		t.Errorf("Expected ID %s, got %s", testWorkspace.ID, workspace.ID)
	}
	if workspace.Name != testWorkspace.Name {
		t.Errorf("Expected name %s, got %s", testWorkspace.Name, workspace.Name)
	}

	// Test non-existent workspace
	_, err = workspaceSvc.GetWorkspace("ws-nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent workspace")
	}

	// Cleanup
	_ = repo.Delete(testWorkspace.ID)
}

func TestListWorkspaces(t *testing.T) {
	cfg := &config.Config{
		DockerHost:   "unix:///var/run/docker.sock",
		DefaultImage: "alpine:latest",
	}

	dockerSvc, err := NewDockerService(cfg)
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer dockerSvc.Close()

	repo := repository.NewMemoryRepository()
	workspaceSvc := NewWorkspaceService(dockerSvc, repo, cfg)

	// List should be empty initially
	workspaces, err := workspaceSvc.ListWorkspaces()
	if err != nil {
		t.Fatalf("Failed to list workspaces: %v", err)
	}
	if len(workspaces) != 0 {
		t.Errorf("Expected 0 workspaces, got %d", len(workspaces))
	}

	// Create test workspaces
	for i := 0; i < 3; i++ {
		ws := &domain.Workspace{
			ID:        utils.GenerateID(),
			Name:      "test-list",
			Status:    domain.StatusRunning,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Config: domain.WorkspaceConfig{
				Image: "alpine:latest",
			},
		}
		_ = repo.Create(ws)
	}

	// List workspaces
	workspaces, err = workspaceSvc.ListWorkspaces()
	if err != nil {
		t.Fatalf("Failed to list workspaces: %v", err)
	}
	if len(workspaces) != 3 {
		t.Errorf("Expected 3 workspaces, got %d", len(workspaces))
	}

	// Cleanup
	for _, ws := range workspaces {
		_ = repo.Delete(ws.ID)
	}
}

func TestDeleteWorkspace(t *testing.T) {
	cfg := &config.Config{
		DockerHost:   "unix:///var/run/docker.sock",
		DefaultImage: "alpine:latest",
	}

	dockerSvc, err := NewDockerService(cfg)
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer dockerSvc.Close()

	repo := repository.NewMemoryRepository()
	workspaceSvc := NewWorkspaceService(dockerSvc, repo, cfg)

	ctx := context.Background()

	// Create a real container
	containerCfg := ContainerConfig{
		Image: "alpine:latest",
		Name:  "test-delete-workspace",
	}
	containerID, err := dockerSvc.CreateContainer(ctx, containerCfg)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}

	// Start container
	err = dockerSvc.StartContainer(ctx, containerID)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	// Create workspace with this container
	workspace := &domain.Workspace{
		ID:          "ws-delete",
		Name:        "test-delete",
		ContainerID: containerID,
		Status:      domain.StatusRunning,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Config: domain.WorkspaceConfig{
			Image: "alpine:latest",
		},
	}
	err = repo.Create(workspace)
	if err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}

	// Delete workspace
	err = workspaceSvc.DeleteWorkspace(ctx, workspace.ID)
	if err != nil {
		t.Fatalf("Failed to delete workspace: %v", err)
	}

	// Verify workspace was deleted from repository
	_, err = repo.Get(workspace.ID)
	if err == nil {
		t.Error("Expected workspace to be deleted from repository")
	}

	// Verify container was deleted
	_, err = dockerSvc.GetContainerStatus(ctx, containerID)
	if err == nil {
		t.Error("Expected container to be deleted")
	}

	// Test deleting non-existent workspace
	err = workspaceSvc.DeleteWorkspace(ctx, "ws-nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent workspace")
	}
}

func TestScriptOrdering(t *testing.T) {
	cfg := &config.Config{
		DockerHost:   "unix:///var/run/docker.sock",
		DefaultImage: "alpine:latest",
	}

	dockerSvc, err := NewDockerService(cfg)
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer dockerSvc.Close()

	repo := repository.NewMemoryRepository()
	workspaceSvc := NewWorkspaceService(dockerSvc, repo, cfg)

	ctx := context.Background()

	// Create container
	containerCfg := ContainerConfig{
		Image: "alpine:latest",
		Name:  "test-script-ordering",
	}
	containerID, err := dockerSvc.CreateContainer(ctx, containerCfg)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer dockerSvc.RemoveContainer(ctx, containerID)

	// Start container
	err = dockerSvc.StartContainer(ctx, containerID)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	// Scripts with different orders
	scripts := []domain.Script{
		{
			Name:    "third",
			Content: "#!/bin/sh\necho 'third' > /tmp/order.txt\n",
			Order:   3,
		},
		{
			Name:    "first",
			Content: "#!/bin/sh\necho 'first' > /tmp/order.txt\n",
			Order:   1,
		},
		{
			Name:    "second",
			Content: "#!/bin/sh\necho 'second' >> /tmp/order.txt\n",
			Order:   2,
		},
	}

	// Execute scripts
	err = workspaceSvc.executeScripts(ctx, containerID, scripts)
	if err != nil {
		t.Fatalf("Failed to execute scripts: %v", err)
	}

	// Check order by reading the file
	output, err := dockerSvc.ExecCommand(ctx, containerID, []string{"cat", "/tmp/order.txt"})
	if err != nil {
		t.Fatalf("Failed to read order file: %v", err)
	}

	expected := "first\nsecond\nthird\n"
	if output != expected {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expected, output)
	}
}
