package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/1PercentSync/vibox/internal/config"
	"github.com/1PercentSync/vibox/internal/domain"
	"github.com/1PercentSync/vibox/internal/repository"
	"github.com/1PercentSync/vibox/internal/service"
	"github.com/1PercentSync/vibox/pkg/utils"
	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	// Initialize logger for tests
	utils.InitLogger()

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Run tests
	code := m.Run()
	os.Exit(code)
}

func TestWorkspaceHandler_List_EmptyList(t *testing.T) {
	// Setup
	cfg := &config.Config{
		DockerHost:   "unix:///var/run/docker.sock",
		DefaultImage: "ubuntu:22.04",
	}

	repo := repository.NewMemoryRepository()
	dockerSvc, err := service.NewDockerService(cfg)
	if err != nil {
		t.Skip("Docker not available, skipping test")
		return
	}
	defer dockerSvc.Close()

	workspaceSvc := service.NewWorkspaceService(dockerSvc, repo, cfg)
	handler := NewWorkspaceHandler(workspaceSvc)

	// Create test router
	router := gin.New()
	router.GET("/api/workspaces", handler.List)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/workspaces", nil)
	router.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var workspaces []domain.Workspace
	if err := json.Unmarshal(w.Body.Bytes(), &workspaces); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if len(workspaces) != 0 {
		t.Errorf("Expected empty list, got %d workspaces", len(workspaces))
	}
}

func TestWorkspaceHandler_Get_NotFound(t *testing.T) {
	// Setup
	cfg := &config.Config{
		DockerHost:   "unix:///var/run/docker.sock",
		DefaultImage: "ubuntu:22.04",
	}

	repo := repository.NewMemoryRepository()
	dockerSvc, err := service.NewDockerService(cfg)
	if err != nil {
		t.Skip("Docker not available, skipping test")
		return
	}
	defer dockerSvc.Close()

	workspaceSvc := service.NewWorkspaceService(dockerSvc, repo, cfg)
	handler := NewWorkspaceHandler(workspaceSvc)

	// Create test router
	router := gin.New()
	router.GET("/api/workspaces/:id", handler.Get)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/workspaces/non-existent-id", nil)
	router.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["code"] != "NOT_FOUND" {
		t.Errorf("Expected code NOT_FOUND, got %v", response["code"])
	}
}

func TestWorkspaceHandler_Create_InvalidRequest(t *testing.T) {
	// Setup
	cfg := &config.Config{
		DockerHost:   "unix:///var/run/docker.sock",
		DefaultImage: "ubuntu:22.04",
	}

	repo := repository.NewMemoryRepository()
	dockerSvc, err := service.NewDockerService(cfg)
	if err != nil {
		t.Skip("Docker not available, skipping test")
		return
	}
	defer dockerSvc.Close()

	workspaceSvc := service.NewWorkspaceService(dockerSvc, repo, cfg)
	handler := NewWorkspaceHandler(workspaceSvc)

	// Create test router
	router := gin.New()
	router.POST("/api/workspaces", handler.Create)

	// Make request with invalid JSON
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/workspaces", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["code"] != "INVALID_REQUEST" {
		t.Errorf("Expected code INVALID_REQUEST, got %v", response["code"])
	}
}

func TestWorkspaceHandler_Create_MissingName(t *testing.T) {
	// Setup
	cfg := &config.Config{
		DockerHost:   "unix:///var/run/docker.sock",
		DefaultImage: "ubuntu:22.04",
	}

	repo := repository.NewMemoryRepository()
	dockerSvc, err := service.NewDockerService(cfg)
	if err != nil {
		t.Skip("Docker not available, skipping test")
		return
	}
	defer dockerSvc.Close()

	workspaceSvc := service.NewWorkspaceService(dockerSvc, repo, cfg)
	handler := NewWorkspaceHandler(workspaceSvc)

	// Create test router
	router := gin.New()
	router.POST("/api/workspaces", handler.Create)

	// Make request with missing name field
	reqBody := map[string]interface{}{
		"image": "ubuntu:22.04",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/workspaces", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["code"] != "INVALID_REQUEST" {
		t.Errorf("Expected code INVALID_REQUEST, got %v", response["code"])
	}
}

func TestProxyHandler_Forward_InvalidPort(t *testing.T) {
	// Setup
	cfg := &config.Config{
		DockerHost:   "unix:///var/run/docker.sock",
		DefaultImage: "ubuntu:22.04",
	}

	repo := repository.NewMemoryRepository()
	dockerSvc, err := service.NewDockerService(cfg)
	if err != nil {
		t.Skip("Docker not available, skipping test")
		return
	}
	defer dockerSvc.Close()

	workspaceSvc := service.NewWorkspaceService(dockerSvc, repo, cfg)
	proxySvc := service.NewProxyService(dockerSvc)
	handler := NewProxyHandler(proxySvc, workspaceSvc, dockerSvc)

	// Create test router
	router := gin.New()
	router.GET("/forward/:id/:port/*path", handler.Forward)

	// Test cases
	testCases := []struct {
		name string
		url  string
	}{
		{"non-numeric port", "/forward/ws-123/abc/path"},
		{"port too high", "/forward/ws-123/99999/path"},
		{"port zero", "/forward/ws-123/0/path"},
		{"negative port", "/forward/ws-123/-1/path"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tc.url, nil)
			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status 400 for %s, got %d", tc.name, w.Code)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Errorf("Failed to unmarshal response: %v", err)
			}

			if response["code"] != "INVALID_REQUEST" {
				t.Errorf("Expected code INVALID_REQUEST, got %v", response["code"])
			}
		})
	}
}

func TestProxyHandler_Forward_WorkspaceNotFound(t *testing.T) {
	// Setup
	cfg := &config.Config{
		DockerHost:   "unix:///var/run/docker.sock",
		DefaultImage: "ubuntu:22.04",
	}

	repo := repository.NewMemoryRepository()
	dockerSvc, err := service.NewDockerService(cfg)
	if err != nil {
		t.Skip("Docker not available, skipping test")
		return
	}
	defer dockerSvc.Close()

	workspaceSvc := service.NewWorkspaceService(dockerSvc, repo, cfg)
	proxySvc := service.NewProxyService(dockerSvc)
	handler := NewProxyHandler(proxySvc, workspaceSvc, dockerSvc)

	// Create test router
	router := gin.New()
	router.GET("/forward/:id/:port/*path", handler.Forward)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/forward/non-existent-workspace/8080/", nil)
	router.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["code"] != "NOT_FOUND" {
		t.Errorf("Expected code NOT_FOUND, got %v", response["code"])
	}
}

func TestTerminalHandler_Connect_WorkspaceNotFound(t *testing.T) {
	// Setup
	cfg := &config.Config{
		DockerHost:   "unix:///var/run/docker.sock",
		DefaultImage: "ubuntu:22.04",
	}

	repo := repository.NewMemoryRepository()
	dockerSvc, err := service.NewDockerService(cfg)
	if err != nil {
		t.Skip("Docker not available, skipping test")
		return
	}
	defer dockerSvc.Close()

	workspaceSvc := service.NewWorkspaceService(dockerSvc, repo, cfg)
	terminalSvc := service.NewTerminalService(dockerSvc)
	handler := NewTerminalHandler(terminalSvc, workspaceSvc, dockerSvc)

	// Create test router
	router := gin.New()
	router.GET("/ws/terminal/:id", handler.Connect)

	// Make request (not a WebSocket request, but we're testing the workspace check)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ws/terminal/non-existent-workspace", nil)
	router.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["code"] != "NOT_FOUND" {
		t.Errorf("Expected code NOT_FOUND, got %v", response["code"])
	}
}

func TestWorkspaceHandler_FullCRUD(t *testing.T) {
	// Setup
	cfg := &config.Config{
		DockerHost:   "unix:///var/run/docker.sock",
		DefaultImage: "alpine:latest",
		MemoryLimit:  64 * 1024 * 1024, // 64MB for quick test
		CPULimit:     500000000,         // 0.5 CPU
	}

	repo := repository.NewMemoryRepository()
	dockerSvc, err := service.NewDockerService(cfg)
	if err != nil {
		t.Skip("Docker not available, skipping test")
		return
	}
	defer dockerSvc.Close()

	workspaceSvc := service.NewWorkspaceService(dockerSvc, repo, cfg)
	handler := NewWorkspaceHandler(workspaceSvc)

	// Create test router
	router := gin.New()
	router.POST("/api/workspaces", handler.Create)
	router.GET("/api/workspaces", handler.List)
	router.GET("/api/workspaces/:id", handler.Get)
	router.DELETE("/api/workspaces/:id", handler.Delete)

	// 1. Create workspace
	reqBody := map[string]interface{}{
		"name":  "test-workspace",
		"image": "alpine:latest",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/workspaces", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d", w.Code)
	}

	var createdWorkspace domain.Workspace
	if err := json.Unmarshal(w.Body.Bytes(), &createdWorkspace); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if createdWorkspace.Name != "test-workspace" {
		t.Errorf("Expected name 'test-workspace', got %s", createdWorkspace.Name)
	}

	workspaceID := createdWorkspace.ID

	// Cleanup at the end
	defer func() {
		ctx := context.Background()
		workspaceSvc.DeleteWorkspace(ctx, workspaceID)
	}()

	// 2. Get workspace
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/workspaces/"+workspaceID, nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// 3. List workspaces
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/workspaces", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var workspaces []domain.Workspace
	if err := json.Unmarshal(w.Body.Bytes(), &workspaces); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if len(workspaces) != 1 {
		t.Errorf("Expected 1 workspace, got %d", len(workspaces))
	}

	// 4. Delete workspace
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/workspaces/"+workspaceID, nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify deletion
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/workspaces/"+workspaceID, nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 after deletion, got %d", w.Code)
	}
}
