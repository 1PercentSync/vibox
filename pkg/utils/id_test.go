package utils

import (
	"strings"
	"testing"
)

func TestGenerateID(t *testing.T) {
	id1 := GenerateID()
	id2 := GenerateID()

	// Check format
	if !strings.HasPrefix(id1, "ws-") {
		t.Errorf("Expected ID to start with 'ws-', got '%s'", id1)
	}
	if len(id1) != 11 {
		t.Errorf("Expected ID length to be 11, got %d", len(id1))
	}

	// Check uniqueness
	if id1 == id2 {
		t.Errorf("Expected IDs to be unique, got same ID twice: %s", id1)
	}
}

func TestValidateID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "valid ID",
			id:      "ws-12345678",
			wantErr: false,
		},
		{
			name:    "empty ID",
			id:      "",
			wantErr: true,
		},
		{
			name:    "missing prefix",
			id:      "12345678",
			wantErr: true,
		},
		{
			name:    "wrong length",
			id:      "ws-123",
			wantErr: true,
		},
		{
			name:    "too long",
			id:      "ws-123456789",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateSessionID(t *testing.T) {
	sid1 := GenerateSessionID()
	sid2 := GenerateSessionID()

	// Check format
	if !strings.HasPrefix(sid1, "session-") {
		t.Errorf("Expected session ID to start with 'session-', got '%s'", sid1)
	}

	// Check uniqueness
	if sid1 == sid2 {
		t.Errorf("Expected session IDs to be unique, got same ID twice: %s", sid1)
	}
}
