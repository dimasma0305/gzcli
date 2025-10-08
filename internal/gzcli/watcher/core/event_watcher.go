//nolint:revive // EventWatcher methods follow interface patterns
package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/challenge"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/database"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/filesystem"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/git"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/scripts"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/types"
	"github.com/dimasma0305/gzcli/internal/log"
)

var challengeFileRegex = regexp.MustCompile(`^challenge\.(yaml|yml)$`)

// EventWatcher manages file watching for a single event
type EventWatcher struct {
	eventName string
	eventPath string
	api       *gzapi.GZAPI

	watcher            *fsnotify.Watcher
	config             types.WatcherConfig
	ctx                context.Context
	cancel             context.CancelFunc
	wg                 sync.WaitGroup
	challengeMutexes   map[string]*sync.Mutex
	challengeMutexesMu sync.RWMutex
	pendingUpdates     map[string]string // challengeName -> latest file path
	pendingUpdatesMu   sync.RWMutex
	updatingChallenges map[string]bool // challengeName -> is updating
	updatingMu         sync.RWMutex

	// Component managers
	challengeMgr *challenge.Manager
	scriptMgr    *scripts.Manager
	db           *database.DB // Shared reference
	gitMgr       *git.Manager

	// Additional state
	debounceTimers map[string]*time.Timer
}

// NewEventWatcher creates a new event-specific watcher
func NewEventWatcher(eventName string, api *gzapi.GZAPI, config types.WatcherConfig, db *database.DB, parentCtx context.Context) (*EventWatcher, error) {
	if api == nil {
		return nil, fmt.Errorf("API client cannot be nil")
	}
	if eventName == "" {
		return nil, fmt.Errorf("event name cannot be empty")
	}

	// Get current working directory and construct event path
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	eventPath := filepath.Join(cwd, "events", eventName)
	if _, err := os.Stat(eventPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("event directory does not exist: %s", eventPath)
	}

	// Create fsnotify watcher for this event
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}

	ctx, cancel := context.WithCancel(parentCtx)

	ew := &EventWatcher{
		eventName:          eventName,
		eventPath:          eventPath,
		api:                api,
		watcher:            watcher,
		config:             config,
		ctx:                ctx,
		cancel:             cancel,
		db:                 db,
		debounceTimers:     make(map[string]*time.Timer),
		challengeMutexes:   make(map[string]*sync.Mutex),
		pendingUpdates:     make(map[string]string),
		updatingChallenges: make(map[string]bool),
	}

	// Initialize component managers
	ew.challengeMgr = challenge.NewManager(watcher)
	ew.scriptMgr = scripts.NewManager(ctx, ew)

	return ew, nil
}

// Start starts watching the event
func (ew *EventWatcher) Start() error {
	log.InfoH2("Starting watcher for event: %s", ew.eventName)

	// Initialize git manager if enabled
	if ew.config.GitPullEnabled {
		ew.gitMgr = git.NewManager(ew.eventPath, ew.config.GitPullInterval, func() {
			log.Info("[%s] Git pull completed, checking for new challenges...", ew.eventName)
			// Re-discover challenges after git pull
			if err := ew.discoverChallenges(); err != nil {
				log.Error("[%s] Failed to rediscover challenges after git pull: %v", ew.eventName, err)
			}
		})
	}

	// Discover and watch challenges
	if err := ew.discoverChallenges(); err != nil {
		return fmt.Errorf("failed to discover challenges: %w", err)
	}

	// Start file system watcher loop
	ew.wg.Add(1)
	go func() {
		defer ew.wg.Done()
		done := make(chan struct{})
		go func() {
			<-ew.ctx.Done()
			close(done)
		}()
		filesystem.WatchLoop(ew.watcher, ew.config, ew, done)
	}()

	// Start git pull loop if enabled
	if ew.config.GitPullEnabled && ew.gitMgr != nil {
		ew.wg.Add(1)
		go func() {
			defer ew.wg.Done()
			ew.gitMgr.StartPullLoop(ew.ctx)
		}()
	}

	ew.LogToDatabase("INFO", "event_watcher", "", "", fmt.Sprintf("Event watcher started for %s", ew.eventName), "", 0)
	log.Info("[%s] Event watcher started successfully", ew.eventName)

	return nil
}

// Stop stops the event watcher
func (ew *EventWatcher) Stop() error {
	log.Info("[%s] Stopping event watcher...", ew.eventName)

	ew.LogToDatabase("INFO", "event_watcher", "", "", fmt.Sprintf("Event watcher shutdown initiated for %s", ew.eventName), "", 0)

	// Stop all interval scripts
	if ew.scriptMgr != nil {
		ew.scriptMgr.StopAllScripts(5 * time.Second)
	}

	// Cancel context
	ew.cancel()

	// Wait for goroutines with timeout
	done := make(chan struct{})
	go func() {
		ew.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.InfoH3("[%s] All goroutines finished", ew.eventName)
	case <-time.After(10 * time.Second):
		log.Error("[%s] Timeout waiting for goroutines to finish", ew.eventName)
	}

	// Close file system watcher
	if ew.watcher != nil {
		if err := ew.watcher.Close(); err != nil {
			log.Error("[%s] Failed to close file watcher: %v", ew.eventName, err)
		}
	}

	ew.LogToDatabase("INFO", "event_watcher", "", "", fmt.Sprintf("Event watcher stopped for %s", ew.eventName), "", 0)
	log.Info("[%s] Event watcher stopped", ew.eventName)

	return nil
}

// discoverChallenges walks the event directory to find and watch all challenges
func (ew *EventWatcher) discoverChallenges() error {
	log.InfoH3("[%s] Discovering challenges in %s", ew.eventName, ew.eventPath)

	var discoveredCount int
	err := filepath.Walk(ew.eventPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip if it's a directory or not a challenge file
		if info.IsDir() || !challengeFileRegex.MatchString(info.Name()) {
			return nil
		}

		// Found a challenge.yaml file
		challengeDir := filepath.Dir(path)
		challengeName := filepath.Base(challengeDir)

		// Add challenge to watcher
		if err := ew.challengeMgr.AddChallenge(challengeName, challengeDir); err != nil {
			log.Error("[%s] Failed to add challenge %s: %v", ew.eventName, challengeName, err)
			return nil // Continue with other challenges
		}

		discoveredCount++
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk event directory: %w", err)
	}

	log.Info("[%s] Discovered %d challenge(s)", ew.eventName, discoveredCount)
	return nil
}

// GetChallengeUpdateMutex gets or creates a mutex for a specific challenge
func (ew *EventWatcher) GetChallengeUpdateMutex(challengeName string) *sync.Mutex {
	ew.challengeMutexesMu.RLock()
	if mutex, exists := ew.challengeMutexes[challengeName]; exists {
		ew.challengeMutexesMu.RUnlock()
		return mutex
	}
	ew.challengeMutexesMu.RUnlock()

	// Need to create new mutex
	ew.challengeMutexesMu.Lock()
	defer ew.challengeMutexesMu.Unlock()

	// Double-check in case another goroutine created it
	if mutex, exists := ew.challengeMutexes[challengeName]; exists {
		return mutex
	}

	// Create new mutex
	mutex := &sync.Mutex{}
	ew.challengeMutexes[challengeName] = mutex
	return mutex
}

// Implement ScriptLogger interface for scripts package
func (ew *EventWatcher) LogToDatabase(level, component, challenge, script, message, errorMsg string, duration int64) {
	if ew.db != nil {
		ew.db.LogToDatabase(level, component, challenge, script, message, errorMsg, duration)
	}
}

func (ew *EventWatcher) LogScriptExecution(challengeName, scriptName, scriptType, command, status string, duration int64, output, errorOutput string, exitCode int) {
	if ew.db != nil {
		ew.db.LogScriptExecution(challengeName, scriptName, scriptType, command, status, duration, output, errorOutput, exitCode)
	}
}

func (ew *EventWatcher) UpdateChallengeState(challengeName, status, errorMessage string, activeScripts map[string][]string) {
	if ew.db != nil {
		ew.db.UpdateChallengeState(challengeName, status, errorMessage, activeScripts)
	}
}

// Implement filesystem.EventHandler interface
func (ew *EventWatcher) HandleFileChange(filePath string) {
	log.InfoH2("[%s] Processing file change: %s", ew.eventName, filePath)

	// Find which challenge this file belongs to
	challengeName, challengeCwd, err := ew.challengeMgr.FindChallengeForFile(filePath)
	if err != nil {
		log.Error("[%s] Failed to find challenge for file %s: %v", ew.eventName, filePath, err)
		return
	}

	if challengeName == "" {
		log.InfoH3("[%s] File %s doesn't belong to any challenge", ew.eventName, filePath)
		return
	}

	log.Info("[%s] File %s belongs to challenge: %s", ew.eventName, filePath, challengeName)

	// Use the challenge-specific mutex to prevent race conditions during update checks
	challengeMutex := ew.GetChallengeUpdateMutex(challengeName)
	challengeMutex.Lock()

	// Check if this challenge is already being updated
	if ew.isUpdating(challengeName) {
		log.InfoH3("[%s] Challenge %s is already being updated, setting as pending", ew.eventName, challengeName)
		ew.setPendingUpdate(challengeName, filePath)
		challengeMutex.Unlock()
		return
	}

	// Mark as updating before releasing the mutex
	ew.setUpdating(challengeName, true)
	challengeMutex.Unlock()

	// Process update
	go func() {
		defer ew.setUpdating(challengeName, false)

		updateType := filesystem.DetermineUpdateType(filePath, challengeCwd)
		log.Info("[%s] Update type for %s: %v", ew.eventName, challengeName, updateType)

		// Check for pending updates
		if pendingFilePath, hasPending := ew.getPendingUpdate(challengeName); hasPending {
			log.InfoH3("[%s] Found pending update for %s, would process: %s", ew.eventName, challengeName, pendingFilePath)
		}
	}()
}

func (ew *EventWatcher) HandleFileRemoval(filePath string) {
	log.InfoH2("[%s] Processing file removal: %s", ew.eventName, filePath)
	// Simplified implementation
}

func (ew *EventWatcher) HandleChallengeRemovalByDir(removedDir string) {
	log.InfoH2("[%s] Processing challenge removal by directory: %s", ew.eventName, removedDir)
	// Simplified implementation
}

// Helper methods for update state management
func (ew *EventWatcher) isUpdating(challengeName string) bool {
	ew.updatingMu.RLock()
	defer ew.updatingMu.RUnlock()
	return ew.updatingChallenges[challengeName]
}

func (ew *EventWatcher) setUpdating(challengeName string, updating bool) {
	ew.updatingMu.Lock()
	defer ew.updatingMu.Unlock()
	if updating {
		ew.updatingChallenges[challengeName] = true
	} else {
		delete(ew.updatingChallenges, challengeName)
	}
}

func (ew *EventWatcher) setPendingUpdate(challengeName, filePath string) {
	ew.pendingUpdatesMu.Lock()
	defer ew.pendingUpdatesMu.Unlock()
	ew.pendingUpdates[challengeName] = filePath
}

func (ew *EventWatcher) getPendingUpdate(challengeName string) (string, bool) {
	ew.pendingUpdatesMu.Lock()
	defer ew.pendingUpdatesMu.Unlock()
	filePath, exists := ew.pendingUpdates[challengeName]
	if exists {
		delete(ew.pendingUpdates, challengeName)
	}
	return filePath, exists
}

// GetWatchedChallenges returns the list of challenges being watched by this event watcher
func (ew *EventWatcher) GetWatchedChallenges() []string {
	challenges := ew.challengeMgr.GetChallenges()
	dirs := make([]string, 0, len(challenges))
	for dir := range challenges {
		dirs = append(dirs, dir)
	}
	return dirs
}

// GetEventName returns the event name
func (ew *EventWatcher) GetEventName() string {
	return ew.eventName
}

// GetScriptManager returns the script manager for this event
func (ew *EventWatcher) GetScriptManager() *scripts.Manager {
	return ew.scriptMgr
}
