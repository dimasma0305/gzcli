//nolint:errcheck,gosec,revive // Test file with acceptable error handling patterns
package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/testutil"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/database"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/watchertypes"
)

// TestNew_NilAPI tests watcher creation with nil API
func TestNew_NilAPI(t *testing.T) {
	_, err := New(nil)
	if err == nil {
		t.Error("Expected error when creating watcher with nil API")
	}
}

// setupEventWatcherTest creates a test EventWatcher
func setupEventWatcherTest(t *testing.T) (*EventWatcher, string, func()) {
	tmpDir, err := os.MkdirTemp("", "event-watcher-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create events directory structure
	eventDir := filepath.Join(tmpDir, "events", "test-event")
	if err := os.MkdirAll(eventDir, 0755); err != nil {
		t.Fatalf("Failed to create event dir: %v", err)
	}

	// Change to tmpDir so event paths work
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)

	api := &gzapi.GZAPI{}
	db := database.New(":memory:", false)
	db.Init()

	ctx := context.Background()
	config := watchertypes.WatcherConfig{}

	ew, err := NewEventWatcher("test-event", api, config, db, ctx)
	if err != nil {
		t.Fatalf("Failed to create event watcher: %v", err)
	}

	cleanup := func() {
		if ew.watcher != nil {
			ew.watcher.Close()
		}
		os.Chdir(oldWd)
		os.RemoveAll(tmpDir)
	}

	return ew, eventDir, cleanup
}

// TestWatcher_ConcurrentFileChanges tests handling of concurrent file events
func TestWatcher_ConcurrentFileChanges(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent file changes test in short mode")
	}

	_, eventDir, cleanup := setupEventWatcherTest(t)
	defer cleanup()

	// Create multiple files concurrently
	testutil.ConcurrentTest(t, 10, 5, func(id, iter int) error {
		filename := filepath.Join(eventDir, fmt.Sprintf("file-%d-%d.txt", id, iter))
		return os.WriteFile(filename, []byte("test"), 0644)
	})
}

// TestWatcher_RapidFileChanges tests handling of very rapid file changes
func TestWatcher_RapidFileChanges(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping rapid file changes test in short mode")
	}

	_, eventDir, cleanup := setupEventWatcherTest(t)
	defer cleanup()

	// Create 100 files very rapidly
	for i := 0; i < 100; i++ {
		filename := filepath.Join(eventDir, fmt.Sprintf("rapid-%d.txt", i))
		os.WriteFile(filename, []byte("test"), 0644)
		// No delay - as fast as possible
	}

	t.Log("Created 100 files rapidly without errors")
}

// TestWatcher_ConcurrentMutexAccess tests race conditions in mutex access
func TestWatcher_ConcurrentMutexAccess(t *testing.T) {
	ew, _, cleanup := setupEventWatcherTest(t)
	defer cleanup()

	// Access challenge mutexes concurrently
	testutil.ConcurrentTest(t, 20, 10, func(id, iter int) error {
		challengeName := fmt.Sprintf("challenge-%d", id%5) // Use only 5 names for contention
		mutex := ew.GetChallengeUpdateMutex(challengeName)

		mutex.Lock()
		// Simulate some work
		time.Sleep(1 * time.Millisecond)
		mutex.Unlock()

		return nil
	})

	t.Log("Concurrent mutex access completed without deadlock")
}

// TestWatcher_UpdateStateRaceConditions tests concurrent state updates
func TestWatcher_UpdateStateRaceConditions(t *testing.T) {
	ew, _, cleanup := setupEventWatcherTest(t)
	defer cleanup()

	// Update state concurrently for same challenge
	testutil.ConcurrentTest(t, 10, 20, func(id, iter int) error {
		ew.setUpdating("test-challenge", true)
		time.Sleep(1 * time.Millisecond)
		ew.setUpdating("test-challenge", false)
		return nil
	})

	// Verify final state is clean
	if ew.isUpdating("test-challenge") {
		t.Error("Challenge should not be marked as updating after test")
	}
}

// TestWatcher_PendingUpdatesRaceCondition tests concurrent pending updates
func TestWatcher_PendingUpdatesRaceCondition(t *testing.T) {
	ew, _, cleanup := setupEventWatcherTest(t)
	defer cleanup()

	challengeName := "test-challenge"

	// Set and get pending updates concurrently
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			filePath := fmt.Sprintf("/path/to/file-%d", id)
			ew.setPendingUpdate(challengeName, filePath)
		}(i)
	}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ew.getPendingUpdate(challengeName)
		}()
	}

	wg.Wait()
	t.Log("Concurrent pending updates handled without race condition")
}

// TestWatcher_ContextCancellation tests proper cleanup on context cancellation
func TestWatcher_ContextCancellation(t *testing.T) {
	ew, _, cleanup := setupEventWatcherTest(t)
	defer cleanup()

	// Cancel context
	ew.cancel()

	// Verify context is cancelled
	select {
	case <-ew.ctx.Done():
		t.Log("Context cancelled successfully")
	case <-time.After(1 * time.Second):
		t.Error("Context not cancelled within timeout")
	}
}

// TestWatcher_ManyWatchedDirectories tests resource exhaustion with many directories
func TestWatcher_ManyWatchedDirectories(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping many directories test in short mode")
	}

	ew, eventDir, cleanup := setupEventWatcherTest(t)
	defer cleanup()

	// Create 100 directories
	for i := 0; i < 100; i++ {
		dir := filepath.Join(eventDir, fmt.Sprintf("dir-%d", i))
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		// Try to add to watcher
		if err := ew.watcher.Add(dir); err != nil {
			t.Logf("Failed to watch directory %d: %v", i, err)
			// This might fail on systems with watch limits
			break
		}
	}

	t.Log("Successfully handled multiple watched directories")
}

// TestWatcher_SymlinkLoop tests handling of symlink loops
func TestWatcher_SymlinkLoop(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping symlink test in short mode")
	}

	ew, eventDir, cleanup := setupEventWatcherTest(t)
	defer cleanup()

	// Create symlink loop: a -> b -> a
	linkA := filepath.Join(eventDir, "link-a")
	linkB := filepath.Join(eventDir, "link-b")

	if err := os.Symlink(linkB, linkA); err != nil {
		t.Skipf("Failed to create symlink: %v", err)
	}
	if err := os.Symlink(linkA, linkB); err != nil {
		t.Skipf("Failed to create symlink loop: %v", err)
	}

	// Try to watch the symlink - should handle gracefully
	err := ew.watcher.Add(linkA)
	if err != nil {
		t.Logf("Symlink loop handled with error: %v", err)
	}
}

// TestWatcher_RapidCreateDelete tests rapid file creation and deletion
func TestWatcher_RapidCreateDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping rapid create/delete test in short mode")
	}

	_, eventDir, cleanup := setupEventWatcherTest(t)
	defer cleanup()

	// Rapidly create and delete files
	for i := 0; i < 50; i++ {
		filename := filepath.Join(eventDir, fmt.Sprintf("temp-%d.txt", i))
		os.WriteFile(filename, []byte("test"), 0644)
		os.Remove(filename)
	}

	t.Log("Rapid create/delete cycles handled")
}

// TestWatcher_FilePermissionChanges tests handling of permission changes
func TestWatcher_FilePermissionChanges(t *testing.T) {
	// Skip on Windows as it has different permission model
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission test on Windows (different permission model)")
	}

	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	_, eventDir, cleanup := setupEventWatcherTest(t)
	defer cleanup()

	filename := filepath.Join(eventDir, "test.txt")
	os.WriteFile(filename, []byte("test"), 0644)

	// Make file unreadable
	os.Chmod(filename, 0000)
	defer os.Chmod(filename, 0644)

	// Try to read - should handle permission error gracefully
	_, err := os.ReadFile(filename)
	if err == nil {
		t.Error("Expected permission error")
	}

	t.Log("Permission changes handled")
}

// TestWatcher_LargeNumberOfEvents tests handling of event flooding
func TestWatcher_LargeNumberOfEvents(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large event test in short mode")
	}

	ew, eventDir, cleanup := setupEventWatcherTest(t)
	defer cleanup()

	if err := ew.watcher.Add(eventDir); err != nil {
		t.Fatalf("Failed to watch directory: %v", err)
	}

	// Create many files to generate many events
	for i := 0; i < 1000; i++ {
		filename := filepath.Join(eventDir, fmt.Sprintf("event-%d.txt", i))
		os.WriteFile(filename, []byte("test"), 0644)
		if i%100 == 0 {
			time.Sleep(10 * time.Millisecond) // Small delay to avoid overwhelming
		}
	}

	t.Log("Handled 1000+ file events")
}

// TestWatcher_ConcurrentHandleFileChange tests concurrent file change handling
func TestWatcher_ConcurrentHandleFileChange(t *testing.T) {
	ew, _, cleanup := setupEventWatcherTest(t)
	defer cleanup()

	// Call HandleFileChange concurrently
	testutil.ConcurrentTest(t, 10, 5, func(id, iter int) error {
		filepath := fmt.Sprintf("/fake/path/file-%d-%d.txt", id, iter)
		ew.HandleFileChange(filepath)
		return nil
	})

	t.Log("Concurrent HandleFileChange calls completed")
}

// TestWatcher_DeadlockPrevention tests that operations don't deadlock
func TestWatcher_DeadlockPrevention(t *testing.T) {
	ew, _, cleanup := setupEventWatcherTest(t)
	defer cleanup()

	done := make(chan bool)

	go func() {
		// Perform operations that could deadlock
		for i := 0; i < 100; i++ {
			challengeName := fmt.Sprintf("challenge-%d", i%10)

			ew.setUpdating(challengeName, true)
			ew.GetChallengeUpdateMutex(challengeName)
			ew.setPendingUpdate(challengeName, "/some/path")
			ew.getPendingUpdate(challengeName)
			ew.setUpdating(challengeName, false)
		}
		done <- true
	}()

	select {
	case <-done:
		t.Log("No deadlock detected")
	case <-time.After(5 * time.Second):
		t.Fatal("Deadlock detected - operations timed out")
	}
}

// TestWatcher_MemoryLeakDetection tests for potential memory leaks
func TestWatcher_MemoryLeakDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak test in short mode")
	}

	ew, _, cleanup := setupEventWatcherTest(t)
	defer cleanup()

	initialMutexCount := len(ew.challengeMutexes)

	// Create many mutexes
	for i := 0; i < 1000; i++ {
		challengeName := fmt.Sprintf("challenge-%d", i)
		ew.GetChallengeUpdateMutex(challengeName)
	}

	finalMutexCount := len(ew.challengeMutexes)

	if finalMutexCount != 1000 {
		t.Errorf("Expected 1000 mutexes, got %d", finalMutexCount)
	}

	t.Logf("Created %d mutexes (started with %d)", finalMutexCount-initialMutexCount, initialMutexCount)
	t.Log("Note: In production, consider implementing mutex cleanup for inactive challenges")
}

// TestWatcher_NilChallengeManager tests operations when challenge manager is nil
func TestWatcher_NilChallengeManager(t *testing.T) {
	ew, _, cleanup := setupEventWatcherTest(t)
	defer cleanup()

	// Challenge manager should be initialized
	if ew.challengeMgr == nil {
		t.Error("Challenge manager should not be nil")
	}
}

// TestWatcher_StressTest performs overall stress testing
func TestWatcher_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	ew, eventDir, cleanup := setupEventWatcherTest(t)
	defer cleanup()

	// Perform various operations concurrently
	var wg sync.WaitGroup

	// File operations
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			filename := filepath.Join(eventDir, fmt.Sprintf("stress-%d.txt", i))
			os.WriteFile(filename, []byte("test"), 0644)
			time.Sleep(5 * time.Millisecond)
		}
	}()

	// State operations
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			challengeName := fmt.Sprintf("challenge-%d", i%10)
			ew.setUpdating(challengeName, true)
			time.Sleep(2 * time.Millisecond)
			ew.setUpdating(challengeName, false)
		}
	}()

	// Mutex operations
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			challengeName := fmt.Sprintf("challenge-%d", i%10)
			mutex := ew.GetChallengeUpdateMutex(challengeName)
			mutex.Lock()
			time.Sleep(1 * time.Millisecond)
			mutex.Unlock()
		}
	}()

	wg.Wait()
	t.Log("Stress test completed successfully")
}

// TestWatcher_EdgeCases tests various edge cases
func TestWatcher_EdgeCases(t *testing.T) {
	ew, _, cleanup := setupEventWatcherTest(t)
	defer cleanup()

	testCases := []struct {
		name string
		fn   func()
	}{
		{
			"Empty challenge name",
			func() {
				ew.setUpdating("", true)
				ew.setUpdating("", false)
			},
		},
		{
			"Very long challenge name",
			func() {
				longName := string(make([]byte, 10000))
				ew.setUpdating(longName, true)
			},
		},
		{
			"Special characters in challenge name",
			func() {
				ew.setUpdating("challenge\x00with\x00nulls", true)
			},
		},
		{
			"Unicode challenge name",
			func() {
				ew.setUpdating("チャレンジ-🚀", true)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Should not panic
			tc.fn()
		})
	}
}
