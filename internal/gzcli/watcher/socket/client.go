// Package socket provides Unix socket client and server for watcher communication
package socket

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/watchertypes"
)

// Client provides a client interface to communicate with the watcher daemon
type Client struct {
	socketPath string
	timeout    time.Duration
}

// NewClient creates a new watcher client
func NewClient(socketPath string) *Client {
	if socketPath == "" {
		socketPath = watchertypes.DefaultWatcherConfig.SocketPath
	}
	return &Client{
		socketPath: socketPath,
		timeout:    30 * time.Second,
	}
}

// SetTimeout sets the connection timeout for the client
func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// SendCommand sends a command to the watcher and returns the response
func (c *Client) SendCommand(action string, data map[string]interface{}) (*watchertypes.WatcherResponse, error) {
	// Connect to the socket
	conn, err := net.DialTimeout("unix", c.socketPath, c.timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to watcher socket %s: %w", c.socketPath, err)
	}
	defer func() {
		_ = conn.Close()
	}()

	// Set read/write deadline
	deadline := time.Now().Add(c.timeout)
	if err := conn.SetDeadline(deadline); err != nil {
		return nil, fmt.Errorf("failed to set deadline: %w", err)
	}

	// Create and send command
	cmd := watchertypes.WatcherCommand{
		Action: action,
		Data:   data,
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(cmd); err != nil {
		return nil, fmt.Errorf("failed to send command: %w", err)
	}

	// Read response
	decoder := json.NewDecoder(conn)
	var response watchertypes.WatcherResponse
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// Status gets the current watcher status
func (c *Client) Status() (*watchertypes.WatcherResponse, error) {
	return c.SendCommand("status", nil)
}

// ListChallenges gets the list of watched challenges
func (c *Client) ListChallenges() (*watchertypes.WatcherResponse, error) {
	return c.SendCommand("list_challenges", nil)
}

// GetMetrics gets script execution metrics
func (c *Client) GetMetrics() (*watchertypes.WatcherResponse, error) {
	return c.SendCommand("get_metrics", nil)
}

// GetLogs gets recent logs from the database
func (c *Client) GetLogs(limit int) (*watchertypes.WatcherResponse, error) {
	data := map[string]interface{}{
		"limit": limit,
	}
	return c.SendCommand("get_logs", data)
}

// StopScript stops a specific interval script
func (c *Client) StopScript(challengeName, scriptName string) (*watchertypes.WatcherResponse, error) {
	data := map[string]interface{}{
		"challenge_name": challengeName,
		"script_name":    scriptName,
	}
	return c.SendCommand("stop_script", data)
}

// RestartChallenge triggers a full restart of a challenge
func (c *Client) RestartChallenge(challengeName string) (*watchertypes.WatcherResponse, error) {
	data := map[string]interface{}{
		"challenge_name": challengeName,
	}
	return c.SendCommand("restart_challenge", data)
}

// GetScriptExecutions gets script execution history
func (c *Client) GetScriptExecutions(challengeName string, limit int) (*watchertypes.WatcherResponse, error) {
	data := map[string]interface{}{
		"limit": limit,
	}
	if challengeName != "" {
		data["challenge_name"] = challengeName
	}
	return c.SendCommand("get_script_executions", data)
}

// IsWatcherRunning checks if the watcher daemon is running
func (c *Client) IsWatcherRunning() bool {
	response, err := c.Status()
	return err == nil && response.Success
}

// WaitForWatcher waits for the watcher to become available
func (c *Client) WaitForWatcher(maxWait time.Duration) error {
	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		if c.IsWatcherRunning() {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("watcher did not become available within %v", maxWait)
}
