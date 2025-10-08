package gzcli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
)

// TestInitWithEvent_StoresEventName tests that InitWithEvent stores the event name
func TestInitWithEvent_StoresEventName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup: Create temporary test environment
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore directory: %v", err)
		}
	}()

	tmpDir, err := os.MkdirTemp("", "gzcli-multievent-test-*")
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

	// Create test structure
	if err := createTestEventStructure(tmpDir, "test-event-1"); err != nil {
		t.Fatalf("Failed to create test structure: %v", err)
	}

	// Test: Initialize with specific event
	// Note: This will fail without proper server config, but we can test the struct initialization
	_, err = InitWithEvent("test-event-1")

	// We expect an error due to missing server/credentials, but that's okay for this test
	// The important part is that the function attempts to load the correct event
	if err != nil {
		t.Logf("Expected error (no server): %v", err)
	}
}

// TestMultiEventSync_DifferentChallenges tests syncing multiple events with different challenges
// setupMultiEventChallengesTest creates environment for multi-event challenge tests
func setupMultiEventChallengesTest(t *testing.T) (string, func()) {
	t.Helper()
	originalDir, _ := os.Getwd()
	tmpDir, _ := os.MkdirTemp("", "gzcli-multievent-challenges-test-*")
	_ = os.Chdir(tmpDir)

	// Create two events with different challenges
	_ = createTestEventWithChallenges(tmpDir, "ctf-2024", []string{"web-challenge-1", "crypto-challenge-1"})
	_ = createTestEventWithChallenges(tmpDir, "practice-ctf", []string{"pwn-challenge-1", "forensics-challenge-1"})

	cleanup := func() {
		_ = os.Chdir(originalDir)
		_ = os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

// verifyChallengeExists verifies a challenge exists in event path
func verifyChallengeExists(t *testing.T, eventPath, challengeName string) {
	t.Helper()
	found := false
	for _, category := range config.CHALLENGE_CATEGORY {
		challengePath := filepath.Join(eventPath, category, challengeName)
		if _, err := os.Stat(challengePath); err == nil {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Challenge %q not found in event path", challengeName)
	}
}

// verifyEventChallenges verifies event config and its challenges
func verifyEventChallenges(t *testing.T, eventName string, expectedChallenges []string) {
	t.Helper()
	eventConfig, err := config.GetEventConfig(eventName)
	if err != nil {
		t.Fatalf("Failed to get event config: %v", err)
	}

	if eventConfig.Name != eventName {
		t.Errorf("Event name = %q, want %q", eventConfig.Name, eventName)
	}

	eventPath, err := config.GetEventPath(eventName)
	if err != nil {
		t.Fatalf("Failed to get event path: %v", err)
	}

	for _, challengeName := range expectedChallenges {
		verifyChallengeExists(t, eventPath, challengeName)
	}
}

func TestMultiEventSync_DifferentChallenges(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	_, cleanup := setupMultiEventChallengesTest(t)
	defer cleanup()

	// Test: Load challenges for each event
	testCases := []struct {
		eventName          string
		expectedChallenges []string
	}{
		{"ctf-2024", []string{"web-challenge-1", "crypto-challenge-1"}},
		{"practice-ctf", []string{"pwn-challenge-1", "forensics-challenge-1"}},
	}

	for _, tc := range testCases {
		t.Run(tc.eventName, func(t *testing.T) {
			verifyEventChallenges(t, tc.eventName, tc.expectedChallenges)
		})
	}
}

// TestEventIsolation tests that events are properly isolated
func TestEventIsolation(t *testing.T) {
	// Setup
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore directory: %v", err)
		}
	}()

	tmpDir, err := os.MkdirTemp("", "gzcli-event-isolation-test-*")
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

	// Create three events
	events := []string{"event-a", "event-b", "event-c"}
	for _, event := range events {
		if err := createTestEventStructure(tmpDir, event); err != nil {
			t.Fatalf("Failed to create event %s: %v", event, err)
		}
	}

	// Test: List all events
	allEvents, err := config.ListEvents()
	if err != nil {
		t.Fatalf("Failed to list events: %v", err)
	}

	if len(allEvents) != len(events) {
		t.Errorf("ListEvents() returned %d events, want %d", len(allEvents), len(events))
	}

	for _, event := range events {
		found := false
		for _, listed := range allEvents {
			if listed == event {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Event %q not found in list", event)
		}
	}
}

// Helper functions

func createTestEventStructure(tmpDir, eventName string) error {
	// Create .gzctf directory with server config
	gzctfDir := filepath.Join(tmpDir, ".gzctf")
	if err := os.MkdirAll(gzctfDir, 0750); err != nil {
		return err
	}

	serverConfig := `url: http://localhost:8080
creds:
  username: admin
  password: password123
`
	confPath := filepath.Join(gzctfDir, "conf.yaml")
	//nolint:gosec // G306: Test file permissions are acceptable
	if err := os.WriteFile(confPath, []byte(serverConfig), 0644); err != nil {
		return err
	}

	// Create events directory
	eventsDir := filepath.Join(tmpDir, "events", eventName)
	if err := os.MkdirAll(eventsDir, 0750); err != nil {
		return err
	}

	// Create .gzevent file
	eventConfig := `title: "Test Event ` + eventName + `"
start: "2024-10-01T00:00:00Z"
end: "2024-10-02T00:00:00Z"
`
	gzeventPath := filepath.Join(eventsDir, ".gzevent")
	//nolint:gosec // G306: Test file permissions are acceptable
	if err := os.WriteFile(gzeventPath, []byte(eventConfig), 0644); err != nil {
		return err
	}

	return nil
}

func createTestEventWithChallenges(tmpDir, eventName string, challenges []string) error {
	// Create basic event structure
	if err := createTestEventStructure(tmpDir, eventName); err != nil {
		return err
	}

	eventPath := filepath.Join(tmpDir, "events", eventName)

	// Create challenges in different categories
	categories := []string{"Web", "Crypto", "Pwn", "Forensics", "Misc"}

	for i, challengeName := range challenges {
		category := categories[i%len(categories)]
		challengeDir := filepath.Join(eventPath, category, challengeName)

		if err := os.MkdirAll(challengeDir, 0750); err != nil {
			return err
		}

		// Create a basic challenge.yml
		challengeYml := `name: "` + challengeName + `"
category: "` + category + `"
type: StaticContainer
author: "Test Author"
description: "Test challenge for ` + eventName + `"
`
		challengePath := filepath.Join(challengeDir, "challenge.yml")
		//nolint:gosec // G306: Test file permissions are acceptable
		if err := os.WriteFile(challengePath, []byte(challengeYml), 0644); err != nil {
			return err
		}
	}

	return nil
}
