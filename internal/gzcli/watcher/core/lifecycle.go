package core

import (
	"fmt"
	"os"
	"time"

	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/daemon"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/database"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/filesystem"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/git"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/socket"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/types"
	"github.com/dimasma0305/gzcli/internal/log"
	godaemon "github.com/sevlyar/go-daemon"
)

// Start starts the file watcher with the given configuration
func (w *Watcher) Start(config types.WatcherConfig) error {
	w.config = config

	// Validate and set defaults
	if w.config.NewChallengeCheckInterval <= 0 {
		w.config.NewChallengeCheckInterval = types.DefaultWatcherConfig.NewChallengeCheckInterval
	}
	if w.config.PidFile == "" {
		w.config.PidFile = types.DefaultWatcherConfig.PidFile
	}
	if w.config.LogFile == "" {
		w.config.LogFile = types.DefaultWatcherConfig.LogFile
	}
	if w.config.DatabasePath == "" {
		w.config.DatabasePath = types.DefaultWatcherConfig.DatabasePath
	}
	if w.config.SocketPath == "" {
		w.config.SocketPath = types.DefaultWatcherConfig.SocketPath
	}

	if w.config.DaemonMode {
		log.Info("Starting file watcher in DAEMON mode...")
		return w.startAsDaemon()
	}

	log.Info("Starting file watcher in foreground mode...")
	return w.startWatcher()
}

// startAsDaemon starts the watcher as a daemon process
func (w *Watcher) startAsDaemon() error {
	// Create daemon context
	daemonCtx := &godaemon.Context{
		PidFileName: w.config.PidFile,
		PidFilePerm: 0644,
		LogFileName: w.config.LogFile,
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
	}

	// Check if we're already in the daemon process
	if godaemon.WasReborn() {
		// This is the child daemon process
		pid := os.Getpid()
		log.Info("🚀 GZCTF Watcher daemon started (PID: %d)", pid)
		log.Info("📄 PID file: %s", w.config.PidFile)
		log.Info("📝 Log file: %s", w.config.LogFile)

		// Write PID file
		if err := daemon.WritePIDFile(w.config.PidFile, pid); err != nil {
			log.Error("Failed to write PID file: %v", err)
			return fmt.Errorf("failed to write PID file: %w", err)
		}

		// Start the actual watcher
		if err := w.startWatcher(); err != nil {
			return err
		}

		// Keep daemon running until context is cancelled
		<-w.ctx.Done()
		return nil
	}

	// This is the parent process - fork the daemon
	child, err := daemonCtx.Reborn()
	if err != nil {
		return fmt.Errorf("failed to fork daemon: %w", err)
	}

	if child != nil {
		// Parent process - daemon started successfully
		log.Info("✅ GZCTF Watcher daemon started successfully")
		log.Info("📄 PID: %d (saved to %s)", child.Pid, w.config.PidFile)
		log.Info("📝 Logs: %s", w.config.LogFile)
		return nil
	}

	return fmt.Errorf("unexpected daemon state")
}

// startWatcher starts the actual watcher functionality
func (w *Watcher) startWatcher() error {
	// Initialize database
	w.db = database.New(w.config.DatabasePath, w.config.DatabaseEnabled)
	if err := w.db.Init(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize socket server
	socketHandler := socket.NewDefaultCommandHandler(w)
	w.socketServer = socket.NewServer(w.config.SocketPath, w.config.SocketEnabled, socketHandler)
	if err := w.socketServer.Init(); err != nil {
		return fmt.Errorf("failed to initialize socket server: %w", err)
	}

	// Initialize git manager
	if w.config.GitPullEnabled {
		w.gitMgr = git.NewManager(w.config.GitRepository, w.config.GitPullInterval, func() {
			log.Info("Git pull completed, checking for new challenges...")
		})
	}

	// Start file system watcher loop
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		done := make(chan struct{})
		go func() {
			<-w.ctx.Done()
			close(done)
		}()
		filesystem.WatchLoop(w.watcher, w.config, w, done)
	}()

	// Start socket server if enabled
	if w.config.SocketEnabled && w.socketServer != nil {
		w.wg.Add(1)
		go func() {
			defer w.wg.Done()
			w.socketServer.Run(w.ctx)
		}()
	}

	// Start git pull loop if enabled
	if w.config.GitPullEnabled && w.gitMgr != nil {
		w.wg.Add(1)
		go func() {
			defer w.wg.Done()
			w.gitMgr.StartPullLoop(w.ctx)
		}()
	}

	log.Info("File watcher started successfully")
	w.LogToDatabase("INFO", "watcher", "", "", "File watcher started", "", 0)

	return nil
}

// Stop stops the file watcher with graceful shutdown
func (w *Watcher) Stop() error {
	log.Info("Stopping file watcher...")

	w.LogToDatabase("INFO", "watcher", "", "", "File watcher shutdown initiated", "", 0)

	// Stop all interval scripts
	if w.scriptMgr != nil {
		w.scriptMgr.StopAllScripts(5 * time.Second)
	}

	// Cancel context
	w.cancel()

	// Wait for all goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.InfoH3("All goroutines finished successfully")
	case <-time.After(10 * time.Second):
		log.Error("Timeout waiting for goroutines to finish")
	}

	// Close socket server
	if w.socketServer != nil {
		if err := w.socketServer.Close(); err != nil {
			log.Error("Failed to close socket server: %v", err)
		}
	}

	// Close file system watcher
	if w.watcher != nil {
		if err := w.watcher.Close(); err != nil {
			log.Error("Failed to close file watcher: %v", err)
		}
	}

	w.LogToDatabase("INFO", "watcher", "", "", "File watcher shutdown completed", "", 0)

	// Close database last
	if w.db != nil {
		if err := w.db.Close(); err != nil {
			log.Error("Failed to close database: %v", err)
		}
	}

	log.Info("File watcher stopped")
	return nil
}

// IsWatching returns true if the watcher is currently active
func (w *Watcher) IsWatching() bool {
	select {
	case <-w.ctx.Done():
		return false
	default:
		return true
	}
}

// GetWatchedChallenges returns the list of currently watched challenges
func (w *Watcher) GetWatchedChallenges() []string {
	challenges := w.challengeMgr.GetChallenges()
	dirs := make([]string, 0, len(challenges))
	for _, dir := range challenges {
		dirs = append(dirs, dir)
	}
	return dirs
}

// GetDaemonStatus returns the status of the daemon watcher
func (w *Watcher) GetDaemonStatus(pidFile string) map[string]interface{} {
	if pidFile == "" {
		pidFile = types.DefaultWatcherConfig.PidFile
	}
	return daemon.GetDaemonStatus(pidFile)
}

// StopDaemon stops the daemon watcher
func (w *Watcher) StopDaemon(pidFile string) error {
	if pidFile == "" {
		pidFile = types.DefaultWatcherConfig.PidFile
	}
	return daemon.StopDaemon(pidFile)
}

// ShowStatus displays the watcher status
func (w *Watcher) ShowStatus(pidFile, logFile string, jsonOutput bool) error {
	if pidFile == "" {
		pidFile = types.DefaultWatcherConfig.PidFile
	}
	if logFile == "" {
		logFile = types.DefaultWatcherConfig.LogFile
	}
	return daemon.ShowStatus(pidFile, logFile, jsonOutput)
}

// FollowLogs follows the daemon log file
func (w *Watcher) FollowLogs(logFile string) error {
	if logFile == "" {
		logFile = types.DefaultWatcherConfig.LogFile
	}
	return daemon.FollowLogs(logFile)
}
