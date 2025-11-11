package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/gorilla/websocket"

	"github.com/1PercentSync/vibox/pkg/utils"
)

// TerminalService manages WebSocket terminal sessions connected to Docker containers
type TerminalService struct {
	dockerSvc *DockerService
	sessions  sync.Map // map[sessionID]*TerminalSession
}

// TerminalSession represents an active terminal session
type TerminalSession struct {
	ID          string
	ContainerID string
	WebSocket   *websocket.Conn
	ExecID      string
	HijackedConn io.Closer
	CreatedAt   time.Time
	CancelFunc  context.CancelFunc
	Done        chan struct{}
}

// TerminalMessage represents a message exchanged over WebSocket
type TerminalMessage struct {
	Type string `json:"type"` // "input", "output", "resize", "error", "close"
	Data string `json:"data,omitempty"`
	Cols int    `json:"cols,omitempty"`
	Rows int    `json:"rows,omitempty"`
}

// NewTerminalService creates a new terminal service
func NewTerminalService(dockerSvc *DockerService) *TerminalService {
	utils.Info("Creating new terminal service")
	return &TerminalService{
		dockerSvc: dockerSvc,
		sessions:  sync.Map{},
	}
}

// CreateSession creates a new terminal session with WebSocket and Docker Exec
func (s *TerminalService) CreateSession(ctx context.Context, ws *websocket.Conn, containerID string) error {
	// Generate session ID
	sessionID := utils.GenerateSessionID()
	utils.Info("Creating terminal session", "sessionID", sessionID, "containerID", containerID)

	// Verify container is running
	status, err := s.dockerSvc.GetContainerStatus(ctx, containerID)
	if err != nil {
		utils.Error("Failed to get container status", "containerID", containerID, "error", err)
		return fmt.Errorf("failed to get container status: %w", err)
	}
	if status != "running" {
		utils.Warn("Container is not running", "containerID", containerID, "status", status)
		return fmt.Errorf("container is not running (status: %s)", status)
	}

	// Create exec instance with TTY
	// Try bash first (for arrow keys and command history), fallback to sh
	shell := "/bin/bash"

	// Check if bash exists in container
	checkExec, err := s.dockerSvc.client.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		Cmd:          []string{"sh", "-c", "which bash || echo notfound"},
		AttachStdout: true,
	})
	if err == nil {
		checkResp, err := s.dockerSvc.client.ContainerExecAttach(ctx, checkExec.ID, container.ExecStartOptions{})
		if err == nil {
			output := make([]byte, 256)
			n, _ := checkResp.Reader.Read(output)
			if n > 0 && string(output[:n]) != "notfound\n" && string(output[:n]) != "notfound" {
				utils.Debug("Bash found in container", "containerID", containerID)
			} else {
				shell = "/bin/sh"
				utils.Debug("Bash not found, using sh", "containerID", containerID)
			}
			checkResp.Close()
		}
	}

	execConfig := container.ExecOptions{
		Cmd:          []string{shell},
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true, // Critical for interactive terminal
	}

	execID, err := s.dockerSvc.client.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		utils.Error("Failed to create exec", "containerID", containerID, "error", err)
		return fmt.Errorf("failed to create exec: %w", err)
	}

	utils.Debug("Created exec instance", "execID", execID.ID)

	// Attach to exec
	attachConfig := container.ExecStartOptions{
		Tty: true,
	}

	hijackedResp, err := s.dockerSvc.client.ContainerExecAttach(ctx, execID.ID, attachConfig)
	if err != nil {
		utils.Error("Failed to attach to exec", "execID", execID.ID, "error", err)
		return fmt.Errorf("failed to attach to exec: %w", err)
	}

	utils.Debug("Attached to exec", "execID", execID.ID)

	// Create cancellable context for this session
	sessionCtx, cancel := context.WithCancel(ctx)

	// Create session
	session := &TerminalSession{
		ID:           sessionID,
		ContainerID:  containerID,
		WebSocket:    ws,
		ExecID:       execID.ID,
		HijackedConn: hijackedResp.Conn,
		CreatedAt:    time.Now(),
		CancelFunc:   cancel,
		Done:         make(chan struct{}),
	}

	// Store session
	s.sessions.Store(sessionID, session)

	utils.Info("Terminal session created", "sessionID", sessionID, "containerID", containerID)

	// Start bidirectional data transfer
	go s.handleWebSocketToExec(sessionCtx, session, hijackedResp.Conn)
	go s.handleExecToWebSocket(sessionCtx, session, hijackedResp.Conn)

	// Wait for session to complete
	<-session.Done

	return nil
}

// handleWebSocketToExec transfers data from WebSocket to Docker Exec
func (s *TerminalService) handleWebSocketToExec(ctx context.Context, session *TerminalSession, execConn io.WriteCloser) {
	defer func() {
		utils.Debug("WebSocket to Exec handler stopped", "sessionID", session.ID)
		s.cleanupSession(session)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Read message from WebSocket
			var msg TerminalMessage
			err := session.WebSocket.ReadJSON(&msg)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					utils.Warn("WebSocket read error", "sessionID", session.ID, "error", err)
				}
				return
			}

			// Handle different message types
			switch msg.Type {
			case "input":
				// Send input to container
				_, err := execConn.Write([]byte(msg.Data))
				if err != nil {
					utils.Error("Failed to write to exec", "sessionID", session.ID, "error", err)
					s.sendMessage(session.WebSocket, TerminalMessage{
						Type: "error",
						Data: "Failed to send input to container",
					})
					return
				}

			case "resize":
				// Resize terminal
				if msg.Cols > 0 && msg.Rows > 0 {
					err := s.resizeTerminal(ctx, session.ExecID, msg.Cols, msg.Rows)
					if err != nil {
						utils.Warn("Failed to resize terminal", "sessionID", session.ID, "error", err)
					}
				}

			default:
				utils.Warn("Unknown message type", "sessionID", session.ID, "type", msg.Type)
			}
		}
	}
}

// handleExecToWebSocket transfers data from Docker Exec to WebSocket
func (s *TerminalService) handleExecToWebSocket(ctx context.Context, session *TerminalSession, execConn io.Reader) {
	defer func() {
		utils.Debug("Exec to WebSocket handler stopped", "sessionID", session.ID)
		s.cleanupSession(session)
	}()

	buffer := make([]byte, 8192)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Read from exec connection
			n, err := execConn.Read(buffer)
			if err != nil {
				if err != io.EOF {
					utils.Warn("Failed to read from exec", "sessionID", session.ID, "error", err)
				}
				return
			}

			if n > 0 {
				// Send output to WebSocket
				msg := TerminalMessage{
					Type: "output",
					Data: string(buffer[:n]),
				}

				err := s.sendMessage(session.WebSocket, msg)
				if err != nil {
					utils.Error("Failed to send to WebSocket", "sessionID", session.ID, "error", err)
					return
				}
			}
		}
	}
}

// resizeTerminal resizes the terminal to the specified dimensions
func (s *TerminalService) resizeTerminal(ctx context.Context, execID string, cols, rows int) error {
	utils.Debug("Resizing terminal", "execID", execID, "cols", cols, "rows", rows)

	resizeOptions := container.ResizeOptions{
		Height: uint(rows),
		Width:  uint(cols),
	}

	err := s.dockerSvc.client.ContainerExecResize(ctx, execID, resizeOptions)
	if err != nil {
		return fmt.Errorf("failed to resize terminal: %w", err)
	}

	return nil
}

// sendMessage sends a message to the WebSocket connection
func (s *TerminalService) sendMessage(ws *websocket.Conn, msg TerminalMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = ws.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return fmt.Errorf("failed to write to websocket: %w", err)
	}

	return nil
}

// cleanupSession cleans up a terminal session
func (s *TerminalService) cleanupSession(session *TerminalSession) {
	// Use sync.Once to ensure cleanup only happens once
	select {
	case <-session.Done:
		// Already cleaned up
		return
	default:
		close(session.Done)
	}

	utils.Info("Cleaning up terminal session", "sessionID", session.ID)

	// Cancel context
	if session.CancelFunc != nil {
		session.CancelFunc()
	}

	// Close hijacked connection
	if session.HijackedConn != nil {
		session.HijackedConn.Close()
	}

	// Close WebSocket
	if session.WebSocket != nil {
		s.sendMessage(session.WebSocket, TerminalMessage{
			Type: "close",
			Data: "Session closed",
		})
		session.WebSocket.Close()
	}

	// Remove from sessions map
	s.sessions.Delete(session.ID)

	utils.Info("Terminal session cleaned up", "sessionID", session.ID)
}

// CloseSession closes a terminal session by ID
func (s *TerminalService) CloseSession(sessionID string) error {
	utils.Info("Closing terminal session", "sessionID", sessionID)

	value, ok := s.sessions.Load(sessionID)
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session := value.(*TerminalSession)
	s.cleanupSession(session)

	return nil
}

// GetSessionCount returns the number of active sessions
func (s *TerminalService) GetSessionCount() int {
	count := 0
	s.sessions.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// CloseAllSessions closes all active sessions
func (s *TerminalService) CloseAllSessions() {
	utils.Info("Closing all terminal sessions")

	s.sessions.Range(func(key, value interface{}) bool {
		session := value.(*TerminalSession)
		s.cleanupSession(session)
		return true
	})

	utils.Info("All terminal sessions closed")
}
