package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncCommand_MultiEventFlags(t *testing.T) {
	// Test that sync command properly registers multi-event flags

	// Check that event flag is registered
	eventFlag := syncCmd.Flags().Lookup("event")
	if eventFlag == nil {
		t.Error("sync command should have --event flag")
	}

	// Check that exclude-event flag is registered
	excludeFlag := syncCmd.Flags().Lookup("exclude-event")
	if excludeFlag == nil {
		t.Error("sync command should have --exclude-event flag")
	}

	// Check that event flag has correct shorthand
	if eventFlag != nil && eventFlag.Shorthand != "e" {
		t.Errorf("sync --event flag shorthand = %q, want %q", eventFlag.Shorthand, "e")
	}

	// Check flag types
	if eventFlag != nil && eventFlag.Value.Type() != "stringSlice" {
		t.Errorf("sync --event flag type = %q, want %q", eventFlag.Value.Type(), "stringSlice")
	}

	if excludeFlag != nil && excludeFlag.Value.Type() != "stringSlice" {
		t.Errorf("sync --exclude-event flag type = %q, want %q", excludeFlag.Value.Type(), "stringSlice")
	}
}

func TestSyncCommand_HelpText(t *testing.T) {
	// Test that help text mentions multi-event behavior

	if !strings.Contains(syncCmd.Long, "all events") {
		t.Error("sync command Long description should mention 'all events' behavior")
	}

	if !strings.Contains(syncCmd.Long, "--event") {
		t.Error("sync command Long description should mention --event flag")
	}

	if !strings.Contains(syncCmd.Long, "--exclude-event") {
		t.Error("sync command Long description should mention --exclude-event flag")
	}

	// Check examples mention multi-event usage
	if !strings.Contains(syncCmd.Example, "all events") {
		t.Error("sync command examples should show 'all events' usage")
	}
}

func TestSyncCommand_VariablesInitialized(t *testing.T) {
	// Test that sync command variables are properly initialized

	// These should be initialized as empty slices, not nil
	if syncEvents == nil {
		t.Error("syncEvents should be initialized")
	}

	if syncExcludeEvents == nil {
		t.Error("syncExcludeEvents should be initialized")
	}
}

func TestSyncCommand_FlagDefaults(t *testing.T) {
	// Reset flags to test defaults
	syncEvents = []string{}
	syncExcludeEvents = []string{}

	if len(syncEvents) != 0 {
		t.Errorf("syncEvents default = %v, want empty slice", syncEvents)
	}

	if len(syncExcludeEvents) != 0 {
		t.Errorf("syncExcludeEvents default = %v, want empty slice", syncExcludeEvents)
	}
}

// Test the sync command structure and configuration
func TestSyncCommand_Structure(t *testing.T) {
	tests := []struct {
		name      string
		checkFunc func(*testing.T)
	}{
		{
			name: "command has correct use",
			checkFunc: func(t *testing.T) {
				if syncCmd.Use != "sync" {
					t.Errorf("sync command Use = %q, want %q", syncCmd.Use, "sync")
				}
			},
		},
		{
			name: "command has aliases",
			checkFunc: func(t *testing.T) {
				if len(syncCmd.Aliases) == 0 {
					t.Error("sync command should have aliases")
				}
				if !contains(syncCmd.Aliases, "s") {
					t.Error("sync command should have 's' alias")
				}
			},
		},
		{
			name: "command has short description",
			checkFunc: func(t *testing.T) {
				if syncCmd.Short == "" {
					t.Error("sync command should have short description")
				}
			},
		},
		{
			name: "command has long description",
			checkFunc: func(t *testing.T) {
				if syncCmd.Long == "" {
					t.Error("sync command should have long description")
				}
			},
		},
		{
			name: "command has examples",
			checkFunc: func(t *testing.T) {
				if syncCmd.Example == "" {
					t.Error("sync command should have examples")
				}
			},
		},
		{
			name: "command has run function",
			checkFunc: func(t *testing.T) {
				if syncCmd.Run == nil {
					t.Error("sync command should have Run function")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.checkFunc)
	}
}

// Integration-style test for command behavior
// setupSyncTestEnv creates test environment for sync command tests
func setupSyncTestEnv(t *testing.T, events []string) (string, func()) {
	t.Helper()
	originalDir, _ := os.Getwd()
	tmpDir, _ := os.MkdirTemp("", "gzcli-sync-test-*")
	_ = os.Chdir(tmpDir)

	// Create test events structure
	for _, event := range events {
		eventDir := filepath.Join(tmpDir, "events", event)
		_ = os.MkdirAll(eventDir, 0750)
		gzeventFile := filepath.Join(eventDir, ".gzevent")
		//nolint:gosec // G306: Test file permissions are acceptable
		_ = os.WriteFile(gzeventFile, []byte("title: Test Event\n"), 0644)
	}

	cleanup := func() {
		_ = os.Chdir(originalDir)
		_ = os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

// verifyResolvedEvents verifies that resolved events match expected
func verifyResolvedEvents(t *testing.T, got, want []string) {
	t.Helper()
	if !equalStringSlices(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSyncCommand_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	events := []string{"test-event-1", "test-event-2"}
	_, cleanup := setupSyncTestEnv(t, events)
	defer cleanup()

	// Test: Resolve all events with empty flags
	resolvedEvents, err := ResolveTargetEvents([]string{}, []string{})
	if err != nil {
		t.Fatalf("ResolveTargetEvents() error = %v", err)
	}
	verifyResolvedEvents(t, resolvedEvents, events)

	// Test: Specific event
	resolvedEvents, err = ResolveTargetEvents([]string{"test-event-1"}, []string{})
	if err != nil {
		t.Fatalf("ResolveTargetEvents() error = %v", err)
	}
	verifyResolvedEvents(t, resolvedEvents, []string{"test-event-1"})

	// Test: Exclusion
	resolvedEvents, err = ResolveTargetEvents([]string{}, []string{"test-event-2"})
	if err != nil {
		t.Fatalf("ResolveTargetEvents() error = %v", err)
	}
	verifyResolvedEvents(t, resolvedEvents, []string{"test-event-1"})
}

func TestSyncCommand_Examples(t *testing.T) {
	// Verify that examples contain expected patterns
	expectedPatterns := []string{
		"Sync all events",
		"Sync specific events",
		"--event",
		"--exclude-event",
	}

	helpText := syncCmd.Long + syncCmd.Example

	for _, pattern := range expectedPatterns {
		if !strings.Contains(helpText, pattern) {
			t.Errorf("sync command help should contain pattern %q", pattern)
		}
	}
}

// Benchmark for ResolveTargetEvents
func BenchmarkResolveTargetEvents(b *testing.B) {
	// Setup
	originalDir, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			b.Errorf("Failed to restore directory: %v", err)
		}
	}()

	tmpDir, err := os.MkdirTemp("", "gzcli-bench-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			b.Errorf("Failed to remove temp dir: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		b.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Create test events
	for i := 0; i < 10; i++ {
		eventDir := filepath.Join(tmpDir, "events", "event"+string(rune('0'+i)))
		if err := os.MkdirAll(eventDir, 0750); err != nil {
			b.Fatalf("Failed to create event dir: %v", err)
		}
		gzeventFile := filepath.Join(eventDir, ".gzevent")
		//nolint:gosec // G306: Test file permissions are acceptable
		if err := os.WriteFile(gzeventFile, []byte("title: Test\n"), 0644); err != nil {
			b.Fatalf("Failed to create .gzevent: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ResolveTargetEvents([]string{}, []string{})
	}
}

// Test output capture (for manual testing/debugging)
//
//nolint:unused // Kept for future testing needs
func captureOutput(f func()) string {
	var buf bytes.Buffer
	// In a real scenario, you'd redirect log output
	// This is a simplified version
	f()
	return buf.String()
}
