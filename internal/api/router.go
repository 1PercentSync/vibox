package api

import (
	"github.com/1PercentSync/vibox/internal/api/handler"
	"github.com/1PercentSync/vibox/internal/api/middleware"
	"github.com/1PercentSync/vibox/internal/config"
	"github.com/1PercentSync/vibox/internal/service"
	"github.com/gin-gonic/gin"
)

// SetupRouter configures and returns the Gin router with all routes and middleware
func SetupRouter(
	cfg *config.Config,
	dockerSvc *service.DockerService,
	workspaceSvc *service.WorkspaceService,
	terminalSvc *service.TerminalService,
	proxySvc *service.ProxyService,
) *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Create router
	router := gin.New()

	// Apply global middleware
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.CORSMiddleware())

	// Create handlers
	workspaceHandler := handler.NewWorkspaceHandler(workspaceSvc)
	terminalHandler := handler.NewTerminalHandler(terminalSvc, workspaceSvc, dockerSvc)
	proxyHandler := handler.NewProxyHandler(proxySvc, workspaceSvc, dockerSvc)

	// Health check endpoint (no auth required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "vibox",
		})
	})

	// API routes (with auth)
	api := router.Group("/api")
	api.Use(middleware.AuthMiddleware(cfg.APIToken))
	{
		// Workspace management
		api.POST("/workspaces", workspaceHandler.Create)
		api.GET("/workspaces", workspaceHandler.List)
		api.GET("/workspaces/:id", workspaceHandler.Get)
		api.DELETE("/workspaces/:id", workspaceHandler.Delete)
	}

	// WebSocket terminal (with auth)
	// Note: WebSocket connections must use ?token= query parameter for auth
	router.GET("/ws/terminal/:id",
		middleware.AuthMiddleware(cfg.APIToken),
		terminalHandler.Connect,
	)

	// Port forwarding (with auth)
	// Matches: /forward/{workspace-id}/{port}/any/path
	router.Any("/forward/:id/:port/*path",
		middleware.AuthMiddleware(cfg.APIToken),
		proxyHandler.Forward,
	)

	return router
}
