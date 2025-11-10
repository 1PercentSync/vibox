package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates the API token from either header or query parameter
//
// Supported authentication methods (in priority order):
// 1. X-ViBox-Token header (recommended for API requests)
// 2. Authorization: Bearer <token> header (legacy support)
// 3. ?token=<token> query parameter (for WebSocket connections)
//
// Note: X-ViBox-Token is preferred to avoid conflicts with container applications
// that may use their own Authorization headers when proxying requests.
func AuthMiddleware(requiredToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Try to get token from X-ViBox-Token header (preferred)
		viboxToken := c.GetHeader("X-ViBox-Token")
		if viboxToken != "" {
			if viboxToken == requiredToken {
				c.Next()
				return
			}
		}

		// 2. Try to get token from Authorization header (legacy support)
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			// Check if it's a Bearer token
			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				if token == requiredToken {
					c.Next()
					return
				}
			}
		}

		// 3. Try to get token from query parameter (for WebSocket connections)
		token := c.Query("token")
		if token != "" && token == requiredToken {
			c.Next()
			return
		}

		// Authentication failed
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized: invalid or missing token",
			"code":  "UNAUTHORIZED",
		})
		c.Abort()
	}
}
