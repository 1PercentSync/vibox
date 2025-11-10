package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Set test environment variables
	os.Setenv("API_TOKEN", "test-token")
	os.Setenv("PORT", "8080")
	defer os.Unsetenv("API_TOKEN")
	defer os.Unsetenv("PORT")

	cfg := Load()

	if cfg.APIToken != "test-token" {
		t.Errorf("Expected API_TOKEN to be 'test-token', got '%s'", cfg.APIToken)
	}
	if cfg.Port != "8080" {
		t.Errorf("Expected PORT to be '8080', got '%s'", cfg.Port)
	}
}

func TestLoadDefaults(t *testing.T) {
	// Clear all environment variables
	os.Unsetenv("API_TOKEN")
	os.Unsetenv("PORT")
	os.Unsetenv("DOCKER_HOST")
	os.Unsetenv("DEFAULT_IMAGE")

	cfg := Load()

	if cfg.Port != "3000" {
		t.Errorf("Expected default PORT to be '3000', got '%s'", cfg.Port)
	}
	if cfg.DockerHost != "unix:///var/run/docker.sock" {
		t.Errorf("Expected default DOCKER_HOST, got '%s'", cfg.DockerHost)
	}
	if cfg.DefaultImage != "ubuntu:22.04" {
		t.Errorf("Expected default DEFAULT_IMAGE, got '%s'", cfg.DefaultImage)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Port:         "3000",
				APIToken:     "test-token",
				DockerHost:   "unix:///var/run/docker.sock",
				DefaultImage: "ubuntu:22.04",
			},
			wantErr: false,
		},
		{
			name: "missing API token",
			config: &Config{
				Port:         "3000",
				APIToken:     "",
				DockerHost:   "unix:///var/run/docker.sock",
				DefaultImage: "ubuntu:22.04",
			},
			wantErr: true,
		},
		{
			name: "missing port",
			config: &Config{
				Port:         "",
				APIToken:     "test-token",
				DockerHost:   "unix:///var/run/docker.sock",
				DefaultImage: "ubuntu:22.04",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
