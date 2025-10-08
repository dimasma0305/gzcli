//nolint:revive // Config struct field names match YAML/API structure
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/utils"
	"github.com/dimasma0305/gzcli/internal/log"
)

const (
	EVENTS_DIR   = "events"
	GZEVENT_FILE = ".gzevent"
)

// EventConfig represents event-specific configuration
type EventConfig struct {
	Name string // Event name (directory name)
	gzapi.Game
}

// GetEventConfig reads event configuration from events/[name]/.gzevent
func GetEventConfig(eventName string) (*EventConfig, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	eventDir := filepath.Join(dir, EVENTS_DIR, eventName)
	eventPath := filepath.Join(eventDir, GZEVENT_FILE)
	var game gzapi.Game
	if err := utils.ParseYamlFromFile(eventPath, &game); err != nil {
		return nil, fmt.Errorf("failed to read event config %s: %w", eventPath, err)
	}

	// Resolve relative paths in the event config
	// If poster path is relative, resolve it relative to the event directory
	if game.Poster != "" && !filepath.IsAbs(game.Poster) {
		resolvedPoster := filepath.Join(eventDir, game.Poster)
		// Check if the resolved path exists, if not keep the original
		if _, err := os.Stat(resolvedPoster); err == nil {
			game.Poster = resolvedPoster
		} else {
			// Try resolving from workspace root
			rootPoster := filepath.Join(dir, game.Poster)
			if _, err := os.Stat(rootPoster); err == nil {
				game.Poster = rootPoster
			}
		}
	}

	return &EventConfig{
		Name: eventName,
		Game: game,
	}, nil
}

// GetEnvEvent returns the GZCLI_EVENT environment variable
func GetEnvEvent() string {
	return os.Getenv("GZCLI_EVENT")
}

// GetCurrentEvent determines the active event from:
// 1. Command-line flag (--event)
// 2. Environment variable (GZCLI_EVENT)
// 3. Default event file (.gzcli/current-event)
// 4. Single event if only one exists
// Returns error if multiple events exist without selection
func GetCurrentEvent(eventFlag string) (string, error) {
	// 1. Check command-line flag
	if eventFlag != "" {
		return eventFlag, nil
	}

	// 2. Check environment variable
	if envEvent := GetEnvEvent(); envEvent != "" {
		return envEvent, nil
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// 3. Check default event file
	defaultEventFile := filepath.Join(dir, ".gzcli", "current-event")
	//nolint:gosec // G304: Path is constructed from working directory
	if data, err := os.ReadFile(defaultEventFile); err == nil {
		eventName := string(data)
		if eventName != "" {
			return eventName, nil
		}
	}

	// 4. Auto-detect if only one event exists
	eventsDir := filepath.Join(dir, EVENTS_DIR)
	entries, err := os.ReadDir(eventsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("no events directory found. Run 'gzcli init' to create one")
		}
		return "", fmt.Errorf("failed to read events directory: %w", err)
	}

	// Filter for directories only
	var eventDirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if .gzevent file exists
			gzeventPath := filepath.Join(eventsDir, entry.Name(), GZEVENT_FILE)
			if _, err := os.Stat(gzeventPath); err == nil {
				eventDirs = append(eventDirs, entry.Name())
			}
		}
	}

	if len(eventDirs) == 0 {
		return "", fmt.Errorf("no events found in %s. Run 'gzcli event create <name>' to create one", eventsDir)
	}

	if len(eventDirs) == 1 {
		log.Info("Auto-selected event: %s", eventDirs[0])
		return eventDirs[0], nil
	}

	// Multiple events without selection
	return "", fmt.Errorf("multiple events found: %v. Please specify with --event flag or set GZCLI_EVENT environment variable", eventDirs)
}

// SetCurrentEvent sets the default event in .gzcli/current-event
func SetCurrentEvent(eventName string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	// Verify event exists
	eventPath := filepath.Join(dir, EVENTS_DIR, eventName, GZEVENT_FILE)
	if _, err := os.Stat(eventPath); err != nil {
		return fmt.Errorf("event %s does not exist", eventName)
	}

	// Create .gzcli directory if it doesn't exist
	gzcliDir := filepath.Join(dir, ".gzcli")
	if err := os.MkdirAll(gzcliDir, 0750); err != nil {
		return fmt.Errorf("failed to create .gzcli directory: %w", err)
	}

	// Write current event file
	defaultEventFile := filepath.Join(gzcliDir, "current-event")
	if err := os.WriteFile(defaultEventFile, []byte(eventName), 0600); err != nil {
		return fmt.Errorf("failed to write current event: %w", err)
	}

	return nil
}

// ListEvents returns all available events
func ListEvents() ([]string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	eventsDir := filepath.Join(dir, EVENTS_DIR)
	entries, err := os.ReadDir(eventsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read events directory: %w", err)
	}

	var events []string
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if .gzevent file exists
			gzeventPath := filepath.Join(eventsDir, entry.Name(), GZEVENT_FILE)
			if _, err := os.Stat(gzeventPath); err == nil {
				events = append(events, entry.Name())
			}
		}
	}

	return events, nil
}

// GetEventPath returns the absolute path to an event directory
func GetEventPath(eventName string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	eventPath := filepath.Join(dir, EVENTS_DIR, eventName)
	if _, err := os.Stat(eventPath); err != nil {
		return "", fmt.Errorf("event %s does not exist: %w", eventName, err)
	}

	return eventPath, nil
}
