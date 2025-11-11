package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates the API token from Cookie or query parameter
//
// Supported authentication methods (in priority order):
// 1. Cookie: vibox-token (recommended for browser access)
// 2. Query parameter: ?token=<token> (for WebSocket connections only)
//
// For browser requests without authentication:
// - HTML requests (Accept: text/html) → Redirect to /login
// - API requests (Accept: application/json) → Return 401 JSON
//
// Note: Cookie authentication is the primary method. Query parameter is only
// supported for WebSocket connections where cookies are difficult to manage.
func AuthMiddleware(requiredToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Try to get token from Cookie (preferred)
		token, err := c.Cookie("vibox-token")
		if err == nil && token == requiredToken {
			c.Next()
			return
		}

		// 2. Try to get token from query parameter (for WebSocket only)
		token = c.Query("token")
		if token != "" && token == requiredToken {
			c.Next()
			return
		}

		// Authentication failed - handle based on request type
		accept := c.GetHeader("Accept")
		isHTMLRequest := strings.Contains(accept, "text/html")

		if isHTMLRequest {
			// Browser request → Redirect to login page
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// API request → Return JSON error
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized: invalid or missing authentication",
			"code":  "UNAUTHORIZED",
		})
		c.Abort()
	}
}
