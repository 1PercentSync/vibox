package handler

import (
	"net/http"

	"github.com/1PercentSync/vibox/internal/service"
	"github.com/1PercentSync/vibox/pkg/utils"
	"github.com/gin-gonic/gin"
)

// WorkspaceHandler handles workspace-related API requests
type WorkspaceHandler struct {
	service *service.WorkspaceService
}

// NewWorkspaceHandler creates a new workspace handler
func NewWorkspaceHandler(service *service.WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{
		service: service,
	}
}

// Create handles POST /api/workspaces - Create a new workspace
func (h *WorkspaceHandler) Create(c *gin.Context) {
	var req service.CreateWorkspaceRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Warn("Invalid create workspace request", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Create workspace
	workspace, err := h.service.CreateWorkspace(c.Request.Context(), req)
	if err != nil {
		utils.Error("Failed to create workspace", "error", err.Error(), "name", req.Name)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create workspace: " + err.Error(),
			"code":  "DOCKER_ERROR",
		})
		return
	}

	utils.Info("Workspace created successfully", "id", workspace.ID, "name", workspace.Name)
	c.JSON(http.StatusCreated, workspace)
}

// List handles GET /api/workspaces - List all workspaces
func (h *WorkspaceHandler) List(c *gin.Context) {
	workspaces, err := h.service.ListWorkspaces()
	if err != nil {
		utils.Error("Failed to list workspaces", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list workspaces: " + err.Error(),
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	utils.Debug("Listed workspaces", "count", len(workspaces))
	c.JSON(http.StatusOK, workspaces)
}

// Get handles GET /api/workspaces/:id - Get workspace by ID
func (h *WorkspaceHandler) Get(c *gin.Context) {
	id := c.Param("id")

	workspace, err := h.service.GetWorkspace(id)
	if err != nil {
		utils.Warn("Workspace not found", "id", id)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Workspace not found",
			"code":  "NOT_FOUND",
		})
		return
	}

	utils.Debug("Retrieved workspace", "id", id, "name", workspace.Name)
	c.JSON(http.StatusOK, workspace)
}

// Delete handles DELETE /api/workspaces/:id - Delete workspace
func (h *WorkspaceHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.service.DeleteWorkspace(c.Request.Context(), id)
	if err != nil {
		utils.Error("Failed to delete workspace", "id", id, "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete workspace: " + err.Error(),
			"code":  "DOCKER_ERROR",
		})
		return
	}

	utils.Info("Workspace deleted successfully", "id", id)
	c.JSON(http.StatusOK, gin.H{
		"message": "Workspace deleted successfully",
		"id":      id,
	})
}
