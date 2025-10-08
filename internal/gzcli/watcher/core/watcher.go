//nolint:revive // Handler methods follow interface patterns with some unused parameters
package core

import (
	"context"
	"fmt"
	"sync"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/database"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/socket"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/types"
	"github.com/dimasma0305/gzcli/internal/log"
)

// Watcher manages file watching and challenge synchronization across multiple events
type Watcher struct {
	api    *gzapi.GZAPI
	config types.WatcherConfig
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Shared components
	db           *database.DB
	socketServer *socket.Server

	// Event-specific watchers
	eventWatchers   map[string]*EventWatcher // eventName -> EventWatcher
	eventWatchersMu sync.RWMutex
}

// New creates a new file watcher instance
func New(api *gzapi.GZAPI) (*Watcher, error) {
	// Validate input
	if api == nil {
		return nil, fmt.Errorf("API client cannot be nil")
	}

	ctx, cancel := context.WithCancel(context.Background())

	w := &Watcher{
		api:           api,
		ctx:           ctx,
		cancel:        cancel,
		eventWatchers: make(map[string]*EventWatcher),
	}

	return w, nil
}

// GetEventWatcher returns the EventWatcher for a specific event
func (w *Watcher) GetEventWatcher(eventName string) (*EventWatcher, bool) {
	w.eventWatchersMu.RLock()
	defer w.eventWatchersMu.RUnlock()
	ew, exists := w.eventWatchers[eventName]
	return ew, exists
}

// GetAllEventWatchers returns all event watchers
func (w *Watcher) GetAllEventWatchers() map[string]*EventWatcher {
	w.eventWatchersMu.RLock()
	defer w.eventWatchersMu.RUnlock()
	watchers := make(map[string]*EventWatcher, len(w.eventWatchers))
	for k, v := range w.eventWatchers {
		watchers[k] = v
	}
	return watchers
}

// AddEventWatcher adds an event watcher
func (w *Watcher) AddEventWatcher(eventName string, ew *EventWatcher) {
	w.eventWatchersMu.Lock()
	defer w.eventWatchersMu.Unlock()
	w.eventWatchers[eventName] = ew
}

// RemoveEventWatcher removes an event watcher
func (w *Watcher) RemoveEventWatcher(eventName string) {
	w.eventWatchersMu.Lock()
	defer w.eventWatchersMu.Unlock()
	delete(w.eventWatchers, eventName)
}

// Implement socket Handler interface
func (w *Watcher) HandleStatusCommand(cmd types.WatcherCommand) types.WatcherResponse {
	// Get event filter from command if specified
	filterEvent := cmd.Event // Prioritize Event field
	if filterEvent == "" && cmd.Data != nil {
		if ev, ok := cmd.Data["event"].(string); ok {
			filterEvent = ev
		}
	}

	eventWatchers := w.GetAllEventWatchers()
	totalChallenges := 0
	allActiveScripts := make(map[string]map[string][]string) // event -> challenge -> []scripts
	events := []string{}

	for eventName, ew := range eventWatchers {
		// Apply event filter if specified
		if filterEvent != "" && eventName != filterEvent {
			continue
		}

		events = append(events, eventName)
		challenges := ew.GetWatchedChallenges()
		totalChallenges += len(challenges)

		scriptMgr := ew.GetScriptManager()
		if scriptMgr != nil {
			allActiveScripts[eventName] = scriptMgr.GetActiveIntervalScripts()
		}
	}

	status := map[string]interface{}{
		"status":             "running",
		"events":             events,
		"watched_challenges": totalChallenges,
		"active_scripts":     allActiveScripts,
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
	// Get event filter from command if specified
	filterEvent := cmd.Event // Prioritize Event field
	if filterEvent == "" && cmd.Data != nil {
		if ev, ok := cmd.Data["event"].(string); ok {
			filterEvent = ev
		}
	}

	eventWatchers := w.GetAllEventWatchers()
	challengeList := make([]map[string]interface{}, 0)

	for eventName, ew := range eventWatchers {
		// Apply event filter if specified
		if filterEvent != "" && eventName != filterEvent {
			continue
		}

		challenges := ew.GetWatchedChallenges()
		for _, challengeName := range challenges {
			challengeInfo := map[string]interface{}{
				"event":    eventName,
				"name":     challengeName,
				"watching": true,
			}
			challengeList = append(challengeList, challengeInfo)
		}
	}

	return types.WatcherResponse{
		Success: true,
		Message: fmt.Sprintf("Found %d challenges", len(challengeList)),
		Data:    map[string]interface{}{"challenges": challengeList},
	}
}

func (w *Watcher) HandleGetMetricsCommand(cmd types.WatcherCommand) types.WatcherResponse {
	// Get event filter from command if specified
	filterEvent := cmd.Event // Prioritize Event field
	if filterEvent == "" && cmd.Data != nil {
		if ev, ok := cmd.Data["event"].(string); ok {
			filterEvent = ev
		}
	}

	eventWatchers := w.GetAllEventWatchers()
	allMetrics := make(map[string]interface{}) // event -> metrics

	for eventName, ew := range eventWatchers {
		// Apply event filter if specified
		if filterEvent != "" && eventName != filterEvent {
			continue
		}

		scriptMgr := ew.GetScriptManager()
		if scriptMgr != nil {
			allMetrics[eventName] = scriptMgr.GetMetrics()
		}
	}

	return types.WatcherResponse{
		Success: true,
		Message: "Script metrics retrieved successfully",
		Data:    map[string]interface{}{"metrics": allMetrics},
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
	// Get event from command
	eventName := cmd.Event
	if eventName == "" && cmd.Data != nil {
		if ev, ok := cmd.Data["event"].(string); ok {
			eventName = ev
		}
	}

	if eventName == "" {
		return types.WatcherResponse{
			Success: false,
			Error:   "Missing event parameter",
		}
	}

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

	// Get the event watcher
	ew, exists := w.GetEventWatcher(eventName)
	if !exists {
		return types.WatcherResponse{
			Success: false,
			Error:   fmt.Sprintf("Event '%s' is not being watched", eventName),
		}
	}

	scriptMgr := ew.GetScriptManager()
	if scriptMgr != nil {
		scriptMgr.StopIntervalScript(challengeName, scriptName)
	}

	return types.WatcherResponse{
		Success: true,
		Message: fmt.Sprintf("Stopped script '%s' for challenge '%s' in event '%s'", scriptName, challengeName, eventName),
	}
}

func (w *Watcher) HandleRestartChallengeCommand(cmd types.WatcherCommand) types.WatcherResponse {
	// Get event from command
	eventName := cmd.Event
	if eventName == "" && cmd.Data != nil {
		if ev, ok := cmd.Data["event"].(string); ok {
			eventName = ev
		}
	}

	if eventName == "" {
		return types.WatcherResponse{
			Success: false,
			Error:   "Missing event parameter",
		}
	}

	if cmd.Data == nil {
		return types.WatcherResponse{
			Success: false,
			Error:   "Missing challenge_name parameter",
		}
	}

	challengeName, ok1 := cmd.Data["challenge_name"].(string)

	if !ok1 {
		return types.WatcherResponse{
			Success: false,
			Error:   "Invalid challenge_name parameter",
		}
	}

	// Get the event watcher
	ew, exists := w.GetEventWatcher(eventName)
	if !exists {
		return types.WatcherResponse{
			Success: false,
			Error:   fmt.Sprintf("Event '%s' is not being watched", eventName),
		}
	}

	// Trigger restart in background
	go func() {
		scriptMgr := ew.GetScriptManager()
		if scriptMgr != nil {
			activeScripts := scriptMgr.GetActiveIntervalScripts()
			ew.UpdateChallengeState(challengeName, "restarting", "", activeScripts)
			log.Info("[%s] Challenge restart requested: %s", eventName, challengeName)
			// Actual restart logic would go here
			ew.UpdateChallengeState(challengeName, "watching", "", activeScripts)
		}
	}()

	return types.WatcherResponse{
		Success: true,
		Message: fmt.Sprintf("Challenge '%s' restart initiated in event '%s'", challengeName, eventName),
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

// StopEventWatcher stops a specific event watcher
func (w *Watcher) StopEventWatcher(eventName string) error {
	ew, exists := w.GetEventWatcher(eventName)
	if !exists {
		return fmt.Errorf("event watcher for '%s' not found", eventName)
	}

	if err := ew.Stop(); err != nil {
		return fmt.Errorf("failed to stop event watcher for '%s': %w", eventName, err)
	}

	w.RemoveEventWatcher(eventName)
	log.Info("Event watcher for '%s' stopped and removed", eventName)
	return nil
}

// HandleStopEventCommand handles stopping a specific event watcher
func (w *Watcher) HandleStopEventCommand(cmd types.WatcherCommand) types.WatcherResponse {
	// Get event from command
	eventName := cmd.Event
	if eventName == "" && cmd.Data != nil {
		if ev, ok := cmd.Data["event"].(string); ok {
			eventName = ev
		}
	}

	if eventName == "" {
		return types.WatcherResponse{
			Success: false,
			Error:   "Missing event parameter",
		}
	}

	// Stop the event watcher
	if err := w.StopEventWatcher(eventName); err != nil {
		return types.WatcherResponse{
			Success: false,
			Error:   err.Error(),
		}
	}

	return types.WatcherResponse{
		Success: true,
		Message: fmt.Sprintf("Event watcher for '%s' stopped successfully", eventName),
	}
}

// GetWatchedChallenges returns all watched challenges from all events
func (w *Watcher) GetWatchedChallenges() []string {
	eventWatchers := w.GetAllEventWatchers()
	var allChallenges []string

	for eventName, ew := range eventWatchers {
		challenges := ew.GetWatchedChallenges()
		for _, ch := range challenges {
			allChallenges = append(allChallenges, fmt.Sprintf("[%s] %s", eventName, ch))
		}
	}

	return allChallenges
}
