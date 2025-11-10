package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/1PercentSync/vibox/pkg/utils"
	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware recovers from panics and returns a 500 error
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic with stack trace
				stack := string(debug.Stack())
				utils.Error("Panic recovered",
					"error", fmt.Sprintf("%v", err),
					"stack", stack,
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
				)

				// Return 500 error
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
					"code":  "INTERNAL_ERROR",
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}
