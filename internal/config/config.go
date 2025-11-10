package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	Port         string
	APIToken     string
	DockerHost   string
	DefaultImage string
	MemoryLimit  int64
	CPULimit     int64
}

// Load reads configuration from environment variables
func Load() *Config {
	cfg := &Config{
		Port:         getEnv("PORT", "3000"),
		APIToken:     getEnv("API_TOKEN", ""),
		DockerHost:   getEnv("DOCKER_HOST", "unix:///var/run/docker.sock"),
		DefaultImage: getEnv("DEFAULT_IMAGE", "ubuntu:22.04"),
		MemoryLimit:  getEnvInt64("MEMORY_LIMIT", 512*1024*1024), // 512MB default
		CPULimit:     getEnvInt64("CPU_LIMIT", 1000000000),       // 1 CPU default
	}

	return cfg
}

// Validate checks that all required configuration is present
func (c *Config) Validate() error {
	if c.APIToken == "" {
		return fmt.Errorf("API_TOKEN environment variable is required but not set")
	}
	if c.Port == "" {
		return fmt.Errorf("PORT cannot be empty")
	}
	if c.DockerHost == "" {
		return fmt.Errorf("DOCKER_HOST cannot be empty")
	}
	if c.DefaultImage == "" {
		return fmt.Errorf("DEFAULT_IMAGE cannot be empty")
	}
	return nil
}

// getEnv gets an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvInt64 gets an integer environment variable with a fallback default value
func getEnvInt64(key string, defaultValue int64) int64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return defaultValue
	}
	return value
}
