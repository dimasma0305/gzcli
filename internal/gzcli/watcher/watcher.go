// Package watcher provides file system watching and automatic challenge synchronization for GZCTF
//
//nolint:revive // Watcher type names kept for backward compatibility
package watcher

import (
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/core"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/socket"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/watchertypes"
)

// Re-export core types for backward compatibility
type (
	// Watcher is the main watcher instance
	Watcher = core.Watcher

	// WatcherConfig holds configuration for the watcher
	WatcherConfig = watchertypes.WatcherConfig

	// WatcherCommand represents commands sent to the watcher via socket
	WatcherCommand = watchertypes.WatcherCommand

	// WatcherResponse represents responses from the watcher
	WatcherResponse = watchertypes.WatcherResponse

	// ScriptMetrics tracks execution statistics for scripts
	ScriptMetrics = watchertypes.ScriptMetrics

	// WatcherLog represents a log entry
	WatcherLog = watchertypes.WatcherLog

	// ChallengeState represents challenge state
	ChallengeState = watchertypes.ChallengeState

	// ScriptExecution represents script execution record
	ScriptExecution = watchertypes.ScriptExecution

	// UpdateType represents the type of update needed
	UpdateType = watchertypes.UpdateType

	// WatcherClient provides client interface for the watcher daemon
	WatcherClient = socket.Client
)

// Re-export constants
const (
	UpdateNone         = watchertypes.UpdateNone
	UpdateAttachment   = watchertypes.UpdateAttachment
	UpdateMetadata     = watchertypes.UpdateMetadata
	UpdateFullRedeploy = watchertypes.UpdateFullRedeploy
)

// Re-export default configuration
var DefaultWatcherConfig = watchertypes.DefaultWatcherConfig

// NewWatcher creates a new file watcher instance
func NewWatcher(api *gzapi.GZAPI) (*Watcher, error) {
	return core.New(api)
}

// NewWatcherClient creates a new watcher client for communicating with the daemon
func NewWatcherClient(socketPath string) *WatcherClient {
	return socket.NewClient(socketPath)
}
