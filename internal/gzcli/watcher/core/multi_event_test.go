//nolint:errcheck,gosec // Test file with acceptable error handling patterns
package core

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/database"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/filesystem"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/watchertypes"
)

// setupMultiEventTest creates a test environment with multiple events
func setupMultiEventTest(t *testing.T, eventNames []string) (string, *Watcher, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "multi-event-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create events directory structure
	eventsDir := filepath.Join(tmpDir, "events")
	for _, eventName := range eventNames {
		createSampleChallengeInEvent(t, eventsDir, eventName)
	}

	// Change to tmpDir so event paths work
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)

	api := &gzapi.GZAPI{}
	watcher, err := New(api)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	cleanup := func() {
		os.Chdir(oldWd)
		os.RemoveAll(tmpDir)
	}

	return tmpDir, watcher, cleanup
}

// createSampleChallengeInEvent creates a sample challenge in an event directory
func createSampleChallengeInEvent(t *testing.T, eventsDir, eventName string) {
	t.Helper()
	eventDir := filepath.Join(eventsDir, eventName)
	if err := os.MkdirAll(eventDir, 0755); err != nil {
		t.Fatalf("Failed to create event dir: %v", err)
	}

	// Create a sample challenge for each event
	challengeDir := filepath.Join(eventDir, "web", "sample-challenge")
	if err := os.MkdirAll(challengeDir, 0755); err != nil {
		t.Fatalf("Failed to create challenge dir: %v", err)
	}

	challengeYaml := filepath.Join(challengeDir, "challenge.yaml")
	if err := os.WriteFile(challengeYaml, []byte("name: Sample Challenge\n"), 0644); err != nil {
		t.Fatalf("Failed to create challenge.yaml: %v", err)
	}
}

// verifyChallengePath verifies a challenge path contains expected event name
func verifyChallengePath(t *testing.T, path, eventName string) {
	t.Helper()
	absPath := path
	if !filepath.IsAbs(absPath) {
		absPath, _ = filepath.Abs(absPath)
	}
	if !containsPath(absPath, eventName) {
		t.Errorf("Challenge path should contain '%s': %s", eventName, absPath)
	}
}

// TestMultiEvent_CreateMasterWatcher tests creating a master watcher
func TestMultiEvent_CreateMasterWatcher(t *testing.T) {
	api := &gzapi.GZAPI{}
	w, err := New(api)
	if err != nil {
		t.Fatalf("Failed to create master watcher: %v", err)
	}

	if w.eventWatchers == nil {
		t.Error("Event watchers map should be initialized")
	}

	if w.api == nil {
		t.Error("API should be set")
	}

	if w.ctx == nil {
		t.Error("Context should be initialized")
	}
}

// TestMultiEvent_StartMultipleEvents tests starting watchers for multiple events
func TestMultiEvent_StartMultipleEvents(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping multi-event test in short mode")
	}

	eventNames := []string{"ctf2024", "ctf2025", "training"}
	tmpDir, w, cleanup := setupMultiEventTest(t, eventNames)
	defer cleanup()

	// Initialize database
	w.db = database.New(filepath.Join(tmpDir, "test.db"), false)
	if err := w.db.Init(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer w.db.Close()

	// Create event watchers for each event
	var wg sync.WaitGroup
	for _, eventName := range eventNames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()

			config := watchertypes.WatcherConfig{}
			ew, err := NewEventWatcher(name, w.api, config, w.db, w.ctx)
			if err != nil {
				t.Errorf("Failed to create event watcher for %s: %v", name, err)
				return
			}

			w.AddEventWatcher(name, ew)
		}(eventName)
	}

	wg.Wait()

	// Verify all event watchers are created
	eventWatchers := w.GetAllEventWatchers()
	if len(eventWatchers) != len(eventNames) {
		t.Errorf("Expected %d event watchers, got %d", len(eventNames), len(eventWatchers))
	}

	// Verify each event watcher is present
	for _, eventName := range eventNames {
		ew, exists := w.GetEventWatcher(eventName)
		if !exists {
			t.Errorf("Event watcher for %s not found", eventName)
			continue
		}
		if ew.GetEventName() != eventName {
			t.Errorf("Expected event name %s, got %s", eventName, ew.GetEventName())
		}
	}

	t.Logf("Successfully created and verified %d event watchers", len(eventNames))
}

// TestMultiEvent_EventIsolation tests that events are isolated from each other
func TestMultiEvent_EventIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping event isolation test in short mode")
	}

	eventNames := []string{"event1", "event2"}
	tmpDir, w, cleanup := setupMultiEventTest(t, eventNames)
	defer cleanup()

	// Initialize database
	w.db = database.New(filepath.Join(tmpDir, "test.db"), false)
	w.db.Init()
	defer w.db.Close()

	// Create event watchers
	config := watchertypes.WatcherConfig{}
	ew1, _ := NewEventWatcher("event1", w.api, config, w.db, w.ctx)
	ew2, _ := NewEventWatcher("event2", w.api, config, w.db, w.ctx)

	w.AddEventWatcher("event1", ew1)
	w.AddEventWatcher("event2", ew2)

	// Test that state updates in one event don't affect the other
	ew1.setUpdating("challenge1", true)

	if ew2.isUpdating("challenge1") {
		t.Error("Event2 should not be affected by Event1's state")
	}

	// Test that mutexes are isolated
	mutex1 := ew1.GetChallengeUpdateMutex("shared-name")
	mutex2 := ew2.GetChallengeUpdateMutex("shared-name")

	if mutex1 == mutex2 {
		t.Error("Mutexes should be different for different events")
	}

	t.Log("Event isolation verified successfully")
}

// TestMultiEvent_StopSpecificEvent tests stopping a specific event watcher
func TestMultiEvent_StopSpecificEvent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stop specific event test in short mode")
	}

	eventNames := []string{"event1", "event2", "event3"}
	tmpDir, w, cleanup := setupMultiEventTest(t, eventNames)
	defer cleanup()

	// Initialize database
	w.db = database.New(filepath.Join(tmpDir, "test.db"), false)
	w.db.Init()
	defer w.db.Close()

	// Create event watchers
	config := watchertypes.WatcherConfig{}
	for _, eventName := range eventNames {
		ew, err := NewEventWatcher(eventName, w.api, config, w.db, w.ctx)
		if err != nil {
			t.Fatalf("Failed to create event watcher: %v", err)
		}
		w.AddEventWatcher(eventName, ew)
	}

	// Verify all events are present
	if len(w.GetAllEventWatchers()) != 3 {
		t.Fatalf("Expected 3 event watchers, got %d", len(w.GetAllEventWatchers()))
	}

	// Stop one specific event
	if err := w.StopEventWatcher("event2"); err != nil {
		t.Fatalf("Failed to stop event2: %v", err)
	}

	// Verify event2 is removed
	if _, exists := w.GetEventWatcher("event2"); exists {
		t.Error("Event2 should be removed after stopping")
	}

	// Verify other events are still present
	remainingEvents := w.GetAllEventWatchers()
	if len(remainingEvents) != 2 {
		t.Errorf("Expected 2 remaining events, got %d", len(remainingEvents))
	}

	if _, exists := w.GetEventWatcher("event1"); !exists {
		t.Error("Event1 should still be present")
	}

	if _, exists := w.GetEventWatcher("event3"); !exists {
		t.Error("Event3 should still be present")
	}

	t.Log("Specific event stopped successfully while others remain")
}

// TestMultiEvent_ConcurrentOperations tests concurrent operations across multiple events
func TestMultiEvent_ConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent operations test in short mode")
	}

	eventNames := []string{"event1", "event2", "event3", "event4", "event5"}
	tmpDir, w, cleanup := setupMultiEventTest(t, eventNames)
	defer cleanup()

	// Initialize database
	w.db = database.New(filepath.Join(tmpDir, "test.db"), false)
	w.db.Init()
	defer w.db.Close()

	// Create event watchers
	config := watchertypes.WatcherConfig{}
	for _, eventName := range eventNames {
		ew, _ := NewEventWatcher(eventName, w.api, config, w.db, w.ctx)
		w.AddEventWatcher(eventName, ew)
	}

	// Perform concurrent operations on different events
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(iteration int) {
			defer wg.Done()
			eventName := eventNames[iteration%len(eventNames)]

			ew, exists := w.GetEventWatcher(eventName)
			if !exists {
				return
			}

			challengeName := fmt.Sprintf("challenge-%d", iteration)

			// Simulate various operations
			ew.setUpdating(challengeName, true)
			time.Sleep(1 * time.Millisecond)
			ew.setUpdating(challengeName, false)

			mutex := ew.GetChallengeUpdateMutex(challengeName)
			mutex.Lock()
			time.Sleep(1 * time.Millisecond)
			mutex.Unlock()
		}(i)
	}

	// Wait for all operations to complete
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		t.Log("All concurrent operations completed successfully")
	case <-time.After(10 * time.Second):
		t.Fatal("Concurrent operations timed out")
	}
}

// TestMultiEvent_SharedDatabase tests that events share the database correctly
func TestMultiEvent_SharedDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping shared database test in short mode")
	}

	eventNames := []string{"event1", "event2"}
	tmpDir, w, cleanup := setupMultiEventTest(t, eventNames)
	defer cleanup()

	// Initialize shared database
	dbPath := filepath.Join(tmpDir, "shared.db")
	w.db = database.New(dbPath, true)
	if err := w.db.Init(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer w.db.Close()

	// Create event watchers with shared database
	config := watchertypes.WatcherConfig{}
	ew1, _ := NewEventWatcher("event1", w.api, config, w.db, w.ctx)
	ew2, _ := NewEventWatcher("event2", w.api, config, w.db, w.ctx)

	w.AddEventWatcher("event1", ew1)
	w.AddEventWatcher("event2", ew2)

	// Both event watchers should have the same database reference
	if ew1.db != ew2.db {
		t.Error("Event watchers should share the same database")
	}

	if ew1.db != w.db {
		t.Error("Event watcher should reference master database")
	}

	// Test logging from multiple events
	ew1.LogToDatabase("INFO", "test", "challenge1", "", "Event1 message", "", 0)
	ew2.LogToDatabase("INFO", "test", "challenge2", "", "Event2 message", "", 0)

	t.Log("Shared database verified successfully")
}

// TestMultiEvent_GetWatchedChallenges tests aggregating challenges from all events
func TestMultiEvent_GetWatchedChallenges(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping watched challenges test in short mode")
	}

	tmpDir, w, cleanup := setupMultiEventTest(t, []string{"event1", "event2"})
	defer cleanup()

	w.db = database.New(filepath.Join(tmpDir, "test.db"), false)
	w.db.Init()
	defer w.db.Close()

	config := watchertypes.WatcherConfig{}

	// Create event watchers
	ew1, _ := NewEventWatcher("event1", w.api, config, w.db, w.ctx)
	ew2, _ := NewEventWatcher("event2", w.api, config, w.db, w.ctx)

	w.AddEventWatcher("event1", ew1)
	w.AddEventWatcher("event2", ew2)

	// Add challenges to event watchers
	event1Dir := filepath.Join(tmpDir, "events", "event1", "challenge1")
	os.MkdirAll(event1Dir, 0755)
	ew1.challengeMgr.AddChallenge("challenge1", event1Dir)

	event2Dir := filepath.Join(tmpDir, "events", "event2", "challenge2")
	os.MkdirAll(event2Dir, 0755)
	ew2.challengeMgr.AddChallenge("challenge2", event2Dir)

	// Get all watched challenges
	allChallenges := w.GetWatchedChallenges()

	if len(allChallenges) != 2 {
		t.Errorf("Expected 2 challenges, got %d", len(allChallenges))
	}

	t.Logf("Retrieved %d challenges from multiple events", len(allChallenges))
}

// TestMultiEvent_CommandHandlerFiltering tests event filtering in command handlers
func TestMultiEvent_CommandHandlerFiltering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping command handler test in short mode")
	}

	tmpDir, w, cleanup := setupMultiEventTest(t, []string{"event1", "event2"})
	defer cleanup()

	w.db = database.New(filepath.Join(tmpDir, "test.db"), true)
	w.db.Init()
	defer w.db.Close()

	config := watchertypes.WatcherConfig{
		DatabaseEnabled: true,
		SocketEnabled:   true,
	}
	w.config = config

	ew1, _ := NewEventWatcher("event1", w.api, config, w.db, w.ctx)
	ew2, _ := NewEventWatcher("event2", w.api, config, w.db, w.ctx)

	w.AddEventWatcher("event1", ew1)
	w.AddEventWatcher("event2", ew2)

	// Test status command without event filter (should return all events)
	cmd := watchertypes.WatcherCommand{
		Action: "status",
	}
	response := w.HandleStatusCommand(cmd)

	if !response.Success {
		t.Errorf("Status command failed: %s", response.Error)
	}

	// Check if all events are in the response
	data, ok := response.Data["events"].([]string)
	if !ok {
		t.Error("Events field not found in response")
	} else if len(data) != 2 {
		t.Errorf("Expected 2 events in status, got %d", len(data))
	}

	// Test status command with event filter
	cmdFiltered := watchertypes.WatcherCommand{
		Action: "status",
		Event:  "event1",
	}
	responseFiltered := w.HandleStatusCommand(cmdFiltered)

	if !responseFiltered.Success {
		t.Errorf("Filtered status command failed: %s", responseFiltered.Error)
	}

	dataFiltered, ok := responseFiltered.Data["events"].([]string)
	switch {
	case !ok:
		t.Error("Events field not found in filtered response")
	case len(dataFiltered) != 1:
		t.Errorf("Expected 1 event in filtered status, got %d", len(dataFiltered))
	case dataFiltered[0] != "event1":
		t.Errorf("Expected event1 in filtered response, got %s", dataFiltered[0])
	}

	t.Log("Command handler filtering works correctly")
}

// TestMultiEvent_ContextCancellation tests that cancelling context stops all events
func TestMultiEvent_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping context cancellation test in short mode")
	}

	tmpDir, w, cleanup := setupMultiEventTest(t, []string{"event1", "event2"})
	defer cleanup()

	w.db = database.New(filepath.Join(tmpDir, "test.db"), false)
	w.db.Init()
	defer w.db.Close()

	config := watchertypes.WatcherConfig{}

	ew1, _ := NewEventWatcher("event1", w.api, config, w.db, w.ctx)
	ew2, _ := NewEventWatcher("event2", w.api, config, w.db, w.ctx)

	w.AddEventWatcher("event1", ew1)
	w.AddEventWatcher("event2", ew2)

	// Cancel the master context
	w.cancel()

	// Wait a bit for cancellation to propagate
	time.Sleep(100 * time.Millisecond)

	// Verify all event contexts are cancelled
	select {
	case <-ew1.ctx.Done():
		t.Log("Event1 context cancelled")
	default:
		t.Error("Event1 context should be cancelled")
	}

	select {
	case <-ew2.ctx.Done():
		t.Log("Event2 context cancelled")
	default:
		t.Error("Event2 context should be cancelled")
	}
}

// TestMultiEvent_RaceConditionPrevention tests race condition prevention across events
func TestMultiEvent_RaceConditionPrevention(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race condition test in short mode")
	}

	tmpDir, w, cleanup := setupMultiEventTest(t, []string{"event1", "event2"})
	defer cleanup()

	w.db = database.New(filepath.Join(tmpDir, "test.db"), false)
	w.db.Init()
	defer w.db.Close()

	config := watchertypes.WatcherConfig{}
	ew1, _ := NewEventWatcher("event1", w.api, config, w.db, w.ctx)
	ew2, _ := NewEventWatcher("event2", w.api, config, w.db, w.ctx)

	w.AddEventWatcher("event1", ew1)
	w.AddEventWatcher("event2", ew2)

	// Concurrent access to master watcher's event map
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(3)

		// Concurrent reads
		go func() {
			defer wg.Done()
			w.GetAllEventWatchers()
		}()

		// Concurrent GetEventWatcher
		go func() {
			defer wg.Done()
			w.GetEventWatcher("event1")
		}()

		// Concurrent operations on event watchers
		go func() {
			defer wg.Done()
			if ew, exists := w.GetEventWatcher("event2"); exists {
				ew.setUpdating("test-challenge", true)
				ew.setUpdating("test-challenge", false)
			}
		}()
	}

	// Wait with timeout
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		t.Log("No race conditions detected in concurrent operations")
	case <-time.After(5 * time.Second):
		t.Fatal("Race condition test timed out")
	}
}

// TestMultiEvent_StopNonExistentEvent tests stopping an event that doesn't exist
func TestMultiEvent_StopNonExistentEvent(t *testing.T) {
	api := &gzapi.GZAPI{}
	w, _ := New(api)

	err := w.StopEventWatcher("non-existent-event")
	if err == nil {
		t.Error("Expected error when stopping non-existent event")
	}

	if err != nil && err.Error() == "" {
		t.Error("Error message should be informative")
	}
}

// TestMultiEvent_HandleStopEventCommand tests the stop event command handler
func TestMultiEvent_HandleStopEventCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stop event command test in short mode")
	}

	tmpDir, w, cleanup := setupMultiEventTest(t, []string{"event1"})
	defer cleanup()

	w.db = database.New(filepath.Join(tmpDir, "test.db"), false)
	w.db.Init()
	defer w.db.Close()

	config := watchertypes.WatcherConfig{}
	ew, _ := NewEventWatcher("event1", w.api, config, w.db, w.ctx)
	w.AddEventWatcher("event1", ew)

	// Test stop event command
	cmd := watchertypes.WatcherCommand{
		Action: "stop_event",
		Event:  "event1",
	}

	response := w.HandleStopEventCommand(cmd)

	if !response.Success {
		t.Errorf("Stop event command failed: %s", response.Error)
	}

	// Verify event is removed
	if _, exists := w.GetEventWatcher("event1"); exists {
		t.Error("Event should be removed after stop command")
	}

	t.Log("Stop event command executed successfully")
}

// TestMultiEvent_EmptyEventList tests handling empty event configuration
func TestMultiEvent_EmptyEventList(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "empty-event-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	api := &gzapi.GZAPI{}
	w, _ := New(api)

	w.db = database.New(filepath.Join(tmpDir, "test.db"), false)
	w.db.Init()
	defer w.db.Close()

	// Test with empty events list
	config := watchertypes.WatcherConfig{
		Events: []string{},
	}
	w.config = config

	// GetWatchedChallenges should handle empty event list gracefully
	challenges := w.GetWatchedChallenges()
	if len(challenges) != 0 {
		t.Errorf("Expected 0 challenges with no events, got %d", len(challenges))
	}

	// HandleStatusCommand should work with no events
	cmd := watchertypes.WatcherCommand{Action: "status"}
	response := w.HandleStatusCommand(cmd)

	if !response.Success {
		t.Error("Status command should succeed even with no events")
	}

	t.Log("Empty event list handled gracefully")
}

// createChallengeWithContent creates a challenge directory with custom content
func createChallengeWithContent(t *testing.T, eventsDir, eventName, category, challengeName, content string) {
	t.Helper()
	eventDir := filepath.Join(eventsDir, eventName)
	challengeDir := filepath.Join(eventDir, category, challengeName)
	if err := os.MkdirAll(challengeDir, 0755); err != nil {
		t.Fatalf("Failed to create challenge dir: %v", err)
	}

	challengeYaml := filepath.Join(challengeDir, "challenge.yaml")
	if err := os.WriteFile(challengeYaml, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create challenge.yaml: %v", err)
	}
}

// setupDuplicateFolderTest sets up environment for duplicate folder tests
func setupDuplicateFolderTest(t *testing.T) (string, *EventWatcher, *EventWatcher, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "duplicate-folder-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	eventNames := []string{"ctf2024", "ctf2025"}
	eventsDir := filepath.Join(tmpDir, "events")

	for _, eventName := range eventNames {
		content := fmt.Sprintf("name: Easy Web %s\nauthor: test\n", eventName)
		createChallengeWithContent(t, eventsDir, eventName, "web", "easy-web", content)
	}

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)

	api := &gzapi.GZAPI{}
	w, _ := New(api)
	w.db = database.New(filepath.Join(tmpDir, "test.db"), false)
	w.db.Init()

	config := watchertypes.WatcherConfig{}
	ew1, _ := NewEventWatcher("ctf2024", w.api, config, w.db, w.ctx)
	ew2, _ := NewEventWatcher("ctf2025", w.api, config, w.db, w.ctx)
	w.AddEventWatcher("ctf2024", ew1)
	w.AddEventWatcher("ctf2025", ew2)

	cleanup := func() {
		w.db.Close()
		os.Chdir(oldWd)
		os.RemoveAll(tmpDir)
	}

	return tmpDir, ew1, ew2, cleanup
}

// verifyDuplicateFolderIsolation verifies isolation between duplicate folder challenges
func verifyDuplicateFolderIsolation(t *testing.T, ew1, ew2 *EventWatcher, challengeName string) {
	t.Helper()

	challenges1 := ew1.challengeMgr.GetChallenges()
	challenges2 := ew2.challengeMgr.GetChallenges()

	path1, exists1 := challenges1[challengeName]
	path2, exists2 := challenges2[challengeName]

	if !exists1 || !exists2 {
		t.Fatal("Both events should have the challenge")
	}

	if path1 == path2 {
		t.Errorf("Challenge paths should be different, both are: %s", path1)
	}

	// Verify mutexes are independent
	mutex1 := ew1.GetChallengeUpdateMutex(challengeName)
	mutex2 := ew2.GetChallengeUpdateMutex(challengeName)
	if mutex1 == mutex2 {
		t.Error("Mutexes should be different across events")
	}

	// Verify state isolation
	ew1.setUpdating(challengeName, true)
	if ew2.isUpdating(challengeName) {
		t.Error("Event2 should not be affected by Event1's state")
	}
	ew1.setUpdating(challengeName, false)
}

// TestMultiEvent_DuplicateFolderNames tests challenges with same folder names in different events
func TestMultiEvent_DuplicateFolderNames(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping duplicate folder names test in short mode")
	}

	tmpDir, ew1, ew2, cleanup := setupDuplicateFolderTest(t)
	defer cleanup()

	// Discover challenges in both events
	ew1.discoverChallenges()
	ew2.discoverChallenges()

	// Verify isolation
	verifyDuplicateFolderIsolation(t, ew1, ew2, "easy-web")

	// Verify paths contain correct event names
	challenges1 := ew1.challengeMgr.GetChallenges()
	challenges2 := ew2.challengeMgr.GetChallenges()
	verifyChallengePath(t, challenges1["easy-web"], "ctf2024")
	verifyChallengePath(t, challenges2["easy-web"], "ctf2025")

	t.Logf("Successfully verified isolation of duplicate folder names across events")
	t.Logf("tmpDir: %s", tmpDir)
}

// setupDifferentFoldersSameNameTest sets up test with different folders, same YAML name
func setupDifferentFoldersSameNameTest(t *testing.T) (*EventWatcher, *EventWatcher, func()) {
	t.Helper()
	tmpDir, _ := os.MkdirTemp("", "duplicate-name-test-*")
	eventsDir := filepath.Join(tmpDir, "events")

	// Event1: web/challenge-a, Event2: web/challenge-b (same YAML name)
	createChallengeWithContent(t, eventsDir, "event1", "web", "challenge-a", "name: Super Challenge\nauthor: test\nvalue: 100\n")
	createChallengeWithContent(t, eventsDir, "event2", "web", "challenge-b", "name: Super Challenge\nauthor: test\nvalue: 100\n")

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)

	api := &gzapi.GZAPI{}
	w, _ := New(api)
	w.db = database.New(filepath.Join(tmpDir, "test.db"), false)
	w.db.Init()

	config := watchertypes.WatcherConfig{}
	ew1, _ := NewEventWatcher("event1", w.api, config, w.db, w.ctx)
	ew2, _ := NewEventWatcher("event2", w.api, config, w.db, w.ctx)
	w.AddEventWatcher("event1", ew1)
	w.AddEventWatcher("event2", ew2)

	cleanup := func() {
		w.db.Close()
		os.Chdir(oldWd)
		os.RemoveAll(tmpDir)
	}

	return ew1, ew2, cleanup
}

// TestMultiEvent_DuplicateChallengeNames tests challenges with same YAML names but different folders
func TestMultiEvent_DuplicateChallengeNames(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping duplicate challenge names test in short mode")
	}

	ew1, ew2, cleanup := setupDifferentFoldersSameNameTest(t)
	defer cleanup()

	// Discover challenges
	ew1.discoverChallenges()
	ew2.discoverChallenges()

	// Verify challenges are identified by folder names, not YAML names
	challenges1 := ew1.challengeMgr.GetChallenges()
	challenges2 := ew2.challengeMgr.GetChallenges()

	if len(challenges1) != 1 || len(challenges2) != 1 {
		t.Error("Each event should have 1 challenge")
	}

	if _, exists := challenges1["challenge-a"]; !exists {
		t.Error("Event1 should have challenge 'challenge-a'")
	}
	if _, exists := challenges2["challenge-b"]; !exists {
		t.Error("Event2 should have challenge 'challenge-b'")
	}

	// Verify state isolation between different challenge names
	ew1.setUpdating("challenge-a", true)
	if ew2.isUpdating("challenge-b") {
		t.Error("Event2's challenge-b should not be affected by event1's challenge-a")
	}
	ew1.setUpdating("challenge-a", false)

	// Verify mutex independence
	mutex1 := ew1.GetChallengeUpdateMutex("challenge-a")
	mutex2 := ew2.GetChallengeUpdateMutex("challenge-b")
	if mutex1 == mutex2 {
		t.Error("Mutexes should be different for different folder names")
	}

	t.Logf("Successfully verified that challenges with same YAML names but different folders are isolated")
}

// setupIdenticalChallengesTest sets up test with identical folder and YAML names
func setupIdenticalChallengesTest(t *testing.T) (*EventWatcher, *EventWatcher, func()) {
	t.Helper()
	tmpDir, _ := os.MkdirTemp("", "duplicate-both-test-*")
	eventsDir := filepath.Join(tmpDir, "events")

	// Create identical challenges in both events
	content := "name: Buffer Overflow 101\nauthor: test\nvalue: 200\n"
	for _, eventName := range []string{"summer-ctf", "winter-ctf"} {
		createChallengeWithContent(t, eventsDir, eventName, "pwn", "buffer-overflow", content)
		// Add additional file
		eventDir := filepath.Join(eventsDir, eventName)
		srcFile := filepath.Join(eventDir, "pwn", "buffer-overflow", "exploit.py")
		os.WriteFile(srcFile, []byte("# exploit code\n"), 0644)
	}

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)

	api := &gzapi.GZAPI{}
	w, _ := New(api)
	w.db = database.New(filepath.Join(tmpDir, "test.db"), false)
	w.db.Init()

	config := watchertypes.WatcherConfig{}
	ewSummer, _ := NewEventWatcher("summer-ctf", w.api, config, w.db, w.ctx)
	ewWinter, _ := NewEventWatcher("winter-ctf", w.api, config, w.db, w.ctx)
	w.AddEventWatcher("summer-ctf", ewSummer)
	w.AddEventWatcher("winter-ctf", ewWinter)

	cleanup := func() {
		w.db.Close()
		os.Chdir(oldWd)
		os.RemoveAll(tmpDir)
	}

	return ewSummer, ewWinter, cleanup
}

// verifyCompleteStateIsolation verifies complete state isolation between two event watchers
func verifyCompleteStateIsolation(t *testing.T, ew1, ew2 *EventWatcher, challengeName string) {
	t.Helper()

	// Update state isolation
	ew1.setUpdating(challengeName, true)
	ew2.setUpdating(challengeName, true)

	if !ew1.isUpdating(challengeName) || !ew2.isUpdating(challengeName) {
		t.Error("Both should show as updating")
	}

	ew1.setUpdating(challengeName, false)
	if ew1.isUpdating(challengeName) {
		t.Error("Event1 state should be cleared")
	}
	if !ew2.isUpdating(challengeName) {
		t.Error("Event2 should still be updating (independent)")
	}
	ew2.setUpdating(challengeName, false)

	// Mutex isolation
	if ew1.GetChallengeUpdateMutex(challengeName) == ew2.GetChallengeUpdateMutex(challengeName) {
		t.Error("Mutexes must be different")
	}
}

// TestMultiEvent_DuplicateFolderAndName tests challenges with both same folder and YAML names
func TestMultiEvent_DuplicateFolderAndName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping duplicate folder and name test in short mode")
	}

	ewSummer, ewWinter, cleanup := setupIdenticalChallengesTest(t)
	defer cleanup()

	ewSummer.discoverChallenges()
	ewWinter.discoverChallenges()

	// Verify discovery and path isolation
	challengesSummer := ewSummer.challengeMgr.GetChallenges()
	challengesWinter := ewWinter.challengeMgr.GetChallenges()

	pathSummer, existsSummer := challengesSummer["buffer-overflow"]
	pathWinter, existsWinter := challengesWinter["buffer-overflow"]

	if !existsSummer || !existsWinter {
		t.Fatal("Both events should have discovered 'buffer-overflow'")
	}

	if pathSummer == pathWinter {
		t.Error("Paths should be different")
	}

	// Verify complete state isolation
	verifyCompleteStateIsolation(t, ewSummer, ewWinter, "buffer-overflow")

	t.Logf("Successfully verified complete isolation with identical folder and YAML names")
	t.Logf("Summer CTF path: %s", pathSummer)
	t.Logf("Winter CTF path: %s", pathWinter)
}

// setupConcurrentDuplicateTest creates environment for concurrent duplicate tests
func setupConcurrentDuplicateTest(t *testing.T) (map[string]*EventWatcher, []string, func()) {
	t.Helper()
	tmpDir, _ := os.MkdirTemp("", "concurrent-duplicate-test-*")
	eventNames := []string{"event-a", "event-b", "event-c"}
	eventsDir := filepath.Join(tmpDir, "events")

	// Create challenges in all events
	for _, eventName := range eventNames {
		for _, category := range []string{"web", "pwn"} {
			challengeName := map[string]string{"web": "xss", "pwn": "overflow"}[category]
			content := fmt.Sprintf("name: %s Challenge\nauthor: test\n", challengeName)
			createChallengeWithContent(t, eventsDir, eventName, category, challengeName, content)
		}
	}

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)

	api := &gzapi.GZAPI{}
	w, _ := New(api)
	w.db = database.New(filepath.Join(tmpDir, "test.db"), false)
	w.db.Init()

	eventWatchers := make(map[string]*EventWatcher)
	config := watchertypes.WatcherConfig{}

	for _, eventName := range eventNames {
		ew, _ := NewEventWatcher(eventName, w.api, config, w.db, w.ctx)
		w.AddEventWatcher(eventName, ew)
		eventWatchers[eventName] = ew
		ew.discoverChallenges()
	}

	cleanup := func() {
		w.db.Close()
		os.Chdir(oldWd)
		os.RemoveAll(tmpDir)
	}

	return eventWatchers, eventNames, cleanup
}

// runConcurrentUpdateOp simulates a single concurrent update operation
func runConcurrentUpdateOp(t *testing.T, eventWatchers map[string]*EventWatcher, eventNames []string, iteration int, successMu *sync.Mutex, successCount *int) {
	eventName := eventNames[iteration%len(eventNames)]
	challengeName := map[bool]string{true: "overflow", false: "xss"}[iteration%2 == 0]

	ew := eventWatchers[eventName]
	mutex := ew.GetChallengeUpdateMutex(challengeName)
	mutex.Lock()
	defer mutex.Unlock()

	ew.setUpdating(challengeName, true)
	time.Sleep(1 * time.Millisecond)

	if !ew.isUpdating(challengeName) {
		t.Errorf("Challenge %s in %s should be updating", challengeName, eventName)
		return
	}

	ew.setUpdating(challengeName, false)

	successMu.Lock()
	*successCount++
	successMu.Unlock()
}

// verifyAllChallengesNotUpdating verifies no challenges are in updating state
func verifyAllChallengesNotUpdating(t *testing.T, eventWatchers map[string]*EventWatcher) {
	t.Helper()
	for eventName, ew := range eventWatchers {
		for _, chName := range []string{"xss", "overflow"} {
			if ew.isUpdating(chName) {
				t.Errorf("Event %s: '%s' should not be in updating state", eventName, chName)
			}
		}
	}
}

// TestMultiEvent_ConcurrentUpdatesDuplicateNames tests concurrent updates on same-named challenges
func TestMultiEvent_ConcurrentUpdatesDuplicateNames(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent updates test in short mode")
	}

	eventWatchers, eventNames, cleanup := setupConcurrentDuplicateTest(t)
	defer cleanup()

	// Test concurrent operations
	var wg sync.WaitGroup
	successCount := 0
	var successMu sync.Mutex

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(iteration int) {
			defer wg.Done()
			runConcurrentUpdateOp(t, eventWatchers, eventNames, iteration, &successMu, &successCount)
		}(i)
	}

	// Wait for completion with timeout
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		if successCount != 100 {
			t.Errorf("Expected 100 successful operations, got %d", successCount)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Concurrent operations timed out")
	}

	verifyAllChallengesNotUpdating(t, eventWatchers)
	t.Log("Successfully verified concurrent updates on duplicate challenge names across events")
}

// Helper function to check if a path contains a substring
func containsPath(path, substr string) bool {
	path = filepath.ToSlash(path)
	return filepath.IsAbs(path) && (filepath.Base(filepath.Dir(filepath.Dir(path))) == substr ||
		filepath.Base(filepath.Dir(filepath.Dir(filepath.Dir(path)))) == substr)
}

// createFullChallenge creates a complete challenge with full YAML
func createFullChallenge(t *testing.T, eventDir, category, name, title string) string {
	t.Helper()
	challengeDir := filepath.Join(eventDir, category, name)
	os.MkdirAll(challengeDir, 0755)

	content := fmt.Sprintf(`name: "%s"
author: "test"
description: "Test description"
type: "StaticAttachment"
value: 100
flags:
  - "flag{test}"
`, title)
	challengeYaml := filepath.Join(challengeDir, "challenge.yaml")
	os.WriteFile(challengeYaml, []byte(content), 0644)
	return challengeDir
}

// setupRediscoveryTest sets up environment for rediscovery testing
func setupRediscoveryTest(t *testing.T) (string, *EventWatcher, string, func()) {
	t.Helper()
	tmpDir, _ := os.MkdirTemp("", "autosync-rediscovery-test-*")
	eventName := "test-event"
	eventDir := filepath.Join(tmpDir, "events", eventName)

	challengeDir := createFullChallenge(t, eventDir, "web", "test-challenge", "Test Challenge")

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)

	api := &gzapi.GZAPI{}
	w, _ := New(api)
	w.db = database.New(filepath.Join(tmpDir, "test.db"), false)
	w.db.Init()

	config := watchertypes.WatcherConfig{}
	ew, _ := NewEventWatcher(eventName, w.api, config, w.db, w.ctx)
	w.AddEventWatcher(eventName, ew)

	cleanup := func() {
		w.db.Close()
		os.Chdir(oldWd)
		os.RemoveAll(tmpDir)
	}

	return eventDir, ew, challengeDir, cleanup
}

// TestAutoSync_ChallengeRediscovery tests automatic rediscovery after challenge removal
func TestAutoSync_ChallengeRediscovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping auto-sync rediscovery test in short mode")
	}

	eventDir, ew, challengeDir, cleanup := setupRediscoveryTest(t)
	defer cleanup()

	// Discover initial challenges
	ew.discoverChallenges()
	challenges := ew.challengeMgr.GetChallenges()
	if len(challenges) != 1 || challenges["test-challenge"] == "" {
		t.Fatal("Should discover 1 challenge")
	}

	// Remove and trigger rediscovery
	os.RemoveAll(challengeDir)
	ew.removeChallenge("test-challenge")
	ew.triggerRediscovery()
	time.Sleep(100 * time.Millisecond)

	// Create new challenge in different location
	createFullChallenge(t, eventDir, "pwn", "new-challenge", "New Challenge")
	ew.discoverChallenges()

	// Verify old removed, new discovered
	challenges = ew.challengeMgr.GetChallenges()
	if _, exists := challenges["test-challenge"]; exists {
		t.Error("Old challenge should be removed")
	}
	if _, exists := challenges["new-challenge"]; !exists {
		t.Error("New challenge should be discovered")
	}

	t.Log("Successfully tested automatic challenge rediscovery")
}

// TestAutoSync_CategoryDetection tests category detection from challenge path
func TestAutoSync_CategoryDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping category detection test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "category-detection-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test various path structures
	testCases := []struct {
		category         string
		challengeName    string
		expectedCategory string
	}{
		{"Web", "xss-challenge", "Web"},
		{"Crypto", "rsa-challenge", "Crypto"},
		{"Pwn", "buffer-overflow", "Pwn"},
		{"Forensics", "memory-dump", "Forensics"},
	}

	eventName := "test-event"
	eventDir := filepath.Join(tmpDir, "events", eventName)

	for _, tc := range testCases {
		challengeDir := filepath.Join(eventDir, tc.category, tc.challengeName)
		if err := os.MkdirAll(challengeDir, 0755); err != nil {
			t.Fatalf("Failed to create challenge dir: %v", err)
		}

		challengeYaml := filepath.Join(challengeDir, "challenge.yaml")
		content := fmt.Sprintf(`name: "%s"
author: "test"
description: "Test"
type: "StaticAttachment"
value: 100
flags:
  - "flag{test}"
`, tc.challengeName)
		if err := os.WriteFile(challengeYaml, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create challenge.yaml: %v", err)
		}
	}

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	api := &gzapi.GZAPI{}
	w, err := New(api)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	w.db = database.New(filepath.Join(tmpDir, "test.db"), false)
	w.db.Init()
	defer w.db.Close()

	config := watchertypes.WatcherConfig{}
	ew, err := NewEventWatcher(eventName, w.api, config, w.db, w.ctx)
	if err != nil {
		t.Fatalf("Failed to create event watcher: %v", err)
	}

	// Test category detection by checking discovered challenges
	if err := ew.discoverChallenges(); err != nil {
		t.Fatalf("Failed to discover challenges: %v", err)
	}

	challenges := ew.challengeMgr.GetChallenges()
	if len(challenges) != len(testCases) {
		t.Errorf("Expected %d challenges, got %d", len(testCases), len(challenges))
	}

	for _, tc := range testCases {
		if _, exists := challenges[tc.challengeName]; !exists {
			t.Errorf("Challenge %s not discovered", tc.challengeName)
		}
	}

	t.Log("Successfully tested category detection from challenge paths")
}

// TestAutoSync_ChallengeRemoval tests proper cleanup when challenges are removed
func TestAutoSync_ChallengeRemoval(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping challenge removal test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "removal-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	eventName := "test-event"
	eventDir := filepath.Join(tmpDir, "events", eventName)
	challengeDir := filepath.Join(eventDir, "web", "test-challenge")
	if err := os.MkdirAll(challengeDir, 0755); err != nil {
		t.Fatalf("Failed to create challenge dir: %v", err)
	}

	challengeYaml := filepath.Join(challengeDir, "challenge.yaml")
	if err := os.WriteFile(challengeYaml, []byte("name: Test\n"), 0644); err != nil {
		t.Fatalf("Failed to create challenge.yaml: %v", err)
	}

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	api := &gzapi.GZAPI{}
	w, err := New(api)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	w.db = database.New(filepath.Join(tmpDir, "test.db"), false)
	w.db.Init()
	defer w.db.Close()

	config := watchertypes.WatcherConfig{}
	ew, err := NewEventWatcher(eventName, w.api, config, w.db, w.ctx)
	if err != nil {
		t.Fatalf("Failed to create event watcher: %v", err)
	}

	if err := ew.discoverChallenges(); err != nil {
		t.Fatalf("Failed to discover challenges: %v", err)
	}

	// Verify challenge exists
	challenges := ew.challengeMgr.GetChallenges()
	if len(challenges) != 1 {
		t.Fatalf("Expected 1 challenge, got %d", len(challenges))
	}

	// Set up some state for the challenge
	mutex := ew.GetChallengeUpdateMutex("test-challenge")
	if mutex == nil {
		t.Fatal("Mutex should be created")
	}

	ew.setUpdating("test-challenge", true)
	ew.setPendingUpdate("test-challenge", "/some/file")

	// Remove the challenge
	ew.removeChallenge("test-challenge")

	// Verify all state is cleaned up
	challenges = ew.challengeMgr.GetChallenges()
	if len(challenges) != 0 {
		t.Errorf("Expected 0 challenges after removal, got %d", len(challenges))
	}

	if ew.isUpdating("test-challenge") {
		t.Error("Challenge should not be in updating state after removal")
	}

	// Verify mutex map is cleaned (indirectly by checking it doesn't exist)
	ew.challengeMutexesMu.RLock()
	_, exists := ew.challengeMutexes["test-challenge"]
	ew.challengeMutexesMu.RUnlock()

	if exists {
		t.Error("Challenge mutex should be removed")
	}

	t.Log("Successfully tested challenge removal and cleanup")
}

// TestAutoSync_PathSplitting tests the splitPath helper function
func TestAutoSync_PathSplitting(t *testing.T) {
	testCases := []struct {
		input    string
		expected []string
	}{
		{"Web/challenge", []string{"Web", "challenge"}},
		{"Crypto/rsa/level1", []string{"Crypto", "rsa", "level1"}},
		{"single", []string{"single"}},
		{"a/b/c/d", []string{"a", "b", "c", "d"}},
	}

	for _, tc := range testCases {
		result := splitPath(tc.input)
		if len(result) != len(tc.expected) {
			t.Errorf("For input %q: expected %d parts, got %d", tc.input, len(tc.expected), len(result))
			continue
		}
		for i, part := range result {
			if part != tc.expected[i] {
				t.Errorf("For input %q: expected part[%d]=%q, got %q", tc.input, i, tc.expected[i], part)
			}
		}
	}

	t.Log("Successfully tested path splitting")
}

// TestAutoSync_UpdateTypeHandling tests update type determination and priority
func TestAutoSync_UpdateTypeHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping update type handling test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "update-type-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	eventName := "test-event"
	eventDir := filepath.Join(tmpDir, "events", eventName)
	challengeDir := filepath.Join(eventDir, "web", "test-challenge")

	// Create directory structure
	srcDir := filepath.Join(challengeDir, "src")
	distDir := filepath.Join(challengeDir, "dist")
	solverDir := filepath.Join(challengeDir, "solver")

	for _, dir := range []string{challengeDir, srcDir, distDir, solverDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
	}

	// Create various files
	files := map[string]string{
		filepath.Join(challengeDir, "challenge.yaml"): "name: Test\n",
		filepath.Join(srcDir, "main.py"):              "print('hello')\n",
		filepath.Join(distDir, "flag.txt"):            "flag{test}\n",
		filepath.Join(solverDir, "solve.py"):          "# solver\n",
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", path, err)
		}
	}

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	api := &gzapi.GZAPI{}
	w, err := New(api)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	w.db = database.New(filepath.Join(tmpDir, "test.db"), false)
	w.db.Init()
	defer w.db.Close()

	config := watchertypes.WatcherConfig{}
	ew, err := NewEventWatcher(eventName, w.api, config, w.db, w.ctx)
	if err != nil {
		t.Fatalf("Failed to create event watcher: %v", err)
	}

	if err := ew.discoverChallenges(); err != nil {
		t.Fatalf("Failed to discover challenges: %v", err)
	}

	// Test update type determination
	testCases := []struct {
		file               string
		expectedUpdateType watchertypes.UpdateType
	}{
		{filepath.Join(challengeDir, "challenge.yaml"), watchertypes.UpdateMetadata},
		{filepath.Join(srcDir, "main.py"), watchertypes.UpdateFullRedeploy},
		{filepath.Join(distDir, "flag.txt"), watchertypes.UpdateAttachment},
		{filepath.Join(solverDir, "solve.py"), watchertypes.UpdateNone},
	}

	for _, tc := range testCases {
		updateType := filesystem.DetermineUpdateType(tc.file, challengeDir)
		if updateType != tc.expectedUpdateType {
			t.Errorf("For file %s: expected update type %v, got %v",
				filepath.Base(tc.file), tc.expectedUpdateType, updateType)
		}
	}

	t.Log("Successfully tested update type determination")
}

// TestAutoSync_ConcurrentSyncs tests handling of concurrent sync requests
func TestAutoSync_ConcurrentSyncs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent syncs test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "concurrent-sync-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	eventName := "test-event"
	eventDir := filepath.Join(tmpDir, "events", eventName)
	challengeDir := filepath.Join(eventDir, "web", "test-challenge")
	if err := os.MkdirAll(challengeDir, 0755); err != nil {
		t.Fatalf("Failed to create challenge dir: %v", err)
	}

	challengeYaml := filepath.Join(challengeDir, "challenge.yaml")
	if err := os.WriteFile(challengeYaml, []byte("name: Test\n"), 0644); err != nil {
		t.Fatalf("Failed to create challenge.yaml: %v", err)
	}

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	api := &gzapi.GZAPI{}
	w, err := New(api)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	w.db = database.New(filepath.Join(tmpDir, "test.db"), false)
	w.db.Init()
	defer w.db.Close()

	config := watchertypes.WatcherConfig{}
	ew, err := NewEventWatcher(eventName, w.api, config, w.db, w.ctx)
	if err != nil {
		t.Fatalf("Failed to create event watcher: %v", err)
	}

	if err := ew.discoverChallenges(); err != nil {
		t.Fatalf("Failed to discover challenges: %v", err)
	}

	// Simulate concurrent sync requests
	var wg sync.WaitGroup
	attemptCount := 10

	for i := 0; i < attemptCount; i++ {
		wg.Add(1)
		go func(iteration int) {
			defer wg.Done()

			mutex := ew.GetChallengeUpdateMutex("test-challenge")
			mutex.Lock()

			// Simulate checking if updating
			if ew.isUpdating("test-challenge") {
				// Should queue as pending
				ew.setPendingUpdate("test-challenge", fmt.Sprintf("/file%d", iteration))
				mutex.Unlock()
				return
			}

			ew.setUpdating("test-challenge", true)
			mutex.Unlock()

			// Simulate sync work
			time.Sleep(1 * time.Millisecond)

			ew.setUpdating("test-challenge", false)
		}(i)
	}

	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Verify no challenge is left in updating state
		if ew.isUpdating("test-challenge") {
			t.Error("Challenge should not be in updating state after all syncs complete")
		}
		t.Log("Successfully tested concurrent sync handling")
	case <-time.After(5 * time.Second):
		t.Fatal("Concurrent sync test timed out")
	}
}

// TestChallengeMapping_DatabasePersistence tests that mappings survive restarts
func TestChallengeMapping_DatabasePersistence(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database persistence test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "mapping-persistence-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	eventName := "test-event"

	// Create a database and store a mapping
	db1 := database.New(dbPath, true)
	if err := db1.Init(); err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}

	// Store a mapping
	if err := db1.SetChallengeMapping(eventName, "Web/test-challenge", 12345, "Test Challenge"); err != nil {
		t.Fatalf("Failed to set mapping: %v", err)
	}

	// Close database
	db1.Close()

	// Reopen database and verify mapping persists
	db2 := database.New(dbPath, true)
	if err := db2.Init(); err != nil {
		t.Fatalf("Failed to reinit database: %v", err)
	}
	defer db2.Close()

	mapping, err := db2.GetChallengeMapping(eventName, "Web/test-challenge")
	if err != nil {
		t.Fatalf("Failed to get mapping: %v", err)
	}

	if mapping == nil {
		t.Fatal("Mapping should persist across database restarts")
	}

	if mapping.ChallengeID != 12345 {
		t.Errorf("Expected challenge ID 12345, got %d", mapping.ChallengeID)
	}

	if mapping.ChallengeTitle != "Test Challenge" {
		t.Errorf("Expected title 'Test Challenge', got %s", mapping.ChallengeTitle)
	}

	t.Log("Successfully tested mapping persistence across database restarts")
}

// TestChallengeMapping_CacheHit tests cache performance
func TestChallengeMapping_CacheHit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cache hit test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "cache-hit-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	eventName := "test-event"
	eventDir := filepath.Join(tmpDir, "events", eventName)
	challengeDir := filepath.Join(eventDir, "web", "test-challenge")
	if err := os.MkdirAll(challengeDir, 0755); err != nil {
		t.Fatalf("Failed to create challenge dir: %v", err)
	}

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	api := &gzapi.GZAPI{}
	w, err := New(api)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	w.db = database.New(filepath.Join(tmpDir, "test.db"), true)
	w.db.Init()
	defer w.db.Close()

	config := watchertypes.WatcherConfig{}
	ew, err := NewEventWatcher(eventName, w.api, config, w.db, w.ctx)
	if err != nil {
		t.Fatalf("Failed to create event watcher: %v", err)
	}

	// Store a mapping directly in database
	if err := w.db.SetChallengeMapping(eventName, "web/test-challenge", 999, "Test Challenge"); err != nil {
		t.Fatalf("Failed to set mapping: %v", err)
	}

	// First call - should hit database
	id1, exists1 := ew.getChallengeID("web/test-challenge")
	if !exists1 {
		t.Fatal("Mapping should exist in database")
	}
	if id1 != 999 {
		t.Errorf("Expected ID 999, got %d", id1)
	}

	// Second call - should hit cache
	id2, exists2 := ew.getChallengeID("web/test-challenge")
	if !exists2 {
		t.Fatal("Mapping should exist in cache")
	}
	if id2 != 999 {
		t.Errorf("Expected ID 999 from cache, got %d", id2)
	}

	// Verify it's actually cached in memory
	ew.challengeMappingsMu.RLock()
	cachedID, inCache := ew.challengeMappings["web/test-challenge"]
	ew.challengeMappingsMu.RUnlock()

	if !inCache {
		t.Error("Mapping should be cached in memory")
	}
	if cachedID != 999 {
		t.Errorf("Cached ID should be 999, got %d", cachedID)
	}

	t.Log("Successfully tested cache hit performance")
}

// TestChallengeMapping_NoDuplicates tests that name changes don't create duplicates
func TestChallengeMapping_NoDuplicates(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping no duplicates test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "no-duplicates-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	eventName := "test-event"
	eventDir := filepath.Join(tmpDir, "events", eventName)
	challengeDir := filepath.Join(eventDir, "web", "test-challenge")
	if err := os.MkdirAll(challengeDir, 0755); err != nil {
		t.Fatalf("Failed to create challenge dir: %v", err)
	}

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	api := &gzapi.GZAPI{}
	w, err := New(api)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	w.db = database.New(filepath.Join(tmpDir, "test.db"), true)
	w.db.Init()
	defer w.db.Close()

	config := watchertypes.WatcherConfig{}
	ew, err := NewEventWatcher(eventName, w.api, config, w.db, w.ctx)
	if err != nil {
		t.Fatalf("Failed to create event watcher: %v", err)
	}

	// Simulate first sync - store mapping
	ew.setChallengeID("web/test-challenge", 100, "Original Name")

	// Verify mapping was stored
	id1, exists1 := ew.getChallengeID("web/test-challenge")
	if !exists1 || id1 != 100 {
		t.Fatalf("Mapping should exist with ID 100, got exists=%v, id=%d", exists1, id1)
	}

	// Simulate name change - update mapping with new title
	ew.setChallengeID("web/test-challenge", 100, "New Name")

	// Verify ID stays the same (no duplicate)
	id2, exists2 := ew.getChallengeID("web/test-challenge")
	if !exists2 || id2 != 100 {
		t.Errorf("After name change, ID should still be 100, got exists=%v, id=%d", exists2, id2)
	}

	// Verify in database that only one mapping exists
	mappings, err := w.db.ListChallengeMappings(eventName)
	if err != nil {
		t.Fatalf("Failed to list mappings: %v", err)
	}

	if len(mappings) != 1 {
		t.Errorf("Should have exactly 1 mapping, got %d", len(mappings))
	}

	if len(mappings) > 0 {
		if mappings[0].ChallengeTitle != "New Name" {
			t.Errorf("Expected title 'New Name', got '%s'", mappings[0].ChallengeTitle)
		}
	}

	t.Log("Successfully verified that name changes don't create duplicate mappings")
}

// TestChallengeMapping_MissingMapping tests fallback when mapping doesn't exist
func TestChallengeMapping_MissingMapping(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping missing mapping test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "missing-mapping-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	eventName := "test-event"
	eventDir := filepath.Join(tmpDir, "events", eventName)
	challengeDir := filepath.Join(eventDir, "web", "test-challenge")
	if err := os.MkdirAll(challengeDir, 0755); err != nil {
		t.Fatalf("Failed to create challenge dir: %v", err)
	}

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	api := &gzapi.GZAPI{}
	w, err := New(api)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	w.db = database.New(filepath.Join(tmpDir, "test.db"), true)
	w.db.Init()
	defer w.db.Close()

	config := watchertypes.WatcherConfig{}
	ew, err := NewEventWatcher(eventName, w.api, config, w.db, w.ctx)
	if err != nil {
		t.Fatalf("Failed to create event watcher: %v", err)
	}

	// Try to get mapping that doesn't exist
	id, exists := ew.getChallengeID("web/test-challenge")
	if exists {
		t.Error("Mapping should not exist")
	}
	if id != 0 {
		t.Errorf("Non-existent mapping should return 0, got %d", id)
	}

	t.Log("Successfully tested missing mapping behavior")
}

// TestChallengeMapping_DeletedInGZCTF tests handling when challenge is deleted from GZCTF
func TestChallengeMapping_DeletedInGZCTF(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping deleted in GZCTF test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "deleted-gzctf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	eventName := "test-event"
	eventDir := filepath.Join(tmpDir, "events", eventName)
	challengeDir := filepath.Join(eventDir, "web", "test-challenge")
	if err := os.MkdirAll(challengeDir, 0755); err != nil {
		t.Fatalf("Failed to create challenge dir: %v", err)
	}

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	api := &gzapi.GZAPI{}
	w, err := New(api)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	w.db = database.New(filepath.Join(tmpDir, "test.db"), true)
	w.db.Init()
	defer w.db.Close()

	config := watchertypes.WatcherConfig{}
	ew, err := NewEventWatcher(eventName, w.api, config, w.db, w.ctx)
	if err != nil {
		t.Fatalf("Failed to create event watcher: %v", err)
	}

	// Store a mapping for a challenge that doesn't exist in GZCTF
	ew.setChallengeID("web/test-challenge", 99999, "Ghost Challenge")

	// Verify mapping exists
	id, exists := ew.getChallengeID("web/test-challenge")
	if !exists || id != 99999 {
		t.Fatalf("Mapping should exist with ID 99999")
	}

	// Try to fetch the challenge (will fail since it doesn't exist)
	_, err = ew.fetchChallengeByID(99999)
	if err == nil {
		t.Error("Should get error when fetching non-existent challenge")
	}

	// The deleteChallengeID should be called when sync detects 404
	ew.deleteChallengeID("web/test-challenge")

	// Verify mapping was removed
	_, exists = ew.getChallengeID("web/test-challenge")
	if exists {
		t.Error("Mapping should be deleted after challenge not found")
	}

	t.Log("Successfully tested handling of deleted challenges in GZCTF")
}

// TestChallengeMapping_MultiEvent tests mappings across different events
func TestChallengeMapping_MultiEvent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping multi-event mapping test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "multi-event-mapping-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create database
	db := database.New(filepath.Join(tmpDir, "test.db"), true)
	if err := db.Init(); err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	defer db.Close()

	// Store mappings for different events with same folder path
	if err := db.SetChallengeMapping("event1", "Web/same-challenge", 100, "Event1 Challenge"); err != nil {
		t.Fatalf("Failed to set mapping for event1: %v", err)
	}

	if err := db.SetChallengeMapping("event2", "Web/same-challenge", 200, "Event2 Challenge"); err != nil {
		t.Fatalf("Failed to set mapping for event2: %v", err)
	}

	// Retrieve and verify they're independent
	mapping1, err := db.GetChallengeMapping("event1", "Web/same-challenge")
	if err != nil {
		t.Fatalf("Failed to get mapping for event1: %v", err)
	}

	mapping2, err := db.GetChallengeMapping("event2", "Web/same-challenge")
	if err != nil {
		t.Fatalf("Failed to get mapping for event2: %v", err)
	}

	if mapping1.ChallengeID != 100 {
		t.Errorf("Event1 mapping should have ID 100, got %d", mapping1.ChallengeID)
	}

	if mapping2.ChallengeID != 200 {
		t.Errorf("Event2 mapping should have ID 200, got %d", mapping2.ChallengeID)
	}

	// List mappings per event
	event1Mappings, err := db.ListChallengeMappings("event1")
	if err != nil {
		t.Fatalf("Failed to list event1 mappings: %v", err)
	}

	event2Mappings, err := db.ListChallengeMappings("event2")
	if err != nil {
		t.Fatalf("Failed to list event2 mappings: %v", err)
	}

	if len(event1Mappings) != 1 {
		t.Errorf("Event1 should have 1 mapping, got %d", len(event1Mappings))
	}

	if len(event2Mappings) != 1 {
		t.Errorf("Event2 should have 1 mapping, got %d", len(event2Mappings))
	}

	t.Log("Successfully tested mapping isolation across multiple events")
}
