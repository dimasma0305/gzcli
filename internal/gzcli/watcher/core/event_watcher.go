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

	challengepkg "github.com/dimasma0305/gzcli/internal/gzcli/challenge"
	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/fileutil"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/challenge"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/database"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/filesystem"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/git"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/scripts"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/watchertypes"
	"github.com/dimasma0305/gzcli/internal/log"
)

var challengeFileRegex = regexp.MustCompile(`^challenge\.(yaml|yml)$`)

// EventWatcher manages file watching for a single event
type EventWatcher struct {
	eventName string
	eventPath string
	api       *gzapi.GZAPI

	watcher            *fsnotify.Watcher
	config             watchertypes.WatcherConfig
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

	// Challenge mapping cache (folder path -> GZCTF challenge ID)
	challengeMappings   map[string]int // folderPath -> challengeID
	challengeMappingsMu sync.RWMutex

	// Additional state
	debounceTimers map[string]*time.Timer
}

// NewEventWatcher creates a new event-specific watcher
func NewEventWatcher(eventName string, api *gzapi.GZAPI, config watchertypes.WatcherConfig, db *database.DB, parentCtx context.Context) (*EventWatcher, error) {
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
		challengeMappings:  make(map[string]int),
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

		// Skip if no update needed
		if updateType == watchertypes.UpdateNone {
			log.InfoH3("[%s] No update needed for %s", ew.eventName, challengeName)
			return
		}

		log.InfoH3("[%s] Sync needed for %s (type: %v)", ew.eventName, challengeName, updateType)
		log.InfoH3("[%s] Challenge path: %s", ew.eventName, challengeCwd)

		// Update challenge state in database
		if ew.scriptMgr != nil {
			activeScripts := ew.scriptMgr.GetActiveIntervalScripts()
			ew.UpdateChallengeState(challengeName, "syncing", "", activeScripts)
		}

		// Check for pending updates and upgrade update type if needed
		if pendingFilePath, hasPending := ew.getPendingUpdate(challengeName); hasPending {
			log.InfoH3("[%s] Found pending update for %s, will also process: %s", ew.eventName, challengeName, pendingFilePath)
			// Re-determine update type for pending file
			pendingUpdateType := filesystem.DetermineUpdateType(pendingFilePath, challengeCwd)
			if pendingUpdateType > updateType {
				updateType = pendingUpdateType
				log.InfoH3("[%s] Upgraded update type to: %v", ew.eventName, updateType)
			}
		}

		// Perform the actual sync
		if err := ew.syncSingleChallenge(challengeName, challengeCwd); err != nil {
			log.Error("[%s] Failed to sync challenge %s: %v", ew.eventName, challengeName, err)
			if ew.scriptMgr != nil {
				activeScripts := ew.scriptMgr.GetActiveIntervalScripts()
				ew.UpdateChallengeState(challengeName, "error", err.Error(), activeScripts)
			}
			return
		}

		// Log completion
		log.Info("[%s] ✓ Sync completed for challenge: %s", ew.eventName, challengeName)
		if ew.scriptMgr != nil {
			activeScripts := ew.scriptMgr.GetActiveIntervalScripts()
			ew.UpdateChallengeState(challengeName, "watching", "", activeScripts)
		}
	}()
}

func (ew *EventWatcher) HandleFileRemoval(filePath string) {
	log.InfoH2("[%s] Processing file removal: %s", ew.eventName, filePath)

	// Check if this is a challenge directory or challenge file removal
	watchedChallenges := ew.challengeMgr.GetChallenges()
	shouldRemove, challengeName, challengeDir := filesystem.CheckFileRemoval(filePath, watchedChallenges)

	if !shouldRemove {
		log.DebugH3("[%s] File removal doesn't affect any watched challenges: %s", ew.eventName, filePath)
		return
	}

	// If we don't have a challenge name, try to find it by path
	if challengeName == "" {
		challengeName = filesystem.FindChallengeByPath(challengeDir, watchedChallenges)
	}

	// Verify the challenge directory was actually removed
	if challengeDir != "" && filesystem.IsChallengeDirectoryRemoved(challengeDir) {
		if challengeName != "" {
			log.InfoH3("[%s] Challenge directory removed: %s (%s)", ew.eventName, challengeName, challengeDir)
			ew.removeChallenge(challengeName)
		} else {
			log.InfoH3("[%s] Challenge directory removed (unknown name): %s", ew.eventName, challengeDir)
		}

		// Trigger rediscovery to find any new or moved challenges
		ew.triggerRediscovery()
	}
}

func (ew *EventWatcher) HandleChallengeRemovalByDir(removedDir string) {
	log.InfoH2("[%s] Processing challenge removal by directory: %s", ew.eventName, removedDir)

	watchedChallenges := ew.challengeMgr.GetChallenges()
	challengeName := filesystem.FindChallengeByPath(removedDir, watchedChallenges)

	if challengeName != "" {
		log.InfoH3("[%s] Removing challenge: %s", ew.eventName, challengeName)
		ew.removeChallenge(challengeName)

		// Trigger rediscovery to find any new or moved challenges
		ew.triggerRediscovery()
	}
}

// removeChallenge removes a challenge from the watcher
func (ew *EventWatcher) removeChallenge(challengeName string) {
	// Stop any running scripts for this challenge
	if ew.scriptMgr != nil {
		ew.scriptMgr.StopAllScriptsForChallenge(challengeName)
	}

	// Remove from challenge manager
	if err := ew.challengeMgr.RemoveChallenge(challengeName); err != nil {
		log.Error("[%s] Failed to remove challenge %s: %v", ew.eventName, challengeName, err)
	} else {
		log.Info("[%s] Successfully removed challenge: %s", ew.eventName, challengeName)
	}

	// Clean up mutexes and state
	ew.challengeMutexesMu.Lock()
	delete(ew.challengeMutexes, challengeName)
	ew.challengeMutexesMu.Unlock()

	ew.updatingMu.Lock()
	delete(ew.updatingChallenges, challengeName)
	ew.updatingMu.Unlock()

	ew.pendingUpdatesMu.Lock()
	delete(ew.pendingUpdates, challengeName)
	ew.pendingUpdatesMu.Unlock()

	// Update database
	if ew.db != nil {
		ew.db.UpdateChallengeState(challengeName, "removed", "", nil)
	}
}

// triggerRediscovery triggers a background rediscovery of challenges
func (ew *EventWatcher) triggerRediscovery() {
	log.InfoH3("[%s] Triggering automatic challenge rediscovery...", ew.eventName)

	go func() {
		if err := ew.discoverChallenges(); err != nil {
			log.Error("[%s] Failed to rediscover challenges: %v", ew.eventName, err)
		} else {
			log.Info("[%s] Challenge rediscovery completed", ew.eventName)
		}
	}()
}

// syncSingleChallenge performs a sync operation for a single challenge
func (ew *EventWatcher) syncSingleChallenge(challengeName, challengePath string) error {
	log.InfoH2("[%s] 🔄 Syncing challenge to GZCTF: %s", ew.eventName, challengeName)

	// Find and load the challenge.yaml file
	challengeYamlPath := filepath.Join(challengePath, "challenge.yaml")
	if _, err := os.Stat(challengeYamlPath); os.IsNotExist(err) {
		// Try challenge.yml
		challengeYamlPath = filepath.Join(challengePath, "challenge.yml")
		if _, err := os.Stat(challengeYamlPath); os.IsNotExist(err) {
			return fmt.Errorf("challenge YAML file not found in %s", challengePath)
		}
	}

	// Load the challenge configuration
	var challengeConf config.ChallengeYaml
	if err := fileutil.ParseYamlFromFile(challengeYamlPath, &challengeConf); err != nil {
		return fmt.Errorf("failed to parse challenge YAML: %w", err)
	}

	// Set the challenge directory
	challengeConf.Cwd = challengePath

	// Determine category from path
	// Path format: events/{event}/{category}/{challenge}/
	relPath, err := filepath.Rel(ew.eventPath, challengePath)
	if err == nil && relPath != "." {
		// Split by path separator
		parts := splitPath(relPath)
		if len(parts) > 0 {
			challengeConf.Category = parts[0]
		}
	}
	if challengeConf.Category == "" {
		// Fallback: extract category from parent directory name
		categoryDir := filepath.Dir(challengePath)
		challengeConf.Category = filepath.Base(categoryDir)
	}

	// Get configuration for this event
	conf, err := config.GetConfigWithEvent(ew.api, ew.eventName,
		ew.noOpGetCache,
		ew.noOpSetCache,
		ew.noOpDeleteCache,
		nil)
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	// Get existing challenges from API
	conf.Event.CS = ew.api
	challenges, err := conf.Event.GetChallenges()
	if err != nil {
		return fmt.Errorf("failed to get challenges from API: %w", err)
	}

	// Sync the challenge using the challenge package
	if err := ew.syncChallengeInternal(conf, challengeConf, challenges); err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	log.Info("[%s] ✅ Successfully synced challenge: %s", ew.eventName, challengeName)
	return nil
}

// syncChallengeInternal performs the actual sync operation
func (ew *EventWatcher) syncChallengeInternal(conf *config.Config, challengeConf config.ChallengeYaml, challenges []gzapi.Challenge) error {
	// Build folder path relative to event (e.g., "Crypto/my-challenge")
	relPath, err := filepath.Rel(ew.eventPath, challengeConf.Cwd)
	if err != nil {
		relPath = challengeConf.Category + "/" + filepath.Base(challengeConf.Cwd)
	}
	folderPath := relPath

	// Step 1: Check if we have a mapping for this folder
	if challengeID, exists := ew.getChallengeID(folderPath); exists {
		log.InfoH3("[%s] Found existing challenge mapping: %s → ID %d", ew.eventName, folderPath, challengeID)

		// Try to fetch the challenge by ID
		existingChallenge, err := ew.fetchChallengeByID(challengeID)
		if err != nil {
			// Challenge might have been deleted in GZCTF - remove mapping and continue
			log.InfoH3("[%s] Challenge ID %d not found in GZCTF (may have been deleted), removing mapping", ew.eventName, challengeID)
			ew.deleteChallengeID(folderPath)
		} else {
			// Found existing challenge - update it with new name
			log.InfoH3("[%s] Updating existing challenge ID %d: %s → %s", ew.eventName, challengeID, existingChallenge.Title, challengeConf.Name)

			// Perform the sync with the existing challenge
			if err := ew.syncToExistingChallenge(conf, challengeConf, existingChallenge); err != nil {
				return fmt.Errorf("failed to update existing challenge: %w", err)
			}

			// Update mapping with new title
			ew.setChallengeID(folderPath, challengeID, challengeConf.Name)
			return nil
		}
	}

	// Step 2: No mapping found - use normal sync flow (create or find by name)
	log.InfoH3("[%s] No mapping found for %s, using normal sync flow", ew.eventName, folderPath)

	// Convert config.ChallengeYaml to challengepkg.ChallengeYaml
	challConfig := &challengepkg.Config{
		Url:   conf.Url,
		Creds: conf.Creds,
		Event: conf.Event,
	}

	challYaml := challengepkg.ChallengeYaml{
		Name:        challengeConf.Name,
		Author:      challengeConf.Author,
		Description: challengeConf.Description,
		Flags:       challengeConf.Flags,
		Value:       challengeConf.Value,
		Provide:     challengeConf.Provide,
		Visible:     challengeConf.Visible,
		Type:        challengeConf.Type,
		Hints:       challengeConf.Hints,
		Container: challengepkg.Container{
			FlagTemplate:         challengeConf.Container.FlagTemplate,
			ContainerImage:       challengeConf.Container.ContainerImage,
			MemoryLimit:          challengeConf.Container.MemoryLimit,
			CpuCount:             challengeConf.Container.CpuCount,
			StorageLimit:         challengeConf.Container.StorageLimit,
			ContainerExposePort:  challengeConf.Container.ContainerExposePort,
			EnableTrafficCapture: challengeConf.Container.EnableTrafficCapture,
		},
		Scripts:   convertScriptsToChallengePkg(challengeConf.Scripts),
		Dashboard: convertDashboardToChallengePkg(challengeConf.Dashboard),
		Category:  challengeConf.Category,
		Cwd:       challengeConf.Cwd,
	}

	// Call the challenge sync function
	if err := challengepkg.SyncChallenge(challConfig, challYaml, challenges, ew.api, ew.noOpGetCache, ew.noOpSetCache); err != nil {
		return err
	}

	// Step 3: After successful sync, get the challenge ID and store mapping
	syncedChallenge, err := conf.Event.GetChallenge(challengeConf.Name)
	if err != nil {
		log.Error("[%s] Failed to fetch synced challenge %s: %v", ew.eventName, challengeConf.Name, err)
	} else {
		// Store the mapping for future syncs
		ew.setChallengeID(folderPath, syncedChallenge.Id, challengeConf.Name)
		log.InfoH3("[%s] Created new challenge mapping: %s → ID %d", ew.eventName, folderPath, syncedChallenge.Id)
	}

	return nil
}

// fetchChallengeByID fetches a challenge from GZCTF by its ID
func (ew *EventWatcher) fetchChallengeByID(challengeID int) (*gzapi.Challenge, error) {
	// Get all challenges and find the one with matching ID
	// Note: GZCTF API might not have a direct "get by ID" endpoint, so we fetch all and filter
	conf, err := config.GetConfigWithEvent(ew.api, ew.eventName,
		ew.noOpGetCache,
		ew.noOpSetCache,
		ew.noOpDeleteCache,
		nil)
	if err != nil {
		return nil, err
	}

	conf.Event.CS = ew.api
	challenges, err := conf.Event.GetChallenges()
	if err != nil {
		return nil, err
	}

	for _, ch := range challenges {
		if ch.Id == challengeID {
			return &ch, nil
		}
	}

	return nil, fmt.Errorf("challenge with ID %d not found", challengeID)
}

// syncToExistingChallenge syncs changes to an existing challenge (handles name changes)
func (ew *EventWatcher) syncToExistingChallenge(conf *config.Config, challengeConf config.ChallengeYaml, existingChallenge *gzapi.Challenge) error {
	// Create a modified challenge config with updated data
	challConfig := &challengepkg.Config{
		Url:   conf.Url,
		Creds: conf.Creds,
		Event: conf.Event,
	}

	challYaml := challengepkg.ChallengeYaml{
		Name:        challengeConf.Name,
		Author:      challengeConf.Author,
		Description: challengeConf.Description,
		Flags:       challengeConf.Flags,
		Value:       challengeConf.Value,
		Provide:     challengeConf.Provide,
		Visible:     challengeConf.Visible,
		Type:        challengeConf.Type,
		Hints:       challengeConf.Hints,
		Container: challengepkg.Container{
			FlagTemplate:         challengeConf.Container.FlagTemplate,
			ContainerImage:       challengeConf.Container.ContainerImage,
			MemoryLimit:          challengeConf.Container.MemoryLimit,
			CpuCount:             challengeConf.Container.CpuCount,
			StorageLimit:         challengeConf.Container.StorageLimit,
			ContainerExposePort:  challengeConf.Container.ContainerExposePort,
			EnableTrafficCapture: challengeConf.Container.EnableTrafficCapture,
		},
		Scripts:   convertScriptsToChallengePkg(challengeConf.Scripts),
		Dashboard: convertDashboardToChallengePkg(challengeConf.Dashboard),
		Category:  challengeConf.Category,
		Cwd:       challengeConf.Cwd,
	}

	// Set the existing challenge in the challenges list to ensure update path is taken
	existingChallenge.CS = ew.api
	challengesList := []gzapi.Challenge{*existingChallenge}

	// Sync using the standard flow (will update existing)
	return challengepkg.SyncChallenge(challConfig, challYaml, challengesList, ew.api, ew.noOpGetCache, ew.noOpSetCache)
}

// Helper methods for update state management
func (ew *EventWatcher) isUpdating(challengeName string) bool {
	ew.updatingMu.RLock()
	defer ew.updatingMu.RUnlock()
	return ew.updatingChallenges[challengeName]
}

// Helper functions for cache operations (no-op for watcher)
func (ew *EventWatcher) noOpGetCache(key string, target interface{}) error {
	return fmt.Errorf("cache miss - watcher doesn't use cache")
}

func (ew *EventWatcher) noOpSetCache(key string, value interface{}) error {
	return nil // Ignore cache writes in watcher
}

func (ew *EventWatcher) noOpDeleteCache(key string) {
	// Ignore cache deletes in watcher
}

// convertScriptsToChallengePkg converts config scripts to challenge package format
func convertScriptsToChallengePkg(scripts map[string]config.ScriptValue) map[string]challengepkg.ScriptValue {
	result := make(map[string]challengepkg.ScriptValue)
	for k, v := range scripts {
		result[k] = challengepkg.ScriptValue{
			Simple: v.Simple,
			Complex: func() *challengepkg.ScriptConfig {
				if v.Complex != nil {
					return &challengepkg.ScriptConfig{
						Execute:  v.Complex.Execute,
						Interval: v.Complex.Interval,
					}
				}
				return nil
			}(),
		}
	}
	return result
}

// convertDashboardToChallengePkg converts config dashboard to challenge package format
func convertDashboardToChallengePkg(dashboard *config.Dashboard) *challengepkg.Dashboard {
	if dashboard == nil {
		return nil
	}
	return &challengepkg.Dashboard{
		Type:   dashboard.Type,
		Config: dashboard.Config,
	}
}

// splitPath splits a path into its components
func splitPath(path string) []string {
	var parts []string
	for path != "" && path != "." {
		dir, file := filepath.Split(path)
		if file != "" {
			parts = append([]string{file}, parts...)
		}
		path = filepath.Clean(dir)
		if path == "." || path == string(filepath.Separator) {
			break
		}
	}
	return parts
}

// getChallengeID retrieves a challenge ID from cache or database
func (ew *EventWatcher) getChallengeID(folderPath string) (int, bool) {
	// Check in-memory cache first
	ew.challengeMappingsMu.RLock()
	if id, exists := ew.challengeMappings[folderPath]; exists {
		ew.challengeMappingsMu.RUnlock()
		log.DebugH3("[%s] Cache hit for challenge mapping: %s → ID %d", ew.eventName, folderPath, id)
		return id, true
	}
	ew.challengeMappingsMu.RUnlock()

	// Cache miss - check database
	if ew.db != nil {
		mapping, err := ew.db.GetChallengeMapping(ew.eventName, folderPath)
		if err != nil {
			log.DebugH3("[%s] Database query error for mapping %s: %v", ew.eventName, folderPath, err)
			return 0, false
		}
		if mapping != nil {
			// Found in database - update cache
			ew.challengeMappingsMu.Lock()
			ew.challengeMappings[folderPath] = mapping.ChallengeID
			ew.challengeMappingsMu.Unlock()
			log.DebugH3("[%s] Database hit for challenge mapping: %s → ID %d", ew.eventName, folderPath, mapping.ChallengeID)
			return mapping.ChallengeID, true
		}
	}

	log.DebugH3("[%s] No mapping found for: %s", ew.eventName, folderPath)
	return 0, false
}

// setChallengeID stores a challenge ID in cache and database
func (ew *EventWatcher) setChallengeID(folderPath string, challengeID int, challengeTitle string) {
	// Update in-memory cache
	ew.challengeMappingsMu.Lock()
	ew.challengeMappings[folderPath] = challengeID
	ew.challengeMappingsMu.Unlock()

	// Store in database for persistence
	if ew.db != nil {
		if err := ew.db.SetChallengeMapping(ew.eventName, folderPath, challengeID, challengeTitle); err != nil {
			log.Error("[%s] Failed to store challenge mapping in database: %v", ew.eventName, err)
		}
	}

	log.DebugH3("[%s] Stored challenge mapping: %s → ID %d (%s)", ew.eventName, folderPath, challengeID, challengeTitle)
}

// deleteChallengeID removes a challenge ID mapping
func (ew *EventWatcher) deleteChallengeID(folderPath string) {
	// Remove from cache
	ew.challengeMappingsMu.Lock()
	delete(ew.challengeMappings, folderPath)
	ew.challengeMappingsMu.Unlock()

	// Remove from database
	if ew.db != nil {
		if err := ew.db.DeleteChallengeMapping(ew.eventName, folderPath); err != nil {
			log.Error("[%s] Failed to delete challenge mapping from database: %v", ew.eventName, err)
		}
	}

	log.DebugH3("[%s] Deleted challenge mapping: %s", ew.eventName, folderPath)
}

// Helper methods for update state management
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
