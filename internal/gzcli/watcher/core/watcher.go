//nolint:revive // Handler methods follow interface patterns with some unused parameters
package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/challenge"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/database"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/filesystem"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/git"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/scripts"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/socket"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/types"
	"github.com/dimasma0305/gzcli/internal/log"
)

// Watcher manages file watching and challenge synchronization
type Watcher struct {
	api                *gzapi.GZAPI
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
	db           *database.DB
	socketServer *socket.Server
	gitMgr       *git.Manager

	// Additional state
	debounceTimers map[string]*time.Timer
}

// New creates a new file watcher instance
func New(api *gzapi.GZAPI) (*Watcher, error) {
	// Validate input
	if api == nil {
		return nil, fmt.Errorf("API client cannot be nil")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	w := &Watcher{
		api:                api,
		watcher:            watcher,
		ctx:                ctx,
		cancel:             cancel,
		debounceTimers:     make(map[string]*time.Timer),
		challengeMutexes:   make(map[string]*sync.Mutex),
		pendingUpdates:     make(map[string]string),
		updatingChallenges: make(map[string]bool),
	}

	// Initialize component managers
	w.challengeMgr = challenge.NewManager(watcher)
	w.scriptMgr = scripts.NewManager(ctx, w)

	return w, nil
}

// getChallengeUpdateMutex gets or creates a mutex for a specific challenge
func (w *Watcher) GetChallengeUpdateMutex(challengeName string) *sync.Mutex {
	w.challengeMutexesMu.RLock()
	if mutex, exists := w.challengeMutexes[challengeName]; exists {
		w.challengeMutexesMu.RUnlock()
		return mutex
	}
	w.challengeMutexesMu.RUnlock()

	// Need to create new mutex
	w.challengeMutexesMu.Lock()
	defer w.challengeMutexesMu.Unlock()

	// Double-check in case another goroutine created it
	if mutex, exists := w.challengeMutexes[challengeName]; exists {
		return mutex
	}

	// Create new mutex
	mutex := &sync.Mutex{}
	w.challengeMutexes[challengeName] = mutex
	return mutex
}

// Implement ScriptLogger interface for scripts package
func (w *Watcher) LogToDatabase(level, component, challenge, script, message, errorMsg string, duration int64) {
	if w.db != nil {
		w.db.LogToDatabase(level, component, challenge, script, message, errorMsg, duration)
	}
}

func (w *Watcher) LogScriptExecution(challengeName, scriptName, scriptType, command, status string, duration int64, output, errorOutput string, exitCode int) {
	if w.db != nil {
		w.db.LogScriptExecution(challengeName, scriptName, scriptType, command, status, duration, output, errorOutput, exitCode)
	}
}

func (w *Watcher) UpdateChallengeState(challengeName, status, errorMessage string, activeScripts map[string][]string) {
	if w.db != nil {
		w.db.UpdateChallengeState(challengeName, status, errorMessage, activeScripts)
	}
}

// Implement socket Handler interface
func (w *Watcher) HandleStatusCommand(cmd types.WatcherCommand) types.WatcherResponse {
	activeScripts := w.scriptMgr.GetActiveIntervalScripts()
	challenges := len(w.challengeMgr.GetChallenges())

	status := map[string]interface{}{
		"status":             "running",
		"watched_challenges": challenges,
		"active_scripts":     activeScripts,
		"database_enabled":   w.config.DatabaseEnabled,
		"socket_enabled":     w.config.SocketEnabled,
	}

	return types.WatcherResponse{
		Success: true,
		Message: "Watcher status retrieved successfully",
		Data:    status,
	}
}

func (w *Watcher) HandleListChallengesCommand(cmd types.WatcherCommand) types.WatcherResponse {
	challenges := w.challengeMgr.GetChallenges()
	challengeList := make([]map[string]interface{}, 0, len(challenges))

	for name, dir := range challenges {
		challengeInfo := map[string]interface{}{
			"name":      name,
			"watching":  true,
			"directory": dir,
		}
		challengeList = append(challengeList, challengeInfo)
	}

	return types.WatcherResponse{
		Success: true,
		Message: fmt.Sprintf("Found %d challenges", len(challengeList)),
		Data:    map[string]interface{}{"challenges": challengeList},
	}
}

func (w *Watcher) HandleGetMetricsCommand(cmd types.WatcherCommand) types.WatcherResponse {
	metrics := w.scriptMgr.GetMetrics()

	return types.WatcherResponse{
		Success: true,
		Message: "Script metrics retrieved successfully",
		Data:    map[string]interface{}{"metrics": metrics},
	}
}

func (w *Watcher) HandleGetLogsCommand(cmd types.WatcherCommand) types.WatcherResponse {
	if !w.config.DatabaseEnabled {
		return types.WatcherResponse{
			Success: false,
			Error:   "Database logging is disabled",
		}
	}

	// Get limit from command data (default to 100)
	limit := 100
	if cmd.Data != nil {
		if l, ok := cmd.Data["limit"].(float64); ok {
			limit = int(l)
		}
	}

	logs, err := w.db.GetRecentLogs(limit)
	if err != nil {
		return types.WatcherResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get logs: %v", err),
		}
	}

	return types.WatcherResponse{
		Success: true,
		Message: fmt.Sprintf("Retrieved %d log entries", len(logs)),
		Data:    map[string]interface{}{"logs": logs},
	}
}

func (w *Watcher) HandleStopScriptCommand(cmd types.WatcherCommand) types.WatcherResponse {
	if cmd.Data == nil {
		return types.WatcherResponse{
			Success: false,
			Error:   "Missing challenge_name and script_name parameters",
		}
	}

	challengeName, ok1 := cmd.Data["challenge_name"].(string)
	scriptName, ok2 := cmd.Data["script_name"].(string)

	if !ok1 || !ok2 {
		return types.WatcherResponse{
			Success: false,
			Error:   "Invalid challenge_name or script_name parameter",
		}
	}

	w.scriptMgr.StopIntervalScript(challengeName, scriptName)

	return types.WatcherResponse{
		Success: true,
		Message: fmt.Sprintf("Stopped script '%s' for challenge '%s'", scriptName, challengeName),
	}
}

func (w *Watcher) HandleRestartChallengeCommand(cmd types.WatcherCommand) types.WatcherResponse {
	if cmd.Data == nil {
		return types.WatcherResponse{
			Success: false,
			Error:   "Missing challenge_name parameter",
		}
	}

	challengeName, ok := cmd.Data["challenge_name"].(string)
	if !ok {
		return types.WatcherResponse{
			Success: false,
			Error:   "Invalid challenge_name parameter",
		}
	}

	// Trigger restart in background
	go func() {
		activeScripts := w.scriptMgr.GetActiveIntervalScripts()
		w.UpdateChallengeState(challengeName, "restarting", "", activeScripts)
		log.Info("Challenge restart requested: %s", challengeName)
		// Actual restart logic would go here
		w.UpdateChallengeState(challengeName, "watching", "", activeScripts)
	}()

	return types.WatcherResponse{
		Success: true,
		Message: fmt.Sprintf("Challenge '%s' restart initiated", challengeName),
	}
}

func (w *Watcher) HandleGetScriptExecutionsCommand(cmd types.WatcherCommand) types.WatcherResponse {
	if !w.config.DatabaseEnabled {
		return types.WatcherResponse{
			Success: false,
			Error:   "Database logging is disabled",
		}
	}

	limit := 100
	challengeName := ""

	if cmd.Data != nil {
		if l, ok := cmd.Data["limit"].(float64); ok {
			limit = int(l)
		}
		if c, ok := cmd.Data["challenge_name"].(string); ok {
			challengeName = c
		}
	}

	executions, err := w.db.GetScriptExecutions(challengeName, limit)
	if err != nil {
		return types.WatcherResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get script executions: %v", err),
		}
	}

	return types.WatcherResponse{
		Success: true,
		Message: fmt.Sprintf("Retrieved %d script executions", len(executions)),
		Data:    map[string]interface{}{"executions": executions},
	}
}

// Implement filesystem.EventHandler interface
func (w *Watcher) HandleFileChange(filePath string) {
	log.InfoH2("Processing file change: %s", filePath)

	// Find which challenge this file belongs to
	challengeName, challengeCwd, err := w.challengeMgr.FindChallengeForFile(filePath)
	if err != nil {
		log.Error("Failed to find challenge for file %s: %v", filePath, err)
		return
	}

	if challengeName == "" {
		log.InfoH3("File %s doesn't belong to any challenge", filePath)
		return
	}

	log.Info("File %s belongs to challenge: %s", filePath, challengeName)

	// Use the challenge-specific mutex to prevent race conditions during update checks
	challengeMutex := w.GetChallengeUpdateMutex(challengeName)
	challengeMutex.Lock()

	// Check if this challenge is already being updated
	if w.isUpdating(challengeName) {
		log.InfoH3("Challenge %s is already being updated, setting as pending", challengeName)
		w.setPendingUpdate(challengeName, filePath)
		challengeMutex.Unlock()
		return
	}

	// Mark as updating before releasing the mutex
	w.setUpdating(challengeName, true)
	challengeMutex.Unlock()

	// Process update (simplified - actual implementation would call update logic)
	go func() {
		defer w.setUpdating(challengeName, false)

		updateType := filesystem.DetermineUpdateType(filePath, challengeCwd)
		log.Info("Update type for %s: %v", challengeName, updateType)

		// Check for pending updates
		if pendingFilePath, hasPending := w.getPendingUpdate(challengeName); hasPending {
			log.InfoH3("Found pending update for %s, would process: %s", challengeName, pendingFilePath)
		}
	}()
}

func (w *Watcher) HandleFileRemoval(filePath string) {
	log.InfoH2("Processing file removal: %s", filePath)
	// Simplified implementation
}

func (w *Watcher) HandleChallengeRemovalByDir(removedDir string) {
	log.InfoH2("Processing challenge removal by directory: %s", removedDir)
	// Simplified implementation
}

// Helper methods for update state management
func (w *Watcher) isUpdating(challengeName string) bool {
	w.updatingMu.RLock()
	defer w.updatingMu.RUnlock()
	return w.updatingChallenges[challengeName]
}

func (w *Watcher) setUpdating(challengeName string, updating bool) {
	w.updatingMu.Lock()
	defer w.updatingMu.Unlock()
	if updating {
		w.updatingChallenges[challengeName] = true
	} else {
		delete(w.updatingChallenges, challengeName)
	}
}

func (w *Watcher) setPendingUpdate(challengeName, filePath string) {
	w.pendingUpdatesMu.Lock()
	defer w.pendingUpdatesMu.Unlock()
	w.pendingUpdates[challengeName] = filePath
}

func (w *Watcher) getPendingUpdate(challengeName string) (string, bool) {
	w.pendingUpdatesMu.Lock()
	defer w.pendingUpdatesMu.Unlock()
	filePath, exists := w.pendingUpdates[challengeName]
	if exists {
		delete(w.pendingUpdates, challengeName)
	}
	return filePath, exists
}
