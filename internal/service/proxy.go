package service

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/1PercentSync/vibox/pkg/utils"
)

// ProxyService handles HTTP proxying to containers
type ProxyService struct {
	dockerSvc *DockerService
}

// NewProxyService creates a new proxy service instance
func NewProxyService(dockerSvc *DockerService) *ProxyService {
	utils.Info("Initializing Proxy service")
	return &ProxyService{
		dockerSvc: dockerSvc,
	}
}

// ProxyRequest proxies an HTTP request to a container's port
// This is the main entry point for forwarding requests to containers
func (s *ProxyService) ProxyRequest(w http.ResponseWriter, r *http.Request, containerID string, port int) error {
	utils.Debug("Proxying request to container",
		"containerID", containerID[:12],
		"port", port,
		"method", r.Method,
		"path", r.URL.Path,
	)

	// Get container IP address
	ctx := context.Background()
	containerIP, err := s.dockerSvc.GetContainerIP(ctx, containerID)
	if err != nil {
		utils.Error("Failed to get container IP",
			"containerID", containerID[:12],
			"error", err,
		)
		http.Error(w, "Container not found or not running", http.StatusBadGateway)
		return fmt.Errorf("failed to get container IP: %w", err)
	}

	// Create and configure reverse proxy
	proxy := s.createReverseProxy(containerIP, port)

	// Proxy the request
	proxy.ServeHTTP(w, r)

	utils.Debug("Request proxied successfully",
		"containerID", containerID[:12],
		"port", port,
	)

	return nil
}

// createReverseProxy creates a configured reverse proxy for the given target
func (s *ProxyService) createReverseProxy(containerIP string, port int) *httputil.ReverseProxy {
	// Build target URL
	targetURL := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", containerIP, port),
	}

	utils.Debug("Creating reverse proxy", "target", targetURL.String())

	// Create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Configure transport with timeouts and keep-alive
	proxy.Transport = &http.Transport{
		// Connection settings
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second, // Connection timeout
			KeepAlive: 30 * time.Second, // Keep-alive period
		}).DialContext,

		// Timeout settings
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		IdleConnTimeout:       90 * time.Second,

		// Connection pool settings
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		MaxConnsPerHost:       100,

		// Disable compression (let the client and container handle it)
		DisableCompression: false,
	}

	// Custom director to modify the request before proxying
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		// Call the original director first
		originalDirector(req)

		// Log the outgoing request
		utils.Debug("Proxying request",
			"method", req.Method,
			"url", req.URL.String(),
			"host", req.Host,
		)
	}

	// Custom error handler
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		utils.Error("Proxy error",
			"method", r.Method,
			"url", r.URL.String(),
			"error", err,
		)

		// Determine appropriate error response
		if err == context.DeadlineExceeded {
			http.Error(w, "Gateway timeout", http.StatusGatewayTimeout)
		} else if _, ok := err.(net.Error); ok {
			// Network error
			http.Error(w, "Bad gateway: unable to reach container", http.StatusBadGateway)
		} else {
			// Generic proxy error
			http.Error(w, "Proxy error", http.StatusBadGateway)
		}
	}

	// Modify response (optional, for logging)
	originalModifyResponse := proxy.ModifyResponse
	proxy.ModifyResponse = func(resp *http.Response) error {
		utils.Debug("Received response from container",
			"status", resp.StatusCode,
			"contentType", resp.Header.Get("Content-Type"),
		)

		// Call original modifier if it exists
		if originalModifyResponse != nil {
			return originalModifyResponse(resp)
		}

		return nil
	}

	return proxy
}

// GetContainerIP is a convenience method to get a container's IP address
// This can be useful for API handlers that need to check if a container is accessible
func (s *ProxyService) GetContainerIP(ctx context.Context, containerID string) (string, error) {
	return s.dockerSvc.GetContainerIP(ctx, containerID)
}
