package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// Helper function to setup test environment
func setupEventTestDir(t *testing.T) (string, func()) {
	t.Helper()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	tmpDir, err := os.MkdirTemp("", "gzcli-event-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	cleanup := func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore directory: %v", err)
		}
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("Failed to remove temp dir: %v", err)
		}
	}

	return tmpDir, cleanup
}

// createEventWithPoster creates an event with a poster file
func createEventWithPoster(t *testing.T, tmpDir, eventName, posterRelPath, posterContent string) (string, string) {
	t.Helper()
	eventDir := filepath.Join(tmpDir, EVENTS_DIR, eventName)
	if err := os.MkdirAll(eventDir, 0750); err != nil {
		t.Fatalf("Failed to create event dir: %v", err)
	}

	// Create poster file if content provided
	var posterPath string
	if posterContent != "" {
		if filepath.IsAbs(posterRelPath) {
			posterPath = posterRelPath
		} else {
			posterPath = filepath.Join(eventDir, posterRelPath)
			posterDir := filepath.Dir(posterPath)
			if err := os.MkdirAll(posterDir, 0750); err != nil {
				t.Fatalf("Failed to create poster dir: %v", err)
			}
		}
		//nolint:gosec // G306: Test file permissions are acceptable
		if err := os.WriteFile(posterPath, []byte(posterContent), 0644); err != nil {
			t.Fatalf("Failed to create poster file: %v", err)
		}
	}

	return eventDir, posterPath
}

// createGZEventFile creates a .gzevent file with given poster path
func createGZEventFile(t *testing.T, eventDir, title, posterPath string) {
	t.Helper()
	gzeventContent := fmt.Sprintf(`title: "%s"
start: "2024-10-01T00:00:00Z"
end: "2024-10-02T00:00:00Z"
poster: "%s"
`, title, posterPath)
	gzeventPath := filepath.Join(eventDir, GZEVENT_FILE)
	//nolint:gosec // G306: Test file permissions are acceptable
	if err := os.WriteFile(gzeventPath, []byte(gzeventContent), 0644); err != nil {
		t.Fatalf("Failed to create .gzevent: %v", err)
	}
}

// verifyPosterPath verifies poster path is absolute and matches expected
func verifyPosterPath(t *testing.T, eventConfig *EventConfig, expectedPath string) {
	t.Helper()
	if !filepath.IsAbs(eventConfig.Poster) {
		t.Errorf("Poster path should be absolute, got %q", eventConfig.Poster)
	}

	// Resolve symlinks in both paths for comparison (handles macOS /var -> /private/var)
	actualPath := eventConfig.Poster
	if canonical, err := filepath.EvalSymlinks(actualPath); err == nil {
		actualPath = canonical
	}

	expectedCanonical := expectedPath
	if canonical, err := filepath.EvalSymlinks(expectedPath); err == nil {
		expectedCanonical = canonical
	}

	if actualPath != expectedCanonical {
		t.Errorf("Poster path = %q, want %q", actualPath, expectedCanonical)
	}
}

func TestGetEventConfig_RelativePathResolution(t *testing.T) {
	tmpDir, cleanup := setupEventTestDir(t)
	defer cleanup()

	// Test case 1: Relative path from event directory
	t.Run("relative_path_from_event_dir", func(t *testing.T) {
		eventName := "test-event-1"
		eventDir, posterPath := createEventWithPoster(t, tmpDir, eventName, "poster.png", "fake image")
		createGZEventFile(t, eventDir, "Test Event", "poster.png")

		eventConfig, err := GetEventConfig(eventName)
		if err != nil {
			t.Fatalf("GetEventConfig() error = %v", err)
		}

		verifyPosterPath(t, eventConfig, posterPath)
	})

	// Test case 2: Relative path from workspace root
	t.Run("relative_path_from_workspace_root", func(t *testing.T) {
		// Create shared poster in .gzctf
		gzctfDir := filepath.Join(tmpDir, ".gzctf")
		_ = os.MkdirAll(gzctfDir, 0750)
		sharedPosterPath := filepath.Join(gzctfDir, "favicon.ico")
		//nolint:gosec // G306: Test file permissions are acceptable
		_ = os.WriteFile(sharedPosterPath, []byte("fake favicon"), 0644)

		eventName := "test-event-2"
		eventDir, _ := createEventWithPoster(t, tmpDir, eventName, "", "")
		createGZEventFile(t, eventDir, "Test Event 2", "../../.gzctf/favicon.ico")

		eventConfig, err := GetEventConfig(eventName)
		if err != nil {
			t.Fatalf("GetEventConfig() error = %v", err)
		}

		verifyPosterPath(t, eventConfig, sharedPosterPath)
	})

	// Test case 3: Absolute path remains unchanged
	t.Run("absolute_path_unchanged", func(t *testing.T) {
		eventName := "test-event-3"
		absolutePath := "/absolute/path/to/poster.png"
		eventDir, _ := createEventWithPoster(t, tmpDir, eventName, "", "")
		createGZEventFile(t, eventDir, "Test Event 3", absolutePath)

		eventConfig, err := GetEventConfig(eventName)
		if err != nil {
			t.Fatalf("GetEventConfig() error = %v", err)
		}

		if eventConfig.Poster != absolutePath {
			t.Errorf("Poster path = %q, want %q", eventConfig.Poster, absolutePath)
		}
	})

	// Test case 4: Missing poster file (path not resolved)
	t.Run("missing_poster_file", func(t *testing.T) {
		eventName := "test-event-4"
		eventDir, _ := createEventWithPoster(t, tmpDir, eventName, "", "")
		createGZEventFile(t, eventDir, "Test Event 4", "nonexistent.png")

		eventConfig, err := GetEventConfig(eventName)
		if err != nil {
			t.Fatalf("GetEventConfig() error = %v", err)
		}

		// When file doesn't exist, path should remain as-is (original behavior)
		if eventConfig.Poster != "nonexistent.png" {
			t.Errorf("Poster path = %q, want %q", eventConfig.Poster, "nonexistent.png")
		}
	})

	// Test case 5: Empty poster path
	t.Run("empty_poster_path", func(t *testing.T) {
		eventName := "test-event-5"
		eventDir, _ := createEventWithPoster(t, tmpDir, eventName, "", "")

		// Create .gzevent without poster
		gzeventContent := `title: "Test Event 5"
start: "2024-10-01T00:00:00Z"
end: "2024-10-02T00:00:00Z"
`
		gzeventPath := filepath.Join(eventDir, GZEVENT_FILE)
		//nolint:gosec // G306: Test file permissions are acceptable
		_ = os.WriteFile(gzeventPath, []byte(gzeventContent), 0644)

		eventConfig, err := GetEventConfig(eventName)
		if err != nil {
			t.Fatalf("GetEventConfig() error = %v", err)
		}

		if eventConfig.Poster != "" {
			t.Errorf("Poster path = %q, want empty string", eventConfig.Poster)
		}
	})
}

func TestGetEventConfig_Basic(t *testing.T) {
	tmpDir, cleanup := setupEventTestDir(t)
	defer cleanup()

	// Create test event
	eventName := "basic-event"
	eventDir := filepath.Join(tmpDir, EVENTS_DIR, eventName)
	if err := os.MkdirAll(eventDir, 0750); err != nil {
		t.Fatalf("Failed to create event dir: %v", err)
	}

	gzeventContent := `title: "Basic Test Event"
start: "2024-10-01T12:00:00Z"
end: "2024-10-02T12:00:00Z"
hidden: false
practiceMode: true
`
	gzeventPath := filepath.Join(eventDir, GZEVENT_FILE)
	//nolint:gosec // G306: Test file permissions are acceptable
	if err := os.WriteFile(gzeventPath, []byte(gzeventContent), 0644); err != nil {
		t.Fatalf("Failed to create .gzevent: %v", err)
	}

	// Test
	eventConfig, err := GetEventConfig(eventName)
	if err != nil {
		t.Fatalf("GetEventConfig() error = %v", err)
	}

	if eventConfig.Name != eventName {
		t.Errorf("Event name = %q, want %q", eventConfig.Name, eventName)
	}

	if eventConfig.Title != "Basic Test Event" {
		t.Errorf("Event title = %q, want %q", eventConfig.Title, "Basic Test Event")
	}

	if eventConfig.Hidden != false {
		t.Errorf("Event hidden = %v, want false", eventConfig.Hidden)
	}

	if eventConfig.PracticeMode != true {
		t.Errorf("Event practiceMode = %v, want true", eventConfig.PracticeMode)
	}
}

func TestGetEventConfig_NonexistentEvent(t *testing.T) {
	tmpDir, cleanup := setupEventTestDir(t)
	defer cleanup()

	// Create events directory but no event
	eventsDir := filepath.Join(tmpDir, EVENTS_DIR)
	if err := os.MkdirAll(eventsDir, 0750); err != nil {
		t.Fatalf("Failed to create events dir: %v", err)
	}

	// Test: Try to get nonexistent event
	_, err := GetEventConfig("nonexistent-event")
	if err == nil {
		t.Error("GetEventConfig() expected error for nonexistent event, got none")
	}
}
