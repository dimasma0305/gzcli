//nolint:revive // Package types provides type definitions for the watcher
package types

import (
	"time"
)

// ScriptMetrics tracks execution statistics for scripts
type ScriptMetrics struct {
	LastExecution  time.Time
	ExecutionCount int64
	LastError      error
	LastDuration   time.Duration
	TotalDuration  time.Duration
	Interval       time.Duration `json:"interval,omitempty"` // For interval scripts
	IsInterval     bool          `json:"is_interval"`        // Whether this is an interval script
}

// WatcherCommand represents commands that can be sent to the watcher via socket
type WatcherCommand struct {
	Action string                 `json:"action"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

// WatcherResponse represents responses from the watcher
type WatcherResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// UpdateType represents the type of update needed based on file changes
type UpdateType int

// Update type constants
const (
	// UpdateNone indicates no update is needed
	UpdateNone UpdateType = iota
	UpdateAttachment
	UpdateMetadata
	UpdateFullRedeploy
)
