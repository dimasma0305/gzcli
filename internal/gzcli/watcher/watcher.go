// Package watcher provides file system watching and automatic challenge synchronization for GZCTF
package watcher

import (
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/core"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/socket"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/types"
)

// Re-export core types for backward compatibility
type (
	// Watcher is the main watcher instance
	Watcher = core.Watcher

	// WatcherConfig holds configuration for the watcher
	WatcherConfig = types.WatcherConfig

	// WatcherCommand represents commands sent to the watcher via socket
	WatcherCommand = types.WatcherCommand

	// WatcherResponse represents responses from the watcher
	WatcherResponse = types.WatcherResponse

	// ScriptMetrics tracks execution statistics for scripts
	ScriptMetrics = types.ScriptMetrics

	// WatcherLog represents a log entry
	WatcherLog = types.WatcherLog

	// ChallengeState represents challenge state
	ChallengeState = types.ChallengeState

	// ScriptExecution represents script execution record
	ScriptExecution = types.ScriptExecution

	// UpdateType represents the type of update needed
	UpdateType = types.UpdateType

	// WatcherClient provides client interface for the watcher daemon
	WatcherClient = socket.Client
)

// Re-export constants
const (
	UpdateNone         = types.UpdateNone
	UpdateAttachment   = types.UpdateAttachment
	UpdateMetadata     = types.UpdateMetadata
	UpdateFullRedeploy = types.UpdateFullRedeploy
)

// Re-export default configuration
var DefaultWatcherConfig = types.DefaultWatcherConfig

// NewWatcher creates a new file watcher instance
func NewWatcher(api *gzapi.GZAPI) (*Watcher, error) {
	return core.New(api)
}

// NewWatcherClient creates a new watcher client for communicating with the daemon
func NewWatcherClient(socketPath string) *WatcherClient {
	return socket.NewClient(socketPath)
}
