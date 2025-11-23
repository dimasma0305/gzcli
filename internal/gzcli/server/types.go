package server

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
)

// ChallengeStatus represents the current state of a challenge
type ChallengeStatus string

// Challenge status constants
const (
	// StatusStopped indicates the challenge is not running
	StatusStopped ChallengeStatus = "stopped"
	// StatusStarting indicates the challenge is in the process of starting
	StatusStarting ChallengeStatus = "starting"
	// StatusRunning indicates the challenge is running and operational
	StatusRunning ChallengeStatus = "running"
	// StatusStopping indicates the challenge is in the process of stopping
	StatusStopping ChallengeStatus = "stopping"
	// StatusRestarting indicates the challenge is restarting
	StatusRestarting ChallengeStatus = "restarting"
	// StatusUnhealthy indicates the challenge is running but not healthy
	StatusUnhealthy ChallengeStatus = "unhealthy"
)

// LauncherType represents the type of launcher configuration
type LauncherType string

// Launcher type constants
const (
	// LauncherTypeCompose represents Docker Compose configuration
	LauncherTypeCompose LauncherType = "compose"
	// LauncherTypeDockerfile represents Dockerfile configuration
	LauncherTypeDockerfile LauncherType = "dockerfile"
	// LauncherTypeKubernetes represents Kubernetes manifest configuration
	LauncherTypeKubernetes LauncherType = "kubernetes"
)

// Dashboard represents the dashboard configuration from challenge.yml
type Dashboard struct {
	Type   string   `yaml:"type"`
	Config string   `yaml:"config"`
	Ports  []string `yaml:"ports"` // For dockerfile type
}

// ChallengeInfo holds information about a discovered challenge
type ChallengeInfo struct {
	Slug           string
	EventName      string
	Category       string
	Name           string
	Description    string
	Cwd            string // Working directory for scripts
	Dashboard      *Dashboard
	Scripts        map[string]config.ScriptValue
	Status         ChallengeStatus
	LastRestart    time.Time
	AllocatedPorts []string        // Dynamically allocated ports (host:container)
	ConnectedIPs   map[string]bool // Track unique IPs connected
	mu             sync.RWMutex
}

// Client represents a WebSocket client connection
type Client struct {
	Conn      *websocket.Conn
	IP        string
	Challenge string // Challenge slug
	Send      chan []byte
}

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type    string      `json:"type"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// StatusMessage represents a status update message
type StatusMessage struct {
	Status         string   `json:"status"`
	ConnectedUsers int      `json:"connected_users"`
	AllocatedPorts []string `json:"allocated_ports,omitempty"`
}

// VoteMessage represents a vote-related message
type VoteMessage struct {
	InitiatorIP  string  `json:"initiator_ip,omitempty"`
	YesPercent   float64 `json:"yes_percent,omitempty"`
	NoPercent    float64 `json:"no_percent,omitempty"`
	TotalUsers   int     `json:"total_users,omitempty"`
	Result       string  `json:"result,omitempty"`
	RemainingMin int     `json:"remaining_min,omitempty"`
}

// Vote represents a restart vote
type Vote struct {
	InitiatedAt time.Time
	Votes       map[string]bool // IP -> true (yes) or false (no)
	mu          sync.RWMutex
}

// PortInfo represents port mapping information
type PortInfo struct {
	Service  string
	Internal string
	External string
	Protocol string
}

// GetConnectedUsers returns the number of unique IPs connected
func (c *ChallengeInfo) GetConnectedUsers() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.ConnectedIPs)
}

// AddConnectedIP adds an IP to the connected users
func (c *ChallengeInfo) AddConnectedIP(ip string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ConnectedIPs == nil {
		c.ConnectedIPs = make(map[string]bool)
	}
	c.ConnectedIPs[ip] = true
}

// RemoveConnectedIP removes an IP from the connected users
func (c *ChallengeInfo) RemoveConnectedIP(ip string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.ConnectedIPs, ip)
}

// SetStatus safely sets the challenge status
func (c *ChallengeInfo) SetStatus(status ChallengeStatus) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Status = status
}

// GetStatus safely gets the challenge status
func (c *ChallengeInfo) GetStatus() ChallengeStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Status
}

// SetAllocatedPorts safely sets the allocated ports
func (c *ChallengeInfo) SetAllocatedPorts(ports []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.AllocatedPorts = ports
}

// GetAllocatedPorts safely gets the allocated ports
func (c *ChallengeInfo) GetAllocatedPorts() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.AllocatedPorts
}

// IsInCooldown checks if the challenge is in restart cooldown period
// Uses a fixed 5-minute cooldown period
func (c *ChallengeInfo) IsInCooldown() (bool, time.Duration) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	const cooldown = 5 * time.Minute
	elapsed := time.Since(c.LastRestart)

	if elapsed < cooldown {
		return true, cooldown - elapsed
	}

	return false, 0
}

// SetLastRestart sets the last restart time
func (c *ChallengeInfo) SetLastRestart(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LastRestart = t
}

// CalculateGracePeriod calculates the auto-stop grace period
// Uses a fixed 2-minute grace period
func (c *ChallengeInfo) CalculateGracePeriod() time.Duration {
	return 2 * time.Minute
}
