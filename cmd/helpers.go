package cmd

import (
	"fmt"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/log"
)

// ResolveTargetEvents determines which events to operate on based on flags
// Priority:
//  1. If eventFlags provided: use only those specific events
//  2. If excludeFlags provided: use all events except excluded ones
//  3. Otherwise: use all available events
//
// Returns error if no events exist or if specified events don't exist
func ResolveTargetEvents(eventFlags []string, excludeFlags []string) ([]string, error) {
	// Get all available events
	allEvents, err := config.ListEvents()
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	if len(allEvents) == 0 {
		return nil, fmt.Errorf("no events found. Run 'gzcli event create <name>' to create one")
	}

	// If specific events are requested, use only those
	if len(eventFlags) > 0 {
		// Validate that all requested events exist
		for _, requestedEvent := range eventFlags {
			if !contains(allEvents, requestedEvent) {
				return nil, fmt.Errorf("event '%s' does not exist", requestedEvent)
			}
		}
		return eventFlags, nil
	}

	// If exclude flags are provided, filter them out
	if len(excludeFlags) > 0 {
		var filteredEvents []string
		for _, event := range allEvents {
			if !contains(excludeFlags, event) {
				filteredEvents = append(filteredEvents, event)
			}
		}

		if len(filteredEvents) == 0 {
			return nil, fmt.Errorf("all events were excluded, no events to process")
		}

		log.Info("Excluding events: %v", excludeFlags)
		return filteredEvents, nil
	}

	// Default: use all available events
	return allEvents, nil
}

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

