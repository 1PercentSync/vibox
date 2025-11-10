package service

import (
	"context"
	"os"
	"testing"

	"github.com/1PercentSync/vibox/internal/config"
	"github.com/1PercentSync/vibox/pkg/utils"
)

// TestMain sets up test environment
func TestMain(m *testing.M) {
	// Initialize logger for tests
	utils.InitLogger()

	// Run tests
	code := m.Run()

	os.Exit(code)
}

// TestNewDockerService tests Docker service initialization
func TestNewDockerService(t *testing.T) {
	cfg := &config.Config{
		DockerHost:   "unix:///var/run/docker.sock",
		DefaultImage: "ubuntu:22.04",
		MemoryLimit:  512 * 1024 * 1024,
		CPULimit:     1000000000,
	}

	svc, err := NewDockerService(cfg)
	if err != nil {
		t.Skipf("Skipping test: Docker is not available: %v", err)
	}
	defer svc.Close()

	if svc.client == nil {
		t.Error("Expected Docker client to be initialized")
	}
	if svc.config != cfg {
		t.Error("Expected config to be set")
	}
}

// TestContainerLifecycle tests container creation, start, stop, and removal
func TestContainerLifecycle(t *testing.T) {
	cfg := &config.Config{
		DockerHost:   "unix:///var/run/docker.sock",
		DefaultImage: "alpine:latest",
		MemoryLimit:  128 * 1024 * 1024,
		CPULimit:     500000000,
	}

	svc, err := NewDockerService(cfg)
	if err != nil {
		t.Skipf("Skipping test: Docker is not available: %v", err)
	}
	defer svc.Close()

	ctx := context.Background()

	// Test container creation
	containerCfg := ContainerConfig{
		Image:       "alpine:latest",
		Name:        "vibox-test-container",
		MemoryLimit: 128 * 1024 * 1024,
		CPULimit:    500000000,
	}

	containerID, err := svc.CreateContainer(ctx, containerCfg)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	if containerID == "" {
		t.Error("Expected container ID to be returned")
	}

	// Ensure cleanup
	defer func() {
		_ = svc.RemoveContainer(ctx, containerID)
	}()

	// Test container start
	err = svc.StartContainer(ctx, containerID)
	if err != nil {
		t.Errorf("Failed to start container: %v", err)
	}

	// Test getting container status
	status, err := svc.GetContainerStatus(ctx, containerID)
	if err != nil {
		t.Errorf("Failed to get container status: %v", err)
	}
	if status != "running" {
		t.Errorf("Expected container status to be 'running', got '%s'", status)
	}

	// Test getting container IP
	ip, err := svc.GetContainerIP(ctx, containerID)
	if err != nil {
		t.Logf("Warning: Failed to get container IP: %v", err)
	} else if ip == "" {
		t.Log("Warning: Container IP is empty")
	}

	// Test container inspection
	inspect, err := svc.InspectContainer(ctx, containerID)
	if err != nil {
		t.Errorf("Failed to inspect container: %v", err)
	}
	if inspect == nil {
		t.Error("Expected inspect result to be non-nil")
	}

	// Test container stop
	err = svc.StopContainer(ctx, containerID, 10)
	if err != nil {
		t.Errorf("Failed to stop container: %v", err)
	}

	// Verify container is stopped
	status, err = svc.GetContainerStatus(ctx, containerID)
	if err != nil {
		t.Errorf("Failed to get container status after stop: %v", err)
	}
	if status == "running" {
		t.Error("Expected container to be stopped")
	}

	// Test container removal
	err = svc.RemoveContainer(ctx, containerID)
	if err != nil {
		t.Errorf("Failed to remove container: %v", err)
	}
}

// TestExecCommand tests command execution in a container
func TestExecCommand(t *testing.T) {
	cfg := &config.Config{
		DockerHost:   "unix:///var/run/docker.sock",
		DefaultImage: "alpine:latest",
		MemoryLimit:  128 * 1024 * 1024,
		CPULimit:     500000000,
	}

	svc, err := NewDockerService(cfg)
	if err != nil {
		t.Skipf("Skipping test: Docker is not available: %v", err)
	}
	defer svc.Close()

	ctx := context.Background()

	// Create and start a container
	containerCfg := ContainerConfig{
		Image: "alpine:latest",
		Name:  "vibox-test-exec",
	}

	containerID, err := svc.CreateContainer(ctx, containerCfg)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer svc.RemoveContainer(ctx, containerID)

	err = svc.StartContainer(ctx, containerID)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	// Test command execution
	output, err := svc.ExecCommand(ctx, containerID, []string{"echo", "hello world"})
	if err != nil {
		t.Errorf("Failed to execute command: %v", err)
	}
	if output == "" {
		t.Error("Expected command output to be non-empty")
	}
}

// TestCopyToContainer tests copying files to a container
func TestCopyToContainer(t *testing.T) {
	cfg := &config.Config{
		DockerHost:   "unix:///var/run/docker.sock",
		DefaultImage: "alpine:latest",
		MemoryLimit:  128 * 1024 * 1024,
		CPULimit:     500000000,
	}

	svc, err := NewDockerService(cfg)
	if err != nil {
		t.Skipf("Skipping test: Docker is not available: %v", err)
	}
	defer svc.Close()

	ctx := context.Background()

	// Create and start a container
	containerCfg := ContainerConfig{
		Image: "alpine:latest",
		Name:  "vibox-test-copy",
	}

	containerID, err := svc.CreateContainer(ctx, containerCfg)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer svc.RemoveContainer(ctx, containerID)

	err = svc.StartContainer(ctx, containerID)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	// Test file copy
	testContent := []byte("#!/bin/bash\necho 'test script'\n")
	err = svc.CopyToContainer(ctx, containerID, "/tmp/test.sh", testContent)
	if err != nil {
		t.Errorf("Failed to copy file to container: %v", err)
	}

	// Verify file exists
	output, err := svc.ExecCommand(ctx, containerID, []string{"cat", "/tmp/test.sh"})
	if err != nil {
		t.Errorf("Failed to verify copied file: %v", err)
	}
	if output == "" {
		t.Error("Expected file content to be non-empty")
	}
}

// TestContainerConfigDefaults tests that default values are used when not specified
func TestContainerConfigDefaults(t *testing.T) {
	cfg := &config.Config{
		DockerHost:   "unix:///var/run/docker.sock",
		DefaultImage: "alpine:latest",
		MemoryLimit:  256 * 1024 * 1024,
		CPULimit:     750000000,
	}

	svc, err := NewDockerService(cfg)
	if err != nil {
		t.Skipf("Skipping test: Docker is not available: %v", err)
	}
	defer svc.Close()

	ctx := context.Background()

	// Create container with no image specified
	containerCfg := ContainerConfig{
		Name: "vibox-test-defaults",
		// Image not specified, should use default
		// MemoryLimit and CPULimit not specified, should use config defaults
	}

	containerID, err := svc.CreateContainer(ctx, containerCfg)
	if err != nil {
		t.Fatalf("Failed to create container with defaults: %v", err)
	}
	defer svc.RemoveContainer(ctx, containerID)

	// Verify container was created
	inspect, err := svc.InspectContainer(ctx, containerID)
	if err != nil {
		t.Errorf("Failed to inspect container: %v", err)
	}
	if inspect == nil {
		t.Error("Expected inspect result to be non-nil")
	}

	// Verify default image was used
	if inspect.Config.Image != cfg.DefaultImage {
		t.Errorf("Expected image to be '%s', got '%s'", cfg.DefaultImage, inspect.Config.Image)
	}

	// Verify resource limits were applied
	if inspect.HostConfig.Memory != cfg.MemoryLimit {
		t.Errorf("Expected memory limit to be %d, got %d", cfg.MemoryLimit, inspect.HostConfig.Memory)
	}
	if inspect.HostConfig.NanoCPUs != cfg.CPULimit {
		t.Errorf("Expected CPU limit to be %d, got %d", cfg.CPULimit, inspect.HostConfig.NanoCPUs)
	}
}
