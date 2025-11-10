package repository

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/1PercentSync/vibox/internal/domain"
	"github.com/1PercentSync/vibox/pkg/utils"
)

func TestMain(m *testing.M) {
	// Initialize logger for tests
	utils.InitLogger()
	code := m.Run()
	os.Exit(code)
}

func TestNewMemoryRepository(t *testing.T) {
	repo := NewMemoryRepository()
	if repo == nil {
		t.Fatal("Expected repository to be created, got nil")
	}
	if repo.store == nil {
		t.Fatal("Expected store to be initialized, got nil")
	}
}

func TestCreate(t *testing.T) {
	repo := NewMemoryRepository()

	// Test successful creation
	ws := &domain.Workspace{
		ID:          "ws-test-001",
		Name:        "test-workspace",
		ContainerID: "container-123",
		Status:      domain.StatusCreating,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Config: domain.WorkspaceConfig{
			Image: "ubuntu:22.04",
		},
	}

	err := repo.Create(ws)
	if err != nil {
		t.Fatalf("Expected successful creation, got error: %v", err)
	}

	// Test duplicate creation
	err = repo.Create(ws)
	if err == nil {
		t.Fatal("Expected error when creating duplicate workspace, got nil")
	}

	// Test nil workspace
	err = repo.Create(nil)
	if err == nil {
		t.Fatal("Expected error when creating nil workspace, got nil")
	}

	// Test empty ID
	wsEmptyID := &domain.Workspace{
		ID:   "",
		Name: "empty-id",
	}
	err = repo.Create(wsEmptyID)
	if err == nil {
		t.Fatal("Expected error when creating workspace with empty ID, got nil")
	}
}

func TestGet(t *testing.T) {
	repo := NewMemoryRepository()

	// Create a test workspace
	ws := &domain.Workspace{
		ID:          "ws-test-002",
		Name:        "test-workspace-2",
		ContainerID: "container-456",
		Status:      domain.StatusRunning,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	repo.Create(ws)

	// Test successful get
	retrieved, err := repo.Get("ws-test-002")
	if err != nil {
		t.Fatalf("Expected successful get, got error: %v", err)
	}
	if retrieved.ID != ws.ID {
		t.Errorf("Expected ID %s, got %s", ws.ID, retrieved.ID)
	}
	if retrieved.Name != ws.Name {
		t.Errorf("Expected Name %s, got %s", ws.Name, retrieved.Name)
	}

	// Test get non-existent workspace
	_, err = repo.Get("ws-non-existent")
	if err == nil {
		t.Fatal("Expected error when getting non-existent workspace, got nil")
	}

	// Test get with empty ID
	_, err = repo.Get("")
	if err == nil {
		t.Fatal("Expected error when getting workspace with empty ID, got nil")
	}
}

func TestList(t *testing.T) {
	repo := NewMemoryRepository()

	// Test empty list
	workspaces, err := repo.List()
	if err != nil {
		t.Fatalf("Expected successful list, got error: %v", err)
	}
	if len(workspaces) != 0 {
		t.Errorf("Expected 0 workspaces, got %d", len(workspaces))
	}

	// Add multiple workspaces
	ws1 := &domain.Workspace{
		ID:   "ws-test-003",
		Name: "workspace-1",
	}
	ws2 := &domain.Workspace{
		ID:   "ws-test-004",
		Name: "workspace-2",
	}
	ws3 := &domain.Workspace{
		ID:   "ws-test-005",
		Name: "workspace-3",
	}

	repo.Create(ws1)
	repo.Create(ws2)
	repo.Create(ws3)

	// Test list with multiple workspaces
	workspaces, err = repo.List()
	if err != nil {
		t.Fatalf("Expected successful list, got error: %v", err)
	}
	if len(workspaces) != 3 {
		t.Errorf("Expected 3 workspaces, got %d", len(workspaces))
	}
}

func TestUpdate(t *testing.T) {
	repo := NewMemoryRepository()

	// Create a test workspace
	ws := &domain.Workspace{
		ID:     "ws-test-006",
		Name:   "original-name",
		Status: domain.StatusCreating,
	}
	repo.Create(ws)

	// Test successful update
	ws.Name = "updated-name"
	ws.Status = domain.StatusRunning
	err := repo.Update(ws)
	if err != nil {
		t.Fatalf("Expected successful update, got error: %v", err)
	}

	// Verify update
	retrieved, _ := repo.Get("ws-test-006")
	if retrieved.Name != "updated-name" {
		t.Errorf("Expected Name 'updated-name', got %s", retrieved.Name)
	}
	if retrieved.Status != domain.StatusRunning {
		t.Errorf("Expected Status %s, got %s", domain.StatusRunning, retrieved.Status)
	}

	// Test update non-existent workspace
	wsNonExistent := &domain.Workspace{
		ID:   "ws-non-existent",
		Name: "non-existent",
	}
	err = repo.Update(wsNonExistent)
	if err == nil {
		t.Fatal("Expected error when updating non-existent workspace, got nil")
	}

	// Test update nil workspace
	err = repo.Update(nil)
	if err == nil {
		t.Fatal("Expected error when updating nil workspace, got nil")
	}

	// Test update with empty ID
	wsEmptyID := &domain.Workspace{
		ID:   "",
		Name: "empty-id",
	}
	err = repo.Update(wsEmptyID)
	if err == nil {
		t.Fatal("Expected error when updating workspace with empty ID, got nil")
	}
}

func TestDelete(t *testing.T) {
	repo := NewMemoryRepository()

	// Create a test workspace
	ws := &domain.Workspace{
		ID:   "ws-test-007",
		Name: "to-be-deleted",
	}
	repo.Create(ws)

	// Verify it exists
	_, err := repo.Get("ws-test-007")
	if err != nil {
		t.Fatalf("Expected workspace to exist, got error: %v", err)
	}

	// Test successful delete
	err = repo.Delete("ws-test-007")
	if err != nil {
		t.Fatalf("Expected successful delete, got error: %v", err)
	}

	// Verify it's deleted
	_, err = repo.Get("ws-test-007")
	if err == nil {
		t.Fatal("Expected workspace to be deleted, but it still exists")
	}

	// Test delete non-existent workspace
	err = repo.Delete("ws-non-existent")
	if err == nil {
		t.Fatal("Expected error when deleting non-existent workspace, got nil")
	}

	// Test delete with empty ID
	err = repo.Delete("")
	if err == nil {
		t.Fatal("Expected error when deleting workspace with empty ID, got nil")
	}
}

func TestConcurrentAccess(t *testing.T) {
	repo := NewMemoryRepository()

	// Test concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			ws := &domain.Workspace{
				ID:   fmt.Sprintf("ws-concurrent-%d", idx),
				Name: fmt.Sprintf("workspace-%d", idx),
			}
			repo.Create(ws)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all workspaces were created
	workspaces, err := repo.List()
	if err != nil {
		t.Fatalf("Expected successful list, got error: %v", err)
	}
	if len(workspaces) != 10 {
		t.Errorf("Expected 10 workspaces, got %d", len(workspaces))
	}

	// Test concurrent reads
	for i := 0; i < 10; i++ {
		go func(idx int) {
			_, _ = repo.Get(fmt.Sprintf("ws-concurrent-%d", idx))
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
