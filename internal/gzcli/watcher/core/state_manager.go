package core

import (
	"sync"
	"time"
)

// StateManager manages the state of the event watcher
type StateManager struct {
	challengeMutexes   map[string]*sync.Mutex
	challengeMutexesMu sync.RWMutex
	pendingUpdates     map[string]string // challengeName -> latest file path
	pendingUpdatesMu   sync.RWMutex
	updatingChallenges map[string]bool // challengeName -> is updating
	updatingMu         sync.RWMutex
	debounceTimers     map[string]*time.Timer
	debounceMu         sync.Mutex
}

// NewStateManager creates a new state manager
func NewStateManager() *StateManager {
	return &StateManager{
		challengeMutexes:   make(map[string]*sync.Mutex),
		pendingUpdates:     make(map[string]string),
		updatingChallenges: make(map[string]bool),
		debounceTimers:     make(map[string]*time.Timer),
	}
}

// GetChallengeMutex gets or creates a mutex for a challenge
func (sm *StateManager) GetChallengeMutex(challengeName string) *sync.Mutex {
	sm.challengeMutexesMu.Lock()
	defer sm.challengeMutexesMu.Unlock()
	
	if sm.challengeMutexes[challengeName] == nil {
		sm.challengeMutexes[challengeName] = &sync.Mutex{}
	}
	return sm.challengeMutexes[challengeName]
}

// SetPendingUpdate sets a pending update for a challenge
func (sm *StateManager) SetPendingUpdate(challengeName, filePath string) {
	sm.pendingUpdatesMu.Lock()
	defer sm.pendingUpdatesMu.Unlock()
	sm.pendingUpdates[challengeName] = filePath
}

// GetPendingUpdate gets a pending update for a challenge
func (sm *StateManager) GetPendingUpdate(challengeName string) (string, bool) {
	sm.pendingUpdatesMu.RLock()
	defer sm.pendingUpdatesMu.RUnlock()
	path, exists := sm.pendingUpdates[challengeName]
	return path, exists
}

// ClearPendingUpdate clears a pending update for a challenge
func (sm *StateManager) ClearPendingUpdate(challengeName string) {
	sm.pendingUpdatesMu.Lock()
	defer sm.pendingUpdatesMu.Unlock()
	delete(sm.pendingUpdates, challengeName)
}

// SetUpdating sets the updating state for a challenge
func (sm *StateManager) SetUpdating(challengeName string, updating bool) {
	sm.updatingMu.Lock()
	defer sm.updatingMu.Unlock()
	sm.updatingChallenges[challengeName] = updating
}

// IsUpdating checks if a challenge is currently updating
func (sm *StateManager) IsUpdating(challengeName string) bool {
	sm.updatingMu.RLock()
	defer sm.updatingMu.RUnlock()
	return sm.updatingChallenges[challengeName]
}

// SetDebounceTimer sets a debounce timer for a challenge
func (sm *StateManager) SetDebounceTimer(challengeName string, timer *time.Timer) {
	sm.debounceMu.Lock()
	defer sm.debounceMu.Unlock()
	
	// Cancel existing timer if any
	if existingTimer, exists := sm.debounceTimers[challengeName]; exists {
		existingTimer.Stop()
	}
	
	sm.debounceTimers[challengeName] = timer
}

// ClearDebounceTimer clears a debounce timer for a challenge
func (sm *StateManager) ClearDebounceTimer(challengeName string) {
	sm.debounceMu.Lock()
	defer sm.debounceMu.Unlock()
	
	if timer, exists := sm.debounceTimers[challengeName]; exists {
		timer.Stop()
		delete(sm.debounceTimers, challengeName)
	}
}

// ClearAllDebounceTimers clears all debounce timers
func (sm *StateManager) ClearAllDebounceTimers() {
	sm.debounceMu.Lock()
	defer sm.debounceMu.Unlock()
	
	for _, timer := range sm.debounceTimers {
		timer.Stop()
	}
	sm.debounceTimers = make(map[string]*time.Timer)
}