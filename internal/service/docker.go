package service

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/1PercentSync/vibox/internal/config"
	"github.com/1PercentSync/vibox/pkg/utils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

// ContainerConfig holds configuration for creating a container
type ContainerConfig struct {
	Image        string
	Name         string
	MemoryLimit  int64
	CPULimit     int64
	ExposedPorts []string
}

// DockerService handles all Docker operations
type DockerService struct {
	client *client.Client
	config *config.Config
}

// NewDockerService creates a new Docker service instance
func NewDockerService(cfg *config.Config) (*DockerService, error) {
	utils.Info("Initializing Docker client", "host", cfg.DockerHost)

	// Create Docker client
	cli, err := client.NewClientWithOpts(
		client.WithHost(cfg.DockerHost),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		utils.Error("Failed to create Docker client", "error", err)
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	// Verify connection by pinging Docker daemon
	ctx := context.Background()
	_, err = cli.Ping(ctx)
	if err != nil {
		utils.Error("Failed to ping Docker daemon", "error", err)
		return nil, fmt.Errorf("failed to connect to Docker daemon: %w", err)
	}

	utils.Info("Docker client initialized successfully")

	return &DockerService{
		client: cli,
		config: cfg,
	}, nil
}

// CreateContainer creates a new Docker container
func (s *DockerService) CreateContainer(ctx context.Context, cfg ContainerConfig) (string, error) {
	utils.Info("Creating container", "image", cfg.Image, "name", cfg.Name)

	// Use default image if not specified
	image := cfg.Image
	if image == "" {
		image = s.config.DefaultImage
		utils.Debug("Using default image", "image", image)
	}

	// Use configured resource limits if not specified
	memoryLimit := cfg.MemoryLimit
	if memoryLimit == 0 {
		memoryLimit = s.config.MemoryLimit
	}
	cpuLimit := cfg.CPULimit
	if cpuLimit == 0 {
		cpuLimit = s.config.CPULimit
	}

	// Pull image if not exists
	utils.Debug("Pulling image if needed", "image", image)
	reader, err := s.client.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		utils.Error("Failed to pull image", "image", image, "error", err)
		return "", fmt.Errorf("failed to pull image %s: %w", image, err)
	}
	// Consume the reader to ensure pull completes
	_, _ = io.Copy(io.Discard, reader)
	reader.Close()
	utils.Debug("Image pulled successfully", "image", image)

	// Create container configuration
	containerConfig := &container.Config{
		Image: image,
		Tty:   true, // Enable TTY for interactive shells
		OpenStdin: true,
		AttachStdin: true,
		AttachStdout: true,
		AttachStderr: true,
		// Keep container running
		Cmd: []string{"/bin/bash"},
	}

	// Host configuration with resource limits
	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			Memory:   memoryLimit,
			NanoCPUs: cpuLimit,
		},
		// Restart policy
		RestartPolicy: container.RestartPolicy{
			Name: "no",
		},
	}

	// Network configuration
	networkConfig := &network.NetworkingConfig{}

	// Create container
	resp, err := s.client.ContainerCreate(
		ctx,
		containerConfig,
		hostConfig,
		networkConfig,
		nil,
		cfg.Name,
	)
	if err != nil {
		utils.Error("Failed to create container", "error", err)
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	utils.Info("Container created successfully", "containerID", resp.ID[:12], "name", cfg.Name)
	return resp.ID, nil
}

// StartContainer starts a container
func (s *DockerService) StartContainer(ctx context.Context, containerID string) error {
	utils.Info("Starting container", "containerID", containerID[:12])

	err := s.client.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
	if err != nil {
		utils.Error("Failed to start container", "containerID", containerID[:12], "error", err)
		return fmt.Errorf("failed to start container: %w", err)
	}

	utils.Info("Container started successfully", "containerID", containerID[:12])
	return nil
}

// StopContainer stops a container
func (s *DockerService) StopContainer(ctx context.Context, containerID string, timeout int) error {
	utils.Info("Stopping container", "containerID", containerID[:12], "timeout", timeout)

	stopOptions := container.StopOptions{}
	if timeout > 0 {
		timeoutInt := timeout
		stopOptions.Timeout = &timeoutInt
	}

	err := s.client.ContainerStop(ctx, containerID, stopOptions)
	if err != nil {
		utils.Error("Failed to stop container", "containerID", containerID[:12], "error", err)
		return fmt.Errorf("failed to stop container: %w", err)
	}

	utils.Info("Container stopped successfully", "containerID", containerID[:12])
	return nil
}

// RemoveContainer removes a container
func (s *DockerService) RemoveContainer(ctx context.Context, containerID string) error {
	utils.Info("Removing container", "containerID", containerID[:12])

	err := s.client.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{
		Force: true, // Force remove even if running
	})
	if err != nil {
		utils.Error("Failed to remove container", "containerID", containerID[:12], "error", err)
		return fmt.Errorf("failed to remove container: %w", err)
	}

	utils.Info("Container removed successfully", "containerID", containerID[:12])
	return nil
}

// GetContainerIP returns the IP address of a container
func (s *DockerService) GetContainerIP(ctx context.Context, containerID string) (string, error) {
	utils.Debug("Getting container IP", "containerID", containerID[:12])

	inspect, err := s.client.ContainerInspect(ctx, containerID)
	if err != nil {
		utils.Error("Failed to inspect container", "containerID", containerID[:12], "error", err)
		return "", fmt.Errorf("failed to inspect container: %w", err)
	}

	// Get IP from default network
	if inspect.NetworkSettings != nil && inspect.NetworkSettings.IPAddress != "" {
		ip := inspect.NetworkSettings.IPAddress
		utils.Debug("Container IP found", "containerID", containerID[:12], "ip", ip)
		return ip, nil
	}

	// Try to get IP from first available network
	if inspect.NetworkSettings != nil && len(inspect.NetworkSettings.Networks) > 0 {
		for networkName, networkConfig := range inspect.NetworkSettings.Networks {
			if networkConfig.IPAddress != "" {
				ip := networkConfig.IPAddress
				utils.Debug("Container IP found in network", "containerID", containerID[:12], "network", networkName, "ip", ip)
				return ip, nil
			}
		}
	}

	utils.Warn("No IP address found for container", "containerID", containerID[:12])
	return "", fmt.Errorf("no IP address found for container")
}

// GetContainerStatus returns the status of a container
func (s *DockerService) GetContainerStatus(ctx context.Context, containerID string) (string, error) {
	utils.Debug("Getting container status", "containerID", containerID[:12])

	inspect, err := s.client.ContainerInspect(ctx, containerID)
	if err != nil {
		utils.Error("Failed to inspect container", "containerID", containerID[:12], "error", err)
		return "", fmt.Errorf("failed to inspect container: %w", err)
	}

	status := inspect.State.Status
	utils.Debug("Container status", "containerID", containerID[:12], "status", status)
	return status, nil
}

// InspectContainer returns detailed container information
func (s *DockerService) InspectContainer(ctx context.Context, containerID string) (*types.ContainerJSON, error) {
	utils.Debug("Inspecting container", "containerID", containerID[:12])

	inspect, err := s.client.ContainerInspect(ctx, containerID)
	if err != nil {
		utils.Error("Failed to inspect container", "containerID", containerID[:12], "error", err)
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	utils.Debug("Container inspected successfully", "containerID", containerID[:12])
	return &inspect, nil
}

// ExecCommand executes a command in a container and returns the output
func (s *DockerService) ExecCommand(ctx context.Context, containerID string, cmd []string) (string, error) {
	utils.Debug("Executing command in container", "containerID", containerID[:12], "cmd", strings.Join(cmd, " "))

	// Create exec instance
	execConfig := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd,
	}

	execID, err := s.client.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		utils.Error("Failed to create exec instance", "containerID", containerID[:12], "error", err)
		return "", fmt.Errorf("failed to create exec instance: %w", err)
	}

	// Attach to exec instance
	resp, err := s.client.ContainerExecAttach(ctx, execID.ID, types.ExecStartCheck{})
	if err != nil {
		utils.Error("Failed to attach to exec instance", "execID", execID.ID[:12], "error", err)
		return "", fmt.Errorf("failed to attach to exec instance: %w", err)
	}
	defer resp.Close()

	// Read output
	var output bytes.Buffer
	_, err = io.Copy(&output, resp.Reader)
	if err != nil {
		utils.Error("Failed to read exec output", "execID", execID.ID[:12], "error", err)
		return "", fmt.Errorf("failed to read exec output: %w", err)
	}

	outputStr := output.String()
	utils.Debug("Command executed successfully", "containerID", containerID[:12], "outputLength", len(outputStr))
	return outputStr, nil
}

// CopyToContainer copies a file to a container
func (s *DockerService) CopyToContainer(ctx context.Context, containerID string, path string, content []byte) error {
	utils.Debug("Copying file to container", "containerID", containerID[:12], "path", path, "size", len(content))

	// Create tar archive
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	// Extract directory and filename from path
	lastSlash := strings.LastIndex(path, "/")
	var dir, filename string
	if lastSlash == -1 {
		dir = "/"
		filename = path
	} else {
		dir = path[:lastSlash]
		if dir == "" {
			dir = "/"
		}
		filename = path[lastSlash+1:]
	}

	// Add file to tar
	header := &tar.Header{
		Name: filename,
		Mode: 0755,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(header); err != nil {
		utils.Error("Failed to write tar header", "error", err)
		return fmt.Errorf("failed to write tar header: %w", err)
	}
	if _, err := tw.Write(content); err != nil {
		utils.Error("Failed to write tar content", "error", err)
		return fmt.Errorf("failed to write tar content: %w", err)
	}
	if err := tw.Close(); err != nil {
		utils.Error("Failed to close tar writer", "error", err)
		return fmt.Errorf("failed to close tar writer: %w", err)
	}

	// Copy to container
	err := s.client.CopyToContainer(ctx, containerID, dir, &buf, types.CopyToContainerOptions{})
	if err != nil {
		utils.Error("Failed to copy to container", "containerID", containerID[:12], "path", path, "error", err)
		return fmt.Errorf("failed to copy to container: %w", err)
	}

	utils.Debug("File copied successfully", "containerID", containerID[:12], "path", path)
	return nil
}

// Close closes the Docker client connection
func (s *DockerService) Close() error {
	utils.Info("Closing Docker client")
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}
