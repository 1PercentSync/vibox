package handler

import (
	"net/http"

	"github.com/1PercentSync/vibox/internal/service"
	"github.com/1PercentSync/vibox/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// TerminalHandler handles terminal WebSocket connections
type TerminalHandler struct {
	terminalService  *service.TerminalService
	workspaceService *service.WorkspaceService
	dockerService    *service.DockerService
}

// NewTerminalHandler creates a new terminal handler
func NewTerminalHandler(
	terminalService *service.TerminalService,
	workspaceService *service.WorkspaceService,
	dockerService *service.DockerService,
) *TerminalHandler {
	return &TerminalHandler{
		terminalService:  terminalService,
		workspaceService: workspaceService,
		dockerService:    dockerService,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
	CheckOrigin: func(r *http.Request) bool {
		// Currently allowing all origins for simplicity.
		// Since we use API token authentication (not cookies),
		// the CSRF risk from CheckOrigin is minimal.
		//
		// If token is compromised, attackers can make direct API calls anyway.
		// CheckOrigin only prevents malicious websites from using
		// tokens stored in user's browser.
		//
		// TODO: For production with sensitive data, consider adding
		// origin whitelist via ALLOWED_ORIGINS env var.
		return true
	},
}

// Connect handles GET /ws/terminal/:id - Connect to workspace terminal
func (h *TerminalHandler) Connect(c *gin.Context) {
	workspaceID := c.Param("id")

	// 1. Verify workspace exists
	workspace, err := h.workspaceService.GetWorkspace(workspaceID)
	if err != nil {
		utils.Warn("Terminal connection failed: workspace not found", "id", workspaceID)
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
		utils.Warn("Terminal connection failed: container not running",
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

	// 3. Upgrade to WebSocket
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		utils.Error("Failed to upgrade to WebSocket", "workspace_id", workspaceID, "error", err.Error())
		// Response already sent by upgrader
		return
	}

	utils.Info("WebSocket connection established", "workspace_id", workspaceID, "container_id", workspace.ContainerID)

	// 4. Create terminal session
	err = h.terminalService.CreateSession(c.Request.Context(), ws, workspace.ContainerID)
	if err != nil {
		utils.Error("Terminal session error", "workspace_id", workspaceID, "error", err.Error())
		// Session will be cleaned up by TerminalService
	}
}
