package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testToken := "test-secret-token"

	tests := []struct {
		name           string
		viboxToken     string
		authHeader     string
		queryToken     string
		expectedStatus int
	}{
		{
			name:           "valid X-ViBox-Token header",
			viboxToken:     "test-secret-token",
			authHeader:     "",
			queryToken:     "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid bearer token",
			viboxToken:     "",
			authHeader:     "Bearer test-secret-token",
			queryToken:     "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid query token",
			viboxToken:     "",
			authHeader:     "",
			queryToken:     "test-secret-token",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "X-ViBox-Token takes priority over Authorization",
			viboxToken:     "test-secret-token",
			authHeader:     "Bearer wrong-token",
			queryToken:     "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid X-ViBox-Token",
			viboxToken:     "wrong-token",
			authHeader:     "",
			queryToken:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid bearer token",
			viboxToken:     "",
			authHeader:     "Bearer wrong-token",
			queryToken:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid query token",
			viboxToken:     "",
			authHeader:     "",
			queryToken:     "wrong-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "no token",
			viboxToken:     "",
			authHeader:     "",
			queryToken:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "malformed bearer token",
			viboxToken:     "",
			authHeader:     "test-secret-token",
			queryToken:     "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			router := gin.New()
			router.Use(AuthMiddleware(testToken))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// Create request
			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.viboxToken != "" {
				req.Header.Set("X-ViBox-Token", tt.viboxToken)
			}
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			if tt.queryToken != "" {
				q := req.URL.Query()
				q.Add("token", tt.queryToken)
				req.URL.RawQuery = q.Encode()
			}

			// Execute
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
