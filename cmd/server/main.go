package main

import (
	"fmt"
	"os"

	"github.com/1PercentSync/vibox/internal/api"
	"github.com/1PercentSync/vibox/internal/config"
	"github.com/1PercentSync/vibox/internal/repository"
	"github.com/1PercentSync/vibox/internal/service"
	"github.com/1PercentSync/vibox/pkg/utils"
)

func main() {
	// Initialize logger
	utils.InitLogger()
	utils.Info("Starting ViBox server...")

	// Load configuration
	cfg := config.Load()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		utils.Error("Configuration validation failed", "error", err.Error())
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	utils.Info("Configuration loaded successfully",
		"port", cfg.Port,
		"docker_host", cfg.DockerHost,
		"default_image", cfg.DefaultImage,
	)

	// Initialize Docker service
	dockerSvc, err := service.NewDockerService(cfg)
	if err != nil {
		utils.Error("Failed to initialize Docker service", "error", err.Error())
		fmt.Fprintf(os.Stderr, "ERROR: Failed to initialize Docker service: %v\n", err)
		os.Exit(1)
	}
	defer dockerSvc.Close()

	utils.Info("Docker service initialized successfully")

	// Initialize repository
	repo := repository.NewMemoryRepository()
	utils.Info("Memory repository initialized")

	// Initialize services
	workspaceSvc := service.NewWorkspaceService(dockerSvc, repo, cfg)
	utils.Info("Workspace service initialized")

	terminalSvc := service.NewTerminalService(dockerSvc)
	utils.Info("Terminal service initialized")

	proxySvc := service.NewProxyService(dockerSvc)
	utils.Info("Proxy service initialized")

	// Setup router with all services
	router := api.SetupRouter(cfg, dockerSvc, workspaceSvc, terminalSvc, proxySvc)

	// Start server
	addr := ":" + cfg.Port
	utils.Info("Server starting", "address", addr)
	if err := router.Run(addr); err != nil {
		utils.Error("Failed to start server", "error", err.Error())
		fmt.Fprintf(os.Stderr, "ERROR: Failed to start server: %v\n", err)
		os.Exit(1)
	}
}
