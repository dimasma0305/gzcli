//nolint:errcheck,gosec,revive // Test file with acceptable error handling patterns
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
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/types"
)

// setupMultiEventTest creates a test environment with multiple events
func setupMultiEventTest(t *testing.T, eventNames []string) (string, *Watcher, func()) {
	tmpDir, err := os.MkdirTemp("", "multi-event-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create events directory structure
	eventsDir := filepath.Join(tmpDir, "events")
	for _, eventName := range eventNames {
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

			config := types.WatcherConfig{}
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
	config := types.WatcherConfig{}
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
	config := types.WatcherConfig{}
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
	config := types.WatcherConfig{}
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
	config := types.WatcherConfig{}
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

	config := types.WatcherConfig{}

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

	config := types.WatcherConfig{
		DatabaseEnabled: true,
		SocketEnabled:   true,
	}
	w.config = config

	ew1, _ := NewEventWatcher("event1", w.api, config, w.db, w.ctx)
	ew2, _ := NewEventWatcher("event2", w.api, config, w.db, w.ctx)

	w.AddEventWatcher("event1", ew1)
	w.AddEventWatcher("event2", ew2)

	// Test status command without event filter (should return all events)
	cmd := types.WatcherCommand{
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
	cmdFiltered := types.WatcherCommand{
		Action: "status",
		Event:  "event1",
	}
	responseFiltered := w.HandleStatusCommand(cmdFiltered)

	if !responseFiltered.Success {
		t.Errorf("Filtered status command failed: %s", responseFiltered.Error)
	}

	dataFiltered, ok := responseFiltered.Data["events"].([]string)
	if !ok {
		t.Error("Events field not found in filtered response")
	} else if len(dataFiltered) != 1 {
		t.Errorf("Expected 1 event in filtered status, got %d", len(dataFiltered))
	} else if dataFiltered[0] != "event1" {
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

	config := types.WatcherConfig{}

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

	config := types.WatcherConfig{}
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

	config := types.WatcherConfig{}
	ew, _ := NewEventWatcher("event1", w.api, config, w.db, w.ctx)
	w.AddEventWatcher("event1", ew)

	// Test stop event command
	cmd := types.WatcherCommand{
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
	config := types.WatcherConfig{
		Events: []string{},
	}
	w.config = config

	// GetWatchedChallenges should handle empty event list gracefully
	challenges := w.GetWatchedChallenges()
	if len(challenges) != 0 {
		t.Errorf("Expected 0 challenges with no events, got %d", len(challenges))
	}

	// HandleStatusCommand should work with no events
	cmd := types.WatcherCommand{Action: "status"}
	response := w.HandleStatusCommand(cmd)

	if !response.Success {
		t.Error("Status command should succeed even with no events")
	}

	t.Log("Empty event list handled gracefully")
}
