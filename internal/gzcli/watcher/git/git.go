// Package git provides Git repository management for the watcher
package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/dimasma0305/gzcli/internal/log"
)

// Manager handles git operations for the watcher
type Manager struct {
	repoPath string
	interval time.Duration
	onUpdate func() // Callback to execute after successful pull
	ctx      context.Context
}

// NewManager creates a new git manager
func NewManager(repoPath string, interval time.Duration, onUpdate func()) *Manager {
	return &Manager{
		repoPath: repoPath,
		interval: interval,
		onUpdate: onUpdate,
	}
}

// StartPullLoop starts the periodic git pull loop
func (m *Manager) StartPullLoop(ctx context.Context) {
	m.ctx = ctx

	// Additional safeguard against zero duration
	interval := m.interval
	if interval <= 0 {
		interval = 1 * time.Minute
		log.Error("GitPullInterval was zero or negative in loop, using default: %v", interval)
	}

	// Initial pull on startup
	log.Info("🔄 Performing initial git pull...")
	if err := m.PerformPull(); err != nil {
		log.Error("Initial git pull failed: %v", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("Git pull loop stopped")
			return
		case <-ticker.C:
			if err := m.PerformPull(); err != nil {
				log.Error("Git pull failed: %v", err)
			}
		}
	}
}

// PerformPull performs a git pull operation on the specified repository
func (m *Manager) PerformPull() error {
	log.InfoH3("🔄 Pulling latest changes from git repository: %s", m.repoPath)

	// Only check for .git in the event root directory (no walking up the tree)
	root := m.repoPath
	gitDir := filepath.Join(m.repoPath, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return fmt.Errorf("no git repository found at %s (looking for .git in event root only): %w", m.repoPath, err)
	}

	// Execute system git pull (inherits env; uses current credentials/SSH config)
	//nolint:gosec // G204: Git repository path is validated and configured by user
	cmd := exec.Command("git", "-C", root, "pull")
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error("git pull failed: %v", err)
		if len(output) > 0 {
			log.Error("git output: %s", strings.TrimSpace(string(output)))
		}
		return fmt.Errorf("git pull failed: %w", err)
	}

	// Log concise success and any non-empty output
	out := strings.TrimSpace(string(output))
	switch {
	case out == "Already up to date.", strings.Contains(out, "Already up to date"):
		log.InfoH3("📄 Repository is already up-to-date")
	case out != "":
		log.InfoH3("✅ Git pull output:\n%s", out)
	default:
		log.InfoH3("✅ Git pull completed successfully")
	}

	// After successful pull, execute callback if provided
	if m.onUpdate != nil {
		log.InfoH3("🔍 Checking for updates after git pull...")
		m.onUpdate()
	}

	return nil
}
