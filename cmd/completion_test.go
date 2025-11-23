package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// TestGetAvailableEvents tests event discovery for completion
func TestGetAvailableEvents(t *testing.T) {
	// Create temporary directory structure
	tmpDir, err := os.MkdirTemp("", "completion-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tmpDir)

	// Create events directory with test events
	eventsDir := filepath.Join(tmpDir, "events")
	_ = os.MkdirAll(eventsDir, 0750)

	// Create event1 with .gzevent file
	event1Dir := filepath.Join(eventsDir, "event1")
	_ = os.MkdirAll(event1Dir, 0750)
	//nolint:gosec // G306: Test file permissions are acceptable
	_ = os.WriteFile(filepath.Join(event1Dir, ".gzevent"), []byte("title: Event 1\n"), 0644)

	// Create event2 with .gzevent file
	event2Dir := filepath.Join(eventsDir, "event2")
	_ = os.MkdirAll(event2Dir, 0750)
	//nolint:gosec // G306: Test file permissions are acceptable
	_ = os.WriteFile(filepath.Join(event2Dir, ".gzevent"), []byte("title: Event 2\n"), 0644)

	// Create event3 WITHOUT .gzevent file (should be ignored)
	event3Dir := filepath.Join(eventsDir, "event3")
	_ = os.MkdirAll(event3Dir, 0750)

	// Test getAvailableEvents
	events, err := getAvailableEvents()
	if err != nil {
		t.Fatalf("getAvailableEvents failed: %v", err)
	}

	// Should find event1 and event2, but not event3
	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d: %v", len(events), events)
	}

	// Verify event names
	eventMap := make(map[string]bool)
	for _, event := range events {
		eventMap[event] = true
	}

	if !eventMap["event1"] {
		t.Error("event1 should be in the list")
	}
	if !eventMap["event2"] {
		t.Error("event2 should be in the list")
	}
	if eventMap["event3"] {
		t.Error("event3 should not be in the list (no .gzevent)")
	}

	t.Logf("Found events: %v", events)
}

// TestGetAvailableEvents_NoEventsDir tests handling when events directory doesn't exist
func TestGetAvailableEvents_NoEventsDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "completion-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tmpDir)

	// Test without events directory
	events, err := getAvailableEvents()
	if err != nil {
		t.Fatalf("getAvailableEvents should not error when events dir is missing: %v", err)
	}

	if len(events) != 0 {
		t.Errorf("Expected 0 events when events dir doesn't exist, got %d", len(events))
	}
}

// TestValidEventNames tests the completion function
func TestValidEventNames(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "completion-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tmpDir)

	// Create events
	eventsDir := filepath.Join(tmpDir, "events")
	_ = os.MkdirAll(eventsDir, 0750)

	for _, eventName := range []string{"ctf2024", "ctf2025", "training"} {
		eventDir := filepath.Join(eventsDir, eventName)
		_ = os.MkdirAll(eventDir, 0750)
		//nolint:gosec // G306: Test file permissions are acceptable
		_ = os.WriteFile(filepath.Join(eventDir, ".gzevent"), []byte("title: Test\n"), 0644)
	}

	// Test the completion function
	cmd := &cobra.Command{}
	completions, directive := validEventNames(cmd, []string{}, "")

	if len(completions) != 3 {
		t.Errorf("Expected 3 completions, got %d: %v", len(completions), completions)
	}

	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("Expected NoFileComp directive, got %v", directive)
	}

	// Verify all events are in completions
	completionMap := make(map[string]bool)
	for _, comp := range completions {
		completionMap[comp] = true
	}

	for _, expected := range []string{"ctf2024", "ctf2025", "training"} {
		if !completionMap[expected] {
			t.Errorf("Expected %s in completions", expected)
		}
	}

	t.Logf("Completions: %v", completions)
}

// TestValidEventNames_EmptyDirectory tests completion with no events
func TestValidEventNames_EmptyDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "completion-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tmpDir)

	// Create empty events directory
	eventsDir := filepath.Join(tmpDir, "events")
	_ = os.MkdirAll(eventsDir, 0750)

	// Test the completion function
	cmd := &cobra.Command{}
	completions, directive := validEventNames(cmd, []string{}, "")

	if len(completions) != 0 {
		t.Errorf("Expected 0 completions, got %d: %v", len(completions), completions)
	}

	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("Expected NoFileComp directive, got %v", directive)
	}
}

// TestCompletionCommand tests the completion command exists
func TestCompletionCommand(t *testing.T) {
	// Find completion command
	var completionCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "completion" {
			completionCmd = cmd
			break
		}
	}

	if completionCmd == nil {
		t.Fatal("Completion command not found")
		return // Help staticcheck understand control flow
	}

	// Verify valid args
	validArgs := completionCmd.ValidArgs
	expectedArgs := []string{"bash", "zsh", "fish", "powershell"}

	if len(validArgs) != len(expectedArgs) {
		t.Errorf("Expected %d valid args, got %d", len(expectedArgs), len(validArgs))
	}

	for i, arg := range expectedArgs {
		if i >= len(validArgs) || validArgs[i] != arg {
			t.Errorf("Expected valid arg %s at position %d", arg, i)
		}
	}

	t.Log("Completion command validated successfully")
}
