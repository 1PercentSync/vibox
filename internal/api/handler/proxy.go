package handler

import (
	"net/http"
	"strconv"

	"github.com/1PercentSync/vibox/internal/service"
	"github.com/1PercentSync/vibox/pkg/utils"
	"github.com/gin-gonic/gin"
)

// ProxyHandler handles HTTP proxy requests to container ports
type ProxyHandler struct {
	proxyService     *service.ProxyService
	workspaceService *service.WorkspaceService
	dockerService    *service.DockerService
}

// NewProxyHandler creates a new proxy handler
func NewProxyHandler(
	proxyService *service.ProxyService,
	workspaceService *service.WorkspaceService,
	dockerService *service.DockerService,
) *ProxyHandler {
	return &ProxyHandler{
		proxyService:     proxyService,
		workspaceService: workspaceService,
		dockerService:    dockerService,
	}
}

// Forward handles ANY /forward/:id/:port/*path - Forward requests to container port
//
// Authentication:
//   - Requires X-ViBox-Token header or Authorization: Bearer token
//   - X-ViBox-Token will be automatically removed before forwarding to container
//   - Container applications can use their own Authorization header
//
// Example:
//   curl -H "X-ViBox-Token: vibox-token" \
//        -H "Authorization: Bearer app-token" \
//        http://localhost:3000/forward/ws-abc/8080/api/data
//
//   The container will receive:
//     Authorization: Bearer app-token  (kept)
//     X-ViBox-Token: (removed)
func (h *ProxyHandler) Forward(c *gin.Context) {
	workspaceID := c.Param("id")
	portStr := c.Param("port")

	// Parse port number
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		utils.Warn("Invalid port number", "workspace_id", workspaceID, "port", portStr)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid port number",
			"code":  "INVALID_REQUEST",
			"details": gin.H{
				"port": portStr,
			},
		})
		return
	}

	// 1. Verify workspace exists
	workspace, err := h.workspaceService.GetWorkspace(workspaceID)
	if err != nil {
		utils.Warn("Proxy request failed: workspace not found", "id", workspaceID)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Workspace not found",
			"code":  "NOT_FOUND",
		})
		return
	}

	// 2. Check container status
	status, err := h.dockerService.GetContainerStatus(c.Request.Context(), workspace.ContainerID)
	if err != nil {
		utils.Error("Failed to get container status", "workspace_id", workspaceID, "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to get container status: " + err.Error(),
			"code":  "CONTAINER_NOT_RUNNING",
		})
		return
	}

	if status != "running" {
		utils.Warn("Proxy request failed: container not running",
			"workspace_id", workspaceID,
			"status", status,
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Container is not running",
			"code":  "CONTAINER_NOT_RUNNING",
			"details": gin.H{
				"workspace_id": workspaceID,
				"status":       status,
			},
		})
		return
	}

	// 3. Proxy the request
	utils.Debug("Proxying request to container",
		"workspace_id", workspaceID,
		"container_id", workspace.ContainerID,
		"port", port,
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
	)

	err = h.proxyService.ProxyRequest(c.Writer, c.Request, workspace.ContainerID, port)
	if err != nil {
		utils.Error("Proxy request failed",
			"workspace_id", workspaceID,
			"port", port,
			"error", err.Error(),
		)
		// Error response already sent by ProxyService
		return
	}

	utils.Debug("Proxy request completed successfully",
		"workspace_id", workspaceID,
		"port", port,
	)
}
