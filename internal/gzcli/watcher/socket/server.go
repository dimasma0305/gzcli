package socket

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/watchertypes"
	"github.com/dimasma0305/gzcli/internal/log"
)

// Server handles Unix socket server operations
type Server struct {
	socketPath string
	listener   net.Listener
	mu         sync.RWMutex
	enabled    bool
	handler    CommandHandler
}

// CommandHandler interface for processing socket commands
type CommandHandler interface {
	HandleCommand(cmd watchertypes.WatcherCommand) watchertypes.WatcherResponse
}

// NewServer creates a new socket server
func NewServer(socketPath string, enabled bool, handler CommandHandler) *Server {
	return &Server{
		socketPath: socketPath,
		enabled:    enabled,
		handler:    handler,
	}
}

// Init initializes the socket server
func (s *Server) Init() error {
	if !s.enabled {
		log.Info("Socket server disabled")
		return nil
	}

	socketPath := s.socketPath
	// Remove existing socket file if it exists
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		log.Error("Failed to remove existing socket file: %v", err)
	}

	// Create socket directory if it doesn't exist
	socketDir := filepath.Dir(socketPath)
	if err := os.MkdirAll(socketDir, 0750); err != nil {
		return fmt.Errorf("failed to create socket directory: %w", err)
	}

	// Create Unix socket
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to create Unix socket: %w", err)
	}

	// Set socket permissions
	//nolint:gosec // G302: Unix socket needs 0666 for multi-user access
	if err := os.Chmod(socketPath, 0666); err != nil {
		_ = listener.Close()
		return fmt.Errorf("failed to set socket permissions: %w", err)
	}

	s.mu.Lock()
	s.listener = listener
	s.mu.Unlock()

	log.Info("Socket server initialized: %s", socketPath)
	return nil
}

// Close closes the socket server
func (s *Server) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener != nil {
		log.Info("Closing socket server")
		err := s.listener.Close()
		s.listener = nil

		// Clean up socket file
		if s.socketPath != "" {
			if removeErr := os.Remove(s.socketPath); removeErr != nil && !os.IsNotExist(removeErr) {
				log.Error("Failed to remove socket file: %v", removeErr)
			}
		}
		return err
	}
	return nil
}

// Run starts the socket server loop
func (s *Server) Run(ctx context.Context) {
	s.mu.RLock()
	listener := s.listener
	s.mu.RUnlock()

	if listener == nil {
		return
	}

	log.Info("Starting socket server loop")

	for {
		select {
		case <-ctx.Done():
			log.Info("Socket server loop stopped")
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				// Check if we're shutting down
				select {
				case <-ctx.Done():
					return
				default:
					log.Error("Failed to accept socket connection: %v", err)
					continue
				}
			}

			// Handle connection in goroutine
			go s.handleConnection(conn)
		}
	}
}

// handleConnection handles a single socket connection
func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		_ = conn.Close()
	}()

	// Set connection timeout
	_ = conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	var cmd watchertypes.WatcherCommand
	if err := decoder.Decode(&cmd); err != nil {
		response := watchertypes.WatcherResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to decode command: %v", err),
		}
		_ = encoder.Encode(response)
		return
	}

	// Process command using handler
	response := s.handler.HandleCommand(cmd)

	// Send response
	if err := encoder.Encode(response); err != nil {
		log.Error("Failed to send socket response: %v", err)
	}
}

// IsEnabled returns whether the socket server is enabled
func (s *Server) IsEnabled() bool {
	return s.enabled
}
