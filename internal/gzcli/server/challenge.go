package server

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/utils"
	"github.com/dimasma0305/gzcli/internal/log"
)

var challengeFileRegex = regexp.MustCompile(`challenge\.(yaml|yml)$`)

// ChallengeManager manages all discovered challenges
type ChallengeManager struct {
	challenges map[string]*ChallengeInfo // slug -> ChallengeInfo
	mu         sync.RWMutex
}

// NewChallengeManager creates a new challenge manager
func NewChallengeManager() *ChallengeManager {
	return &ChallengeManager{
		challenges: make(map[string]*ChallengeInfo),
	}
}

// portParser is used to extract ports from configuration files
var portParser = NewPortParser()

// processChallengeFile processes a single challenge file and adds it to the manager
func (cm *ChallengeManager) processChallengeFile(path, eventName, category string) error {
	var challYaml config.ChallengeYaml
	if err := utils.ParseYamlFromFile(path, &challYaml); err != nil {
		return fmt.Errorf("failed to parse: %w", err)
	}

	// Only include challenges with Dashboard configuration
	if challYaml.Dashboard == nil {
		return nil
	}

	// Set challenge metadata
	challYaml.Category = category
	challYaml.Cwd = filepath.Dir(path)

	// Generate slug
	slug := config.GenerateSlug(eventName, category, challYaml.Name)

	// Parse ports from configuration file
	ports := portParser.ParsePorts(
		challYaml.Dashboard.Type,
		challYaml.Dashboard.Config,
		challYaml.Cwd,
	)

	// Convert to our Dashboard type
	dashboard := &Dashboard{
		Type:   challYaml.Dashboard.Type,
		Config: challYaml.Dashboard.Config,
		Ports:  ports,
	}

	// Create ChallengeInfo
	challengeInfo := &ChallengeInfo{
		Slug:         slug,
		EventName:    eventName,
		Category:     category,
		Name:         challYaml.Name,
		Description:  challYaml.Description,
		Cwd:          challYaml.Cwd,
		Dashboard:    dashboard,
		Scripts:      challYaml.Scripts,
		Status:       StatusStopped,
		ConnectedIPs: make(map[string]bool),
	}

	// Add to manager
	cm.mu.Lock()
	cm.challenges[slug] = challengeInfo
	cm.mu.Unlock()

	log.InfoH3("Discovered: %s (slug: %s)", challYaml.Name, slug)
	return nil
}

// scanCategory scans a category directory for challenges
func (cm *ChallengeManager) scanCategory(eventPath, eventName, category string) int {
	categoryPath := filepath.Join(eventPath, category)

	if _, err := os.Stat(categoryPath); os.IsNotExist(err) {
		return 0
	}

	count := 0
	err := filepath.Walk(categoryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !challengeFileRegex.MatchString(info.Name()) {
			return nil
		}

		if err := cm.processChallengeFile(path, eventName, category); err != nil {
			log.Error("Failed to process %s: %v", path, err)
			return nil
		}

		count++
		return nil
	})

	if err != nil {
		log.Error("Error walking category %s: %v", category, err)
	}

	return count
}

// scanEvent scans an event for all challenges
func (cm *ChallengeManager) scanEvent(eventName string) (int, error) {
	eventPath, err := config.GetEventPath(eventName)
	if err != nil {
		return 0, fmt.Errorf("failed to get event path: %w", err)
	}

	log.InfoH2("Scanning event: %s", eventName)

	count := 0
	for _, category := range config.CHALLENGE_CATEGORY {
		count += cm.scanCategory(eventPath, eventName, category)
	}

	return count, nil
}

// validateEventsDirectory checks if events directory exists and has events
func validateEventsDirectory() ([]string, error) {
	workspaceRoot, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	eventsDir := filepath.Join(workspaceRoot, config.EVENTS_DIR)

	if _, err := os.Stat(eventsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("events directory not found: %s", eventsDir)
	}

	events, err := config.ListEvents()
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("no events found in %s", eventsDir)
	}

	return events, nil
}

// DiscoverChallenges discovers all challenges with dashboard configuration across all events
func (cm *ChallengeManager) DiscoverChallenges() error {
	log.Info("Discovering challenges with dashboard configuration...")

	events, err := validateEventsDirectory()
	if err != nil {
		return err
	}

	discoveredCount := 0

	// Iterate through each event
	for _, eventName := range events {
		count, err := cm.scanEvent(eventName)
		if err != nil {
			log.Error("Failed to scan event %s: %v", eventName, err)
			continue
		}
		discoveredCount += count
	}

	log.Info("Discovered %d challenge(s) with launcher configuration", discoveredCount)

	if discoveredCount == 0 {
		return fmt.Errorf("no challenges with dashboard configuration found")
	}

	return nil
}

// GetChallenge retrieves a challenge by slug
func (cm *ChallengeManager) GetChallenge(slug string) (*ChallengeInfo, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	challenge, ok := cm.challenges[slug]
	return challenge, ok
}

// ListChallenges returns all discovered challenges
func (cm *ChallengeManager) ListChallenges() []*ChallengeInfo {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	challenges := make([]*ChallengeInfo, 0, len(cm.challenges))
	for _, challenge := range cm.challenges {
		challenges = append(challenges, challenge)
	}

	return challenges
}

// GetChallengeCount returns the number of discovered challenges
func (cm *ChallengeManager) GetChallengeCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.challenges)
}
