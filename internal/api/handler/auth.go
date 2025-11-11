package handler

import (
	"net/http"

	"github.com/1PercentSync/vibox/pkg/utils"
	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related API requests
type AuthHandler struct {
	apiToken string
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(apiToken string) *AuthHandler {
	return &AuthHandler{
		apiToken: apiToken,
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Token string `json:"token" binding:"required"`
}

// Login handles POST /api/auth/login - Authenticate and set cookie
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Warn("Invalid login request", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: token is required",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Verify token
	if req.Token != h.apiToken {
		utils.Warn("Login failed: invalid token")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid token",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	// Set cookie (24 hours, HttpOnly, SameSite=Lax)
	c.SetCookie(
		"vibox-token",            // name
		req.Token,                // value
		86400,                    // maxAge: 24 hours
		"/",                      // path: global
		"",                       // domain
		false,                    // secure (should be true in production with HTTPS)
		true,                     // httpOnly: prevent JavaScript access
	)
	c.SetSameSite(http.SameSiteLaxMode) // CSRF protection

	utils.Info("User logged in successfully")
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
	})
}

// Logout handles POST /api/auth/logout - Clear authentication cookie
func (h *AuthHandler) Logout(c *gin.Context) {
	// Clear cookie by setting maxAge to -1
	c.SetCookie(
		"vibox-token",
		"",
		-1,    // maxAge: -1 deletes the cookie
		"/",   // path: must match the original cookie
		"",
		false,
		true,
	)

	utils.Info("User logged out successfully")
	c.JSON(http.StatusOK, gin.H{
		"message": "Logout successful",
	})
}
