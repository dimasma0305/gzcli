package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupResolveEventsTestDir creates test environment with events
func setupResolveEventsTestDir(t *testing.T, events []string) (string, func()) {
	t.Helper()
	originalDir, _ := os.Getwd()
	tmpDir, _ := os.MkdirTemp("", "gzcli-resolve-events-test-*")
	_ = os.Chdir(tmpDir)

	// Create test events
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

// runResolveTargetEventsTestCase runs a single test case
func runResolveTargetEventsTestCase(t *testing.T, tc struct {
	name         string
	eventFlags   []string
	excludeFlags []string
	wantEvents   []string
	wantErr      bool
	errContains  string
}) {
	t.Helper()
	result, err := ResolveTargetEvents(tc.eventFlags, tc.excludeFlags)

	// Check error expectation
	if tc.wantErr && err == nil {
		t.Errorf("%s: expected error but got none", tc.name)
		return
	}
	if !tc.wantErr && err != nil {
		t.Errorf("%s: unexpected error: %v", tc.name, err)
		return
	}
	if tc.wantErr && tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
		t.Errorf("%s: error should contain %q, got %q", tc.name, tc.errContains, err.Error())
		return
	}
	if tc.wantErr {
		return
	}

	// Verify result
	if len(result) != len(tc.wantEvents) {
		t.Errorf("%s: expected %d events, got %d", tc.name, len(tc.wantEvents), len(result))
	}
	for _, want := range tc.wantEvents {
		found := false
		for _, got := range result {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s: expected event %q not found", tc.name, want)
		}
	}
}

func TestResolveTargetEvents(t *testing.T) {
	_, cleanup := setupResolveEventsTestDir(t, []string{"ctf2024", "ctf2025", "practice"})
	defer cleanup()

	tests := []struct {
		name         string
		eventFlags   []string
		excludeFlags []string
		wantEvents   []string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "no flags returns all events",
			eventFlags:   []string{},
			excludeFlags: []string{},
			wantEvents:   []string{"ctf2024", "ctf2025", "practice"},
			wantErr:      false,
		},
		{
			name:         "specific event flag",
			eventFlags:   []string{"ctf2024"},
			excludeFlags: []string{},
			wantEvents:   []string{"ctf2024"},
			wantErr:      false,
		},
		{
			name:         "multiple specific event flags",
			eventFlags:   []string{"ctf2024", "practice"},
			excludeFlags: []string{},
			wantEvents:   []string{"ctf2024", "practice"},
			wantErr:      false,
		},
		{
			name:         "exclude one event",
			eventFlags:   []string{},
			excludeFlags: []string{"practice"},
			wantEvents:   []string{"ctf2024", "ctf2025"},
			wantErr:      false,
		},
		{
			name:         "exclude multiple events",
			eventFlags:   []string{},
			excludeFlags: []string{"ctf2024", "practice"},
			wantEvents:   []string{"ctf2025"},
			wantErr:      false,
		},
		{
			name:         "event flag takes priority over exclude",
			eventFlags:   []string{"ctf2024"},
			excludeFlags: []string{"practice"},
			wantEvents:   []string{"ctf2024"},
			wantErr:      false,
		},
		{
			name:         "nonexistent event returns error",
			eventFlags:   []string{"nonexistent"},
			excludeFlags: []string{},
			wantEvents:   nil,
			wantErr:      true,
			errContains:  "does not exist",
		},
		{
			name:         "exclude all events returns error",
			eventFlags:   []string{},
			excludeFlags: []string{"ctf2024", "ctf2025", "practice"},
			wantEvents:   nil,
			wantErr:      true,
			errContains:  "all events were excluded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runResolveTargetEventsTestCase(t, tt)
		})
	}
}

func TestResolveTargetEvents_NoEventsDirectory(t *testing.T) {
	// Setup: Create temporary test directory without events
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore directory: %v", err)
		}
	}()

	tmpDir, err := os.MkdirTemp("", "gzcli-no-events-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("Failed to remove temp dir: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	_, err = ResolveTargetEvents([]string{}, []string{})
	if err == nil {
		t.Error("ResolveTargetEvents() expected error when no events exist, but got none")
	}
}

func TestResolveTargetEvents_EmptyEventsDirectory(t *testing.T) {
	// Setup: Create temporary test directory with empty events directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore directory: %v", err)
		}
	}()

	tmpDir, err := os.MkdirTemp("", "gzcli-empty-events-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("Failed to remove temp dir: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Create empty events directory
	eventsDir := filepath.Join(tmpDir, "events")
	if err := os.MkdirAll(eventsDir, 0750); err != nil {
		t.Fatalf("Failed to create events dir: %v", err)
	}

	_, err = ResolveTargetEvents([]string{}, []string{})
	if err == nil {
		t.Error("ResolveTargetEvents() expected error when events directory is empty, but got none")
	}
	if err != nil && !containsSubstring(err.Error(), "no events found") {
		t.Errorf("ResolveTargetEvents() error = %v, want error containing 'no events found'", err)
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		item  string
		want  bool
	}{
		{
			name:  "item exists",
			slice: []string{"a", "b", "c"},
			item:  "b",
			want:  true,
		},
		{
			name:  "item does not exist",
			slice: []string{"a", "b", "c"},
			item:  "d",
			want:  false,
		},
		{
			name:  "empty slice",
			slice: []string{},
			item:  "a",
			want:  false,
		},
		{
			name:  "empty item in slice",
			slice: []string{"", "a", "b"},
			item:  "",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contains(tt.slice, tt.item); got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to check if two string slices contain the same elements (order independent)
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	aMap := make(map[string]int)
	for _, item := range a {
		aMap[item]++
	}

	for _, item := range b {
		if aMap[item] == 0 {
			return false
		}
		aMap[item]--
	}

	return true
}

// Helper function to check if a string contains a substring
func containsSubstring(s, substr string) bool {
	return strings.Contains(s, substr)
}
