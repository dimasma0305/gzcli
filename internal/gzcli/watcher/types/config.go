package types

import (
	"time"
)

// WatcherConfig holds configuration for the watcher
type WatcherConfig struct {
	Events                    []string // Event names to watch (empty means use current event)
	PollInterval              time.Duration
	DebounceTime              time.Duration
	IgnorePatterns            []string
	WatchPatterns             []string
	NewChallengeCheckInterval time.Duration // New field for checking new challenges
	DaemonMode                bool          // Run watcher as daemon
	PidFile                   string        // PID file location
	LogFile                   string        // Log file location
	GitPullEnabled            bool          // Enable automatic git pull
	GitPullInterval           time.Duration // Interval for git pull (default: 1 minute)
	GitRepository             string        // Git repository path (default: current directory)
	// Database configuration
	DatabaseEnabled bool   // Enable database logging
	DatabasePath    string // SQLite database file path
	// Socket configuration
	SocketEnabled bool   // Enable socket server
	SocketPath    string // Unix socket path for communication
}

// DefaultWatcherConfig provides default configuration values
var DefaultWatcherConfig = WatcherConfig{
	PollInterval:              5 * time.Second,
	DebounceTime:              2 * time.Second,
	IgnorePatterns:            []string{},       // No ignore patterns
	WatchPatterns:             []string{},       // Empty means watch all files
	NewChallengeCheckInterval: 10 * time.Second, // Check for new challenges every 10 seconds
	DaemonMode:                true,             // Default to daemon mode
	PidFile:                   ".gzcli/watcher/watcher.pid",
	LogFile:                   ".gzcli/watcher/watcher.log",
	GitPullEnabled:            true,            // Enable git pull by default
	GitPullInterval:           1 * time.Minute, // Pull every minute
	GitRepository:             ".",             // Current directory
	// Database defaults
	DatabaseEnabled: true, // Enable database logging by default
	DatabasePath:    ".gzcli/watcher/watcher.db",
	// Socket defaults
	SocketEnabled: true, // Enable socket server by default
	SocketPath:    ".gzcli/watcher/watcher.sock",
}
