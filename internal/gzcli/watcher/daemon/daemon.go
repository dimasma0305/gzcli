// Package daemon provides daemon process management for the watcher
package daemon

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/dimasma0305/gzcli/internal/log"
)

// GetDaemonStatus returns the status of the daemon watcher
func GetDaemonStatus(pidFile string) map[string]interface{} {
	status := map[string]interface{}{
		"daemon":   false,
		"pid_file": pidFile,
	}

	pid, err := ReadPIDFromFile(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			status["status"] = "stopped"
			status["message"] = "PID file not found"
		} else {
			status["status"] = "error"
			status["message"] = err.Error()
		}
		return status
	}

	status["pid"] = pid

	// Check if process is running
	process, err := os.FindProcess(pid)
	if err != nil {
		status["status"] = "error"
		status["message"] = fmt.Sprintf("Failed to find process: %v", err)
		return status
	}

	// Send signal 0 to check if process exists
	if err := process.Signal(syscall.Signal(0)); err != nil {
		status["daemon"] = false
		status["status"] = "dead"
		status["message"] = "Process not running (stale PID file)"
		// Clean up stale PID file
		if removeErr := os.Remove(pidFile); removeErr != nil && !os.IsNotExist(removeErr) {
			status["message"] = fmt.Sprintf("Process not running, failed to clean stale PID file: %v", removeErr)
		} else {
			status["message"] = "Process not running (cleaned up stale PID file)"
		}
		return status
	}

	status["daemon"] = true
	status["status"] = "running"
	status["message"] = "Daemon is running"
	return status
}

// StopDaemon stops the daemon watcher
func StopDaemon(pidFile string) error {
	// Read PID from file
	pid, err := ReadPIDFromFile(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("daemon is not running (PID file not found)")
		}
		return err
	}

	// Find the process
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	// Send SIGTERM first
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM to process %d: %w", pid, err)
	}

	// Wait a bit for graceful shutdown
	time.Sleep(2 * time.Second)

	// Check if process is still running
	if err := process.Signal(syscall.Signal(0)); err == nil {
		// Process is still running, send SIGKILL
		log.Info("Process still running, sending SIGKILL...")
		if err := process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process %d: %w", pid, err)
		}
	}

	// Clean up PID file
	if err := os.Remove(pidFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove PID file: %w", err)
	}

	log.Info("✅ GZCTF Watcher daemon stopped successfully")
	return nil
}
