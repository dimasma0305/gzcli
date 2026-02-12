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
	oldHead, oldHeadErr := m.getHeadSHA(root)
	if oldHeadErr != nil {
		log.Debug("Unable to read HEAD before pull in %s: %v", root, oldHeadErr)
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

	newHead, newHeadErr := m.getHeadSHA(root)
	if newHeadErr != nil {
		log.Debug("Unable to read HEAD after pull in %s: %v", root, newHeadErr)
	}

	headChanged := oldHead != "" && newHead != "" && oldHead != newHead
	if !headChanged {
		log.InfoH3("📄 No new commits pulled; skipping post-pull sync callback")
		return nil
	}

	// After successful pull with new commits, execute callback if provided.
	if m.onUpdate != nil {
		log.InfoH3("🔍 Checking for updates after git pull...")
		m.onUpdate()
	}

	return nil
}

func (m *Manager) getHeadSHA(root string) (string, error) {
	cmd := exec.Command("git", "-C", root, "rev-parse", "HEAD")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git rev-parse HEAD failed: %w (%s)", err, strings.TrimSpace(string(out)))
	}

	return strings.TrimSpace(string(out)), nil
}

// ResolveRepoPaths attempts to find git repositories in the following order:
// 1. Current working directory (returns single path)
// 2. ./events directory (returns single path)
// 3. ./events/[eventName] directory (returns single path)
// 4. Subdirectories of ./events/[eventName] (can return multiple paths)
// Returns a list of paths to directories containing .git.
func ResolveRepoPaths(cwd, eventName string) ([]string, error) {
	// 1. Check current directory
	if isGitRepo(cwd) {
		log.Debug("Found git repo at current directory: %s", cwd)
		return []string{cwd}, nil
	}

	// 2. Check ./events
	eventsPath := filepath.Join(cwd, "events")
	if isGitRepo(eventsPath) {
		log.Debug("Found git repo at events directory: %s", eventsPath)
		return []string{eventsPath}, nil
	}

	// 3. Check ./events/[eventName]
	if eventName != "" {
		eventPath := filepath.Join(eventsPath, eventName)

		// If the event directory itself is a git repo
		if isGitRepo(eventPath) {
			log.Debug("Found git repo at specific event directory: %s", eventPath)
			return []string{eventPath}, nil
		}

		// 4. Scan subdirectories of ./events/[eventName]
		entries, err := os.ReadDir(eventPath)
		if err == nil {
			var repos []string
			for _, entry := range entries {
				if entry.IsDir() {
					subPath := filepath.Join(eventPath, entry.Name())
					if isGitRepo(subPath) {
						log.Debug("Found git repo at event subdirectory: %s", subPath)
						repos = append(repos, subPath)
					}
				}
			}
			if len(repos) > 0 {
				return repos, nil
			}
		}
	}

	return nil, fmt.Errorf("no git repositories found in search paths")
}

// isGitRepo checks if a specific directory contains a .git subdirectory
func isGitRepo(path string) bool {
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}
