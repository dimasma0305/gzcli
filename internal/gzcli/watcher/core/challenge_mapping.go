package core

import (
	"sync"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/log"
)

// ChallengeMapping manages the mapping between folder paths and challenge IDs
type ChallengeMapping struct {
	mappings   map[string]int // folderPath -> challengeID
	mappingsMu sync.RWMutex
}

// NewChallengeMapping creates a new challenge mapping manager
func NewChallengeMapping() *ChallengeMapping {
	return &ChallengeMapping{
		mappings: make(map[string]int),
	}
}

// SetChallengeID sets the challenge ID for a folder path
func (cm *ChallengeMapping) SetChallengeID(folderPath string, challengeID int, challengeName string) {
	cm.mappingsMu.Lock()
	defer cm.mappingsMu.Unlock()
	cm.mappings[folderPath] = challengeID
	log.Debug("Mapped folder %s to challenge %s (ID: %d)", folderPath, challengeName, challengeID)
}

// GetChallengeID gets the challenge ID for a folder path
func (cm *ChallengeMapping) GetChallengeID(folderPath string) (int, bool) {
	cm.mappingsMu.RLock()
	defer cm.mappingsMu.RUnlock()
	challengeID, exists := cm.mappings[folderPath]
	return challengeID, exists
}

// RemoveChallengeID removes the challenge ID for a folder path
func (cm *ChallengeMapping) RemoveChallengeID(folderPath string) {
	cm.mappingsMu.Lock()
	defer cm.mappingsMu.Unlock()
	delete(cm.mappings, folderPath)
	log.Debug("Removed mapping for folder %s", folderPath)
}

// FindChallengeByID finds a challenge by ID in the challenges list
func (cm *ChallengeMapping) FindChallengeByID(challenges []config.ChallengeYaml, challengeID int) *config.ChallengeYaml {
	for i := range challenges {
		// This would need to be implemented based on how challenge IDs are stored
		// For now, return nil as this is a placeholder
		_ = i
	}
	return nil
}

// GetAllMappings returns all current mappings
func (cm *ChallengeMapping) GetAllMappings() map[string]int {
	cm.mappingsMu.RLock()
	defer cm.mappingsMu.RUnlock()
	
	// Return a copy to avoid race conditions
	mappings := make(map[string]int)
	for k, v := range cm.mappings {
		mappings[k] = v
	}
	return mappings
}

// ClearAllMappings clears all mappings
func (cm *ChallengeMapping) ClearAllMappings() {
	cm.mappingsMu.Lock()
	defer cm.mappingsMu.Unlock()
	cm.mappings = make(map[string]int)
	log.Debug("Cleared all challenge mappings")
}