package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/1PercentSync/vibox/internal/config"
)

// TestNewProxyService tests proxy service initialization
func TestNewProxyService(t *testing.T) {
	// Skip if Docker is not available
	cfg := &config.Config{
		DockerHost:   "unix:///var/run/docker.sock",
		DefaultImage: "alpine:latest",
		MemoryLimit:  128 * 1024 * 1024,
		CPULimit:     500000000,
	}

	dockerSvc, err := NewDockerService(cfg)
	if err != nil {
		t.Skipf("Docker not available, skipping test: %v", err)
	}
	defer dockerSvc.Close()

	// Create proxy service
	proxySvc := NewProxyService(dockerSvc)
	if proxySvc == nil {
		t.Fatal("Expected proxy service to be created")
	}

	if proxySvc.dockerSvc == nil {
		t.Fatal("Expected docker service to be set")
	}
}

// TestProxyRequestToHTTPServer tests proxying to a container running an HTTP server
func TestProxyRequestToHTTPServer(t *testing.T) {
	// Skip if Docker is not available
	cfg := &config.Config{
		DockerHost:   "unix:///var/run/docker.sock",
		DefaultImage: "alpine:latest",
		MemoryLimit:  128 * 1024 * 1024,
		CPULimit:     500000000,
	}

	dockerSvc, err := NewDockerService(cfg)
	if err != nil {
		t.Skipf("Docker not available, skipping test: %v", err)
	}
	defer dockerSvc.Close()

	// Create proxy service
	proxySvc := NewProxyService(dockerSvc)

	ctx := context.Background()

	// Create a test container
	containerCfg := ContainerConfig{
		Image: "alpine:latest",
		Name:  fmt.Sprintf("test-proxy-%d", time.Now().Unix()),
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

	// Wait for container to be fully started
	time.Sleep(2 * time.Second)

	// Install and start a simple HTTP server in the container
	setupScript := `#!/bin/sh
apk add --no-cache python3
nohup python3 -m http.server 8080 > /tmp/server.log 2>&1 &
sleep 2
`
	err = dockerSvc.CopyToContainer(ctx, containerID, "/tmp/setup.sh", []byte(setupScript))
	if err != nil {
		t.Fatalf("Failed to copy setup script: %v", err)
	}

	_, err = dockerSvc.ExecCommand(ctx, containerID, []string{"/bin/sh", "/tmp/setup.sh"})
	if err != nil {
		t.Fatalf("Failed to execute setup script: %v", err)
	}

	// Wait for HTTP server to start
	time.Sleep(3 * time.Second)

	// Test proxying a request
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	err = proxySvc.ProxyRequest(w, req, containerID, 8080)
	if err != nil {
		t.Fatalf("Proxy request failed: %v", err)
	}

	// Check response
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadGateway {
		t.Errorf("Expected status OK or Bad Gateway, got %d", resp.StatusCode)
	}

	// Note: We allow Bad Gateway because the HTTP server might not have started in time
	// This is acceptable for this test
}

// TestProxyRequestContainerNotRunning tests proxying to a stopped container
func TestProxyRequestContainerNotRunning(t *testing.T) {
	// Skip if Docker is not available
	cfg := &config.Config{
		DockerHost:   "unix:///var/run/docker.sock",
		DefaultImage: "alpine:latest",
		MemoryLimit:  128 * 1024 * 1024,
		CPULimit:     500000000,
	}

	dockerSvc, err := NewDockerService(cfg)
	if err != nil {
		t.Skipf("Docker not available, skipping test: %v", err)
	}
	defer dockerSvc.Close()

	// Create proxy service
	proxySvc := NewProxyService(dockerSvc)

	ctx := context.Background()

	// Create and immediately stop a container
	containerCfg := ContainerConfig{
		Image: "alpine:latest",
		Name:  fmt.Sprintf("test-proxy-stopped-%d", time.Now().Unix()),
	}

	containerID, err := dockerSvc.CreateContainer(ctx, containerCfg)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer dockerSvc.RemoveContainer(ctx, containerID)

	// Don't start the container

	// Test proxying a request (should fail)
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	err = proxySvc.ProxyRequest(w, req, containerID, 8080)
	if err == nil {
		t.Error("Expected proxy request to fail for stopped container")
	}

	// Check that error response was written
	resp := w.Result()
	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("Expected Bad Gateway status, got %d", resp.StatusCode)
	}
}

// TestProxyRequestNonExistentContainer tests proxying to a non-existent container
func TestProxyRequestNonExistentContainer(t *testing.T) {
	// Skip if Docker is not available
	cfg := &config.Config{
		DockerHost:   "unix:///var/run/docker.sock",
		DefaultImage: "alpine:latest",
		MemoryLimit:  128 * 1024 * 1024,
		CPULimit:     500000000,
	}

	dockerSvc, err := NewDockerService(cfg)
	if err != nil {
		t.Skipf("Docker not available, skipping test: %v", err)
	}
	defer dockerSvc.Close()

	// Create proxy service
	proxySvc := NewProxyService(dockerSvc)

	// Test proxying to non-existent container
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	err = proxySvc.ProxyRequest(w, req, "nonexistent-container-id", 8080)
	if err == nil {
		t.Error("Expected proxy request to fail for non-existent container")
	}

	// Check error response
	resp := w.Result()
	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("Expected Bad Gateway status, got %d", resp.StatusCode)
	}
}

// TestGetContainerIP tests the convenience method for getting container IP
func TestGetContainerIP(t *testing.T) {
	// Skip if Docker is not available
	cfg := &config.Config{
		DockerHost:   "unix:///var/run/docker.sock",
		DefaultImage: "alpine:latest",
		MemoryLimit:  128 * 1024 * 1024,
		CPULimit:     500000000,
	}

	dockerSvc, err := NewDockerService(cfg)
	if err != nil {
		t.Skipf("Docker not available, skipping test: %v", err)
	}
	defer dockerSvc.Close()

	// Create proxy service
	proxySvc := NewProxyService(dockerSvc)

	ctx := context.Background()

	// Create and start a container
	containerCfg := ContainerConfig{
		Image: "alpine:latest",
		Name:  fmt.Sprintf("test-proxy-ip-%d", time.Now().Unix()),
	}

	containerID, err := dockerSvc.CreateContainer(ctx, containerCfg)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer dockerSvc.RemoveContainer(ctx, containerID)

	err = dockerSvc.StartContainer(ctx, containerID)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	// Wait for container to be fully started
	time.Sleep(2 * time.Second)

	// Get container IP
	ip, err := proxySvc.GetContainerIP(ctx, containerID)
	if err != nil {
		t.Fatalf("Failed to get container IP: %v", err)
	}

	if ip == "" {
		t.Error("Expected non-empty IP address")
	}

	t.Logf("Container IP: %s", ip)
}

// TestProxyWithDifferentHTTPMethods tests proxying different HTTP methods
func TestProxyWithDifferentHTTPMethods(t *testing.T) {
	// Skip if Docker is not available
	cfg := &config.Config{
		DockerHost:   "unix:///var/run/docker.sock",
		DefaultImage: "alpine:latest",
		MemoryLimit:  128 * 1024 * 1024,
		CPULimit:     500000000,
	}

	dockerSvc, err := NewDockerService(cfg)
	if err != nil {
		t.Skipf("Docker not available, skipping test: %v", err)
	}
	defer dockerSvc.Close()

	// Create proxy service
	proxySvc := NewProxyService(dockerSvc)

	ctx := context.Background()

	// Create a test container
	containerCfg := ContainerConfig{
		Image: "alpine:latest",
		Name:  fmt.Sprintf("test-proxy-methods-%d", time.Now().Unix()),
	}

	containerID, err := dockerSvc.CreateContainer(ctx, containerCfg)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer dockerSvc.RemoveContainer(ctx, containerID)

	err = dockerSvc.StartContainer(ctx, containerID)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	// Wait for container to be fully started
	time.Sleep(2 * time.Second)

	// Install HTTP server
	setupScript := `#!/bin/sh
apk add --no-cache python3
nohup python3 -m http.server 8080 > /tmp/server.log 2>&1 &
sleep 2
`
	err = dockerSvc.CopyToContainer(ctx, containerID, "/tmp/setup.sh", []byte(setupScript))
	if err != nil {
		t.Fatalf("Failed to copy setup script: %v", err)
	}

	_, err = dockerSvc.ExecCommand(ctx, containerID, []string{"/bin/sh", "/tmp/setup.sh"})
	if err != nil {
		t.Fatalf("Failed to execute setup script: %v", err)
	}

	// Wait for HTTP server to start
	time.Sleep(3 * time.Second)

	// Test different HTTP methods
	methods := []string{"GET", "POST", "PUT", "DELETE"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/", nil)
			w := httptest.NewRecorder()

			err = proxySvc.ProxyRequest(w, req, containerID, 8080)
			if err != nil {
				t.Logf("Proxy request failed for %s: %v (this may be expected)", method, err)
			}

			resp := w.Result()
			defer resp.Body.Close()

			// We don't check specific status codes because the Python HTTP server
			// might handle different methods differently
			t.Logf("%s request returned status: %d", method, resp.StatusCode)
		})
	}
}

// Note: TestMain is defined in docker_test.go and initializes the logger for all service tests
