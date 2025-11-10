package utils

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// GenerateID generates a new unique ID for workspaces
// Format: ws-{8-char-hex}
func GenerateID() string {
	id := uuid.New()
	// Use first 8 characters of UUID hex representation
	shortID := strings.ReplaceAll(id.String(), "-", "")[:8]
	return fmt.Sprintf("ws-%s", shortID)
}

// ValidateID validates that an ID has the correct format
func ValidateID(id string) error {
	if id == "" {
		return fmt.Errorf("ID cannot be empty")
	}
	if !strings.HasPrefix(id, "ws-") {
		return fmt.Errorf("ID must start with 'ws-' prefix")
	}
	if len(id) != 11 { // ws- (3) + 8 characters
		return fmt.Errorf("ID must be exactly 11 characters long (ws-XXXXXXXX)")
	}
	return nil
}

// GenerateSessionID generates a unique session ID for terminal sessions
func GenerateSessionID() string {
	id := uuid.New()
	// Use first 8 characters for session ID
	shortID := strings.ReplaceAll(id.String(), "-", "")[:8]
	return fmt.Sprintf("session-%s", shortID)
}

// ShortID returns a shortened version of a Docker ID (first 12 characters)
// This is safe even if the ID is shorter than 12 characters
func ShortID(id string) string {
	if len(id) >= 12 {
		return id[:12]
	}
	return id
}
