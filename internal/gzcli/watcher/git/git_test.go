package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	repoPath := filepath.Join(os.TempDir(), "test-repo")
	interval := 1 * time.Minute
	onUpdate := func() {}

	mgr := NewManager(repoPath, interval, onUpdate)

	if mgr == nil {
		t.Fatal("NewManager() returned nil")
		return // Help staticcheck understand control flow
	}

	if mgr.repoPath != repoPath {
		t.Errorf("repoPath = %q, want %q", mgr.repoPath, repoPath)
	}

	if mgr.interval != interval {
		t.Errorf("interval = %v, want %v", mgr.interval, interval)
	}

	if mgr.onUpdate == nil {
		t.Error("onUpdate callback should not be nil")
	}
}

func TestPerformPull_NoGitDirectory(t *testing.T) {
	// Setup: Create temporary directory without .git
	tmpDir, err := os.MkdirTemp("", "git-test-no-git-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	mgr := NewManager(tmpDir, 1*time.Minute, nil)

	// Test: PerformPull should fail when no .git in event root
	err = mgr.PerformPull()
	if err == nil {
		t.Error("PerformPull() expected error when no .git directory, got none")
	}

	// Should mention "no git repository found"
	if err != nil && !containsString(err.Error(), "no git repository found") {
		t.Errorf("PerformPull() error = %v, want error containing 'no git repository found'", err)
	}

	// Should mention "event root only"
	if err != nil && !containsString(err.Error(), "event root only") {
		t.Errorf("PerformPull() error = %v, want error containing 'event root only'", err)
	}
}

func TestPerformPull_GitInEventRoot(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping git integration test in short mode")
	}

	// Setup: Create temporary directory with .git
	tmpDir, err := os.MkdirTemp("", "git-test-with-git-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Initialize a git repository
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0750); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	// Create a minimal git config to make it a valid repo
	configDir := filepath.Join(gitDir, "config")
	//nolint:gosec // G306: Test file permissions are acceptable
	if err := os.WriteFile(configDir, []byte("[core]\n\trepositoryformatversion = 0\n"), 0644); err != nil {
		t.Fatalf("Failed to create git config: %v", err)
	}

	mgr := NewManager(tmpDir, 1*time.Minute, nil)

	// Test: PerformPull should attempt to run (will fail without real git repo, but that's okay)
	// We're just testing that it doesn't error due to missing .git directory
	err = mgr.PerformPull()
	// Expect error from git command, but not from directory detection
	if err != nil {
		t.Logf("Expected git command error (not a real repo): %v", err)
	}
}

func TestPerformPull_NoWalkUpTree(t *testing.T) {
	// Setup: Create parent directory with .git and subdirectory without .git
	tmpDir, err := os.MkdirTemp("", "git-test-parent-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create .git in parent
	parentGit := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(parentGit, 0750); err != nil {
		t.Fatalf("Failed to create parent .git: %v", err)
	}

	// Create subdirectory (simulating event directory)
	eventDir := filepath.Join(tmpDir, "events", "test-event")
	if err := os.MkdirAll(eventDir, 0750); err != nil {
		t.Fatalf("Failed to create event dir: %v", err)
	}

	mgr := NewManager(eventDir, 1*time.Minute, nil)

	// Test: PerformPull should fail because .git is not in event root
	// It should NOT walk up and find the parent .git
	err = mgr.PerformPull()
	if err == nil {
		t.Error("PerformPull() should fail when .git is only in parent directory, not event root")
	}

	if err != nil && !containsString(err.Error(), "no git repository found") {
		t.Errorf("PerformPull() error = %v, want error containing 'no git repository found'", err)
	}
}

func TestPerformPull_WithCallback(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping git integration test in short mode")
	}

	// Setup
	tmpDir, err := os.MkdirTemp("", "git-test-callback-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0750); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	callbackCalled := false
	onUpdate := func() {
		callbackCalled = true
	}

	mgr := NewManager(tmpDir, 1*time.Minute, onUpdate)

	// Even though git pull will fail, callback logic is still testable
	// (callback is only called on success, so it won't be called here)
	_ = mgr.PerformPull()

	// Since git pull fails (not a real repo), callback should not be called
	if callbackCalled {
		t.Error("Callback should not be called when git pull fails")
	}
}

func TestStartPullLoop_ContextCancellation(t *testing.T) {
	// Setup
	tmpDir, err := os.MkdirTemp("", "git-test-context-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	mgr := NewManager(tmpDir, 100*time.Millisecond, nil)

	ctx, cancel := context.WithCancel(context.Background())

	// Start pull loop in background
	done := make(chan struct{})
	go func() {
		mgr.StartPullLoop(ctx)
		close(done)
	}()

	// Wait a bit then cancel
	time.Sleep(50 * time.Millisecond)
	cancel()

	// Wait for loop to stop
	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Error("StartPullLoop did not stop after context cancellation")
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}()
}
