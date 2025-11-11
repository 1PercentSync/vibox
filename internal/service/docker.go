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
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// ContainerConfig holds configuration for creating a container
type ContainerConfig struct {
	Image       string
	Name        string
	MemoryLimit int64
	CPULimit    int64
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
	imageName := cfg.Image
	if imageName == "" {
		imageName = s.config.DefaultImage
		utils.Debug("Using default image", "image", imageName)
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
	utils.Debug("Pulling image if needed", "image", imageName)
	reader, err := s.client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		utils.Error("Failed to pull image", "image", imageName, "error", err)
		return "", fmt.Errorf("failed to pull image %s: %w", imageName, err)
	}
	// Consume the reader to ensure pull completes
	_, _ = io.Copy(io.Discard, reader)
	reader.Close()
	utils.Debug("Image pulled successfully", "image", imageName)

	// Create container configuration
	containerConfig := &container.Config{
		Image: imageName,
		Tty:   true, // Enable TTY for interactive shells
		OpenStdin: true,
		AttachStdin: true,
		AttachStdout: true,
		AttachStderr: true,
		// Keep container running - use /bin/sh for maximum compatibility (including Alpine)
		Cmd: []string{"/bin/sh"},
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

	// Network configuration - connect to vibox-network for inter-container communication
	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"vibox-network": {},
		},
	}

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

	utils.Info("Container created successfully", "containerID", utils.ShortID(resp.ID), "name", cfg.Name)
	return resp.ID, nil
}

// StartContainer starts a container
func (s *DockerService) StartContainer(ctx context.Context, containerID string) error {
	utils.Info("Starting container", "containerID", utils.ShortID(containerID))

	err := s.client.ContainerStart(ctx, containerID, container.StartOptions{})
	if err != nil {
		utils.Error("Failed to start container", "containerID", utils.ShortID(containerID), "error", err)
		return fmt.Errorf("failed to start container: %w", err)
	}

	utils.Info("Container started successfully", "containerID", utils.ShortID(containerID))
	return nil
}

// StopContainer stops a container
func (s *DockerService) StopContainer(ctx context.Context, containerID string, timeout int) error {
	utils.Info("Stopping container", "containerID", utils.ShortID(containerID), "timeout", timeout)

	stopOptions := container.StopOptions{}
	if timeout > 0 {
		timeoutInt := timeout
		stopOptions.Timeout = &timeoutInt
	}

	err := s.client.ContainerStop(ctx, containerID, stopOptions)
	if err != nil {
		utils.Error("Failed to stop container", "containerID", utils.ShortID(containerID), "error", err)
		return fmt.Errorf("failed to stop container: %w", err)
	}

	utils.Info("Container stopped successfully", "containerID", utils.ShortID(containerID))
	return nil
}

// RemoveContainer removes a container
func (s *DockerService) RemoveContainer(ctx context.Context, containerID string) error {
	utils.Info("Removing container", "containerID", utils.ShortID(containerID))

	err := s.client.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force: true, // Force remove even if running
	})
	if err != nil {
		utils.Error("Failed to remove container", "containerID", utils.ShortID(containerID), "error", err)
		return fmt.Errorf("failed to remove container: %w", err)
	}

	utils.Info("Container removed successfully", "containerID", utils.ShortID(containerID))
	return nil
}

// GetContainerIP returns the IP address of a container
func (s *DockerService) GetContainerIP(ctx context.Context, containerID string) (string, error) {
	utils.Debug("Getting container IP", "containerID", utils.ShortID(containerID))

	inspect, err := s.client.ContainerInspect(ctx, containerID)
	if err != nil {
		utils.Error("Failed to inspect container", "containerID", utils.ShortID(containerID), "error", err)
		return "", fmt.Errorf("failed to inspect container: %w", err)
	}

	// Get IP from default network
	if inspect.NetworkSettings != nil && inspect.NetworkSettings.IPAddress != "" {
		ip := inspect.NetworkSettings.IPAddress
		utils.Debug("Container IP found", "containerID", utils.ShortID(containerID), "ip", ip)
		return ip, nil
	}

	// Try to get IP from first available network
	if inspect.NetworkSettings != nil && len(inspect.NetworkSettings.Networks) > 0 {
		for networkName, networkConfig := range inspect.NetworkSettings.Networks {
			if networkConfig.IPAddress != "" {
				ip := networkConfig.IPAddress
				utils.Debug("Container IP found in network", "containerID", utils.ShortID(containerID), "network", networkName, "ip", ip)
				return ip, nil
			}
		}
	}

	utils.Warn("No IP address found for container", "containerID", utils.ShortID(containerID))
	return "", fmt.Errorf("no IP address found for container")
}

// GetContainerStatus returns the status of a container
func (s *DockerService) GetContainerStatus(ctx context.Context, containerID string) (string, error) {
	utils.Debug("Getting container status", "containerID", utils.ShortID(containerID))

	inspect, err := s.client.ContainerInspect(ctx, containerID)
	if err != nil {
		utils.Error("Failed to inspect container", "containerID", utils.ShortID(containerID), "error", err)
		return "", fmt.Errorf("failed to inspect container: %w", err)
	}

	status := inspect.State.Status
	utils.Debug("Container status", "containerID", utils.ShortID(containerID), "status", status)
	return status, nil
}

// InspectContainer returns detailed container information
func (s *DockerService) InspectContainer(ctx context.Context, containerID string) (*types.ContainerJSON, error) {
	utils.Debug("Inspecting container", "containerID", utils.ShortID(containerID))

	inspect, err := s.client.ContainerInspect(ctx, containerID)
	if err != nil {
		utils.Error("Failed to inspect container", "containerID", utils.ShortID(containerID), "error", err)
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	utils.Debug("Container inspected successfully", "containerID", utils.ShortID(containerID))
	return &inspect, nil
}

// ExecCommand executes a command in a container and returns the output
func (s *DockerService) ExecCommand(ctx context.Context, containerID string, cmd []string) (string, error) {
	utils.Debug("Executing command in container", "containerID", utils.ShortID(containerID), "cmd", strings.Join(cmd, " "))

	// Create exec instance
	execConfig := container.ExecOptions{
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false, // No TTY for script execution (keeps stdout/stderr separate)
		Cmd:          cmd,
	}

	execID, err := s.client.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		utils.Error("Failed to create exec instance", "containerID", utils.ShortID(containerID), "error", err)
		return "", fmt.Errorf("failed to create exec instance: %w", err)
	}

	// Attach to exec instance
	resp, err := s.client.ContainerExecAttach(ctx, execID.ID, container.ExecStartOptions{})
	if err != nil {
		utils.Error("Failed to attach to exec instance", "execID", utils.ShortID(execID.ID), "error", err)
		return "", fmt.Errorf("failed to attach to exec instance: %w", err)
	}
	defer resp.Close()

	// Read output
	// When Tty=false, Docker uses stream multiplexing with 8-byte headers
	// We need to use stdcopy.StdCopy to properly demultiplex stdout/stderr
	var stdout, stderr bytes.Buffer
	_, err = stdcopy.StdCopy(&stdout, &stderr, resp.Reader)
	if err != nil {
		utils.Error("Failed to read exec output", "execID", utils.ShortID(execID.ID), "error", err)
		return "", fmt.Errorf("failed to read exec output: %w", err)
	}

	// Combine stdout and stderr
	var output bytes.Buffer
	output.Write(stdout.Bytes())
	output.Write(stderr.Bytes())

	outputStr := output.String()
	utils.Debug("Command executed successfully", "containerID", utils.ShortID(containerID), "outputLength", len(outputStr))
	return outputStr, nil
}

// CopyToContainer copies a file to a container
func (s *DockerService) CopyToContainer(ctx context.Context, containerID string, path string, content []byte) error {
	utils.Debug("Copying file to container", "containerID", utils.ShortID(containerID), "path", path, "size", len(content))

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
	err := s.client.CopyToContainer(ctx, containerID, dir, &buf, container.CopyToContainerOptions{})
	if err != nil {
		utils.Error("Failed to copy to container", "containerID", utils.ShortID(containerID), "path", path, "error", err)
		return fmt.Errorf("failed to copy to container: %w", err)
	}

	utils.Debug("File copied successfully", "containerID", utils.ShortID(containerID), "path", path)
	return nil
}

// ListContainers lists containers matching the given filters
func (s *DockerService) ListContainers(ctx context.Context, filterMap map[string]string) ([]types.Container, error) {
	utils.Debug("Listing containers", "filters", filterMap)

	// Create filter args
	filterArgs := filters.NewArgs()

	// Build filter arguments if provided
	if len(filterMap) > 0 {
		// For simplicity, we support label filters which is the main use case
		for key, value := range filterMap {
			if key == "label" {
				// Label filter: we pass the label value directly
				// Docker will match containers with this label
				filterArgs.Add("label", value)
			}
		}
	}

	// Create list options with all containers (including stopped)
	listOptions := container.ListOptions{
		All:     true,
		Filters: filterArgs,
	}

	containers, err := s.client.ContainerList(ctx, listOptions)
	if err != nil {
		utils.Error("Failed to list containers", "error", err)
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	utils.Debug("Listed containers", "count", len(containers))
	return containers, nil
}

// Close closes the Docker client connection
func (s *DockerService) Close() error {
	utils.Info("Closing Docker client")
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}
