package main

import (
	"fmt"
	"os"

	"github.com/1PercentSync/vibox/internal/api/middleware"
	"github.com/1PercentSync/vibox/internal/config"
	"github.com/1PercentSync/vibox/pkg/utils"
	"github.com/gin-gonic/gin"
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

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Create Gin router
	router := gin.New()

	// Apply global middleware
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.CORSMiddleware())

	// Health check endpoint (no auth required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"service": "vibox",
		})
	})

	// API routes (with auth)
	api := router.Group("/api")
	api.Use(middleware.AuthMiddleware(cfg.APIToken))
	{
		// Placeholder routes - will be implemented in later modules
		api.GET("/workspaces", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Workspace API - Coming soon",
			})
		})
	}

	// Start server
	addr := ":" + cfg.Port
	utils.Info("Server starting", "address", addr)
	if err := router.Run(addr); err != nil {
		utils.Error("Failed to start server", "error", err.Error())
		fmt.Fprintf(os.Stderr, "ERROR: Failed to start server: %v\n", err)
		os.Exit(1)
	}
}
