package middleware

import (
	"time"

	"github.com/1PercentSync/vibox/pkg/utils"
	"github.com/gin-gonic/gin"
)

// LoggerMiddleware logs HTTP requests
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		startTime := time.Now()

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(startTime)

		// Get status code
		statusCode := c.Writer.Status()

		// Log request details
		utils.Info("HTTP request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", statusCode,
			"latency", latency.String(),
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)

		// Log errors if any
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				utils.Error("Request error", "error", err.Error())
			}
		}
	}
}
