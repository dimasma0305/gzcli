package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dimasma0305/gzcli/internal/log"
)

// EnsureDirectoriesExist ensures that the directories for the given file paths exist
func EnsureDirectoriesExist(paths ...string) error {
	for _, path := range paths {
		if path == "" {
			continue
		}
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0750); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

// WritePIDFile writes the PID to the specified file
func WritePIDFile(pidFile string, pid int) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(pidFile)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create PID file directory: %w", err)
	}

	// Write PID to file
	pidStr := fmt.Sprintf("%d\n", pid)
	if err := os.WriteFile(pidFile, []byte(pidStr), 0600); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	log.Info("✅ PID file written successfully: %s", pidFile)
	return nil
}

// ReadPIDFromFile reads a PID integer from the given pid file.
// Returns os.ErrNotExist if the file does not exist, or a formatted error for invalid/empty PID content.
func ReadPIDFromFile(pidFile string) (int, error) {
	//nolint:gosec // G304: PID file path is constructed by application
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, err
	}
	pidStr := strings.TrimSpace(string(data))
	if pidStr == "" {
		return 0, fmt.Errorf("PID file is empty")
	}
	var pid int
	if _, err := fmt.Sscanf(pidStr, "%d", &pid); err != nil {
		return 0, fmt.Errorf("invalid PID in file: %w", err)
	}
	return pid, nil
}
