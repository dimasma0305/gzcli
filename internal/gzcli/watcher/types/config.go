package types

import (
	"time"
)

// WatcherConfig holds configuration for the watcher
type WatcherConfig struct {
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
	PidFile:                   "/tmp/gzctf-watcher.pid",
	LogFile:                   "/tmp/gzctf-watcher.log",
	GitPullEnabled:            true,            // Enable git pull by default
	GitPullInterval:           1 * time.Minute, // Pull every minute
	GitRepository:             ".",             // Current directory
	// Database defaults
	DatabaseEnabled: true, // Enable database logging by default
	DatabasePath:    "/tmp/gzctf-watcher.db",
	// Socket defaults
	SocketEnabled: true, // Enable socket server by default
	SocketPath:    "/tmp/gzctf-watcher.sock",
}
