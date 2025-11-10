package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates the API token from either header or query parameter
func AuthMiddleware(requiredToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Try to get token from Authorization header
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

		// 2. Try to get token from query parameter (for WebSocket connections)
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
