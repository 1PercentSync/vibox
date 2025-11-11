package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
		"data_dir", cfg.DataDir,
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

	// Initialize repository with file persistence
	repo, err := repository.NewWorkspaceRepository(cfg.DataDir)
	if err != nil {
		utils.Error("Failed to initialize workspace repository", "error", err.Error())
		fmt.Fprintf(os.Stderr, "ERROR: Failed to initialize repository: %v\n", err)
		os.Exit(1)
	}
	utils.Info("Workspace repository initialized with persistence", "dataDir", cfg.DataDir)

	// Initialize services
	workspaceSvc := service.NewWorkspaceService(dockerSvc, repo, cfg)
	utils.Info("Workspace service initialized")

	terminalSvc := service.NewTerminalService(dockerSvc)
	utils.Info("Terminal service initialized")

	proxySvc := service.NewProxyService(dockerSvc)
	utils.Info("Proxy service initialized")

	// Restore workspaces from persistent storage
	ctx := context.Background()
	utils.Info("Restoring workspaces from persistent storage...")
	if err := workspaceSvc.RestoreWorkspaces(ctx); err != nil {
		utils.Warn("Failed to restore workspaces", "error", err.Error())
		// Continue anyway - workspaces will be in error state
	} else {
		utils.Info("Workspace restoration initiated")
	}

	// Setup router with all services
	router := api.SetupRouter(cfg, dockerSvc, workspaceSvc, terminalSvc, proxySvc)

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		utils.Info("Server starting", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.Error("Failed to start server", "error", err.Error())
			fmt.Fprintf(os.Stderr, "ERROR: Failed to start server: %v\n", err)
			os.Exit(1)
		}
	}()

	utils.Info("ViBox server is running", "port", cfg.Port)
	utils.Info("Press Ctrl+C to stop")

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for interrupt signal
	<-sigChan
	utils.Info("Received shutdown signal")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	utils.Info("Shutting down HTTP server...")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		utils.Error("Server shutdown error", "error", err.Error())
	}

	// Cleanup workspace containers
	utils.Info("Cleaning up workspace containers...")
	if err := workspaceSvc.Shutdown(shutdownCtx); err != nil {
		utils.Error("Workspace service shutdown error", "error", err.Error())
	}

	utils.Info("ViBox server stopped gracefully")
}
