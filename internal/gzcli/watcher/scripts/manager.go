package scripts

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/watchertypes"
	"github.com/dimasma0305/gzcli/internal/log"
)

// ScriptLogger interface for logging script events
type ScriptLogger interface {
	LogToDatabase(level, component, challenge, script, message, errorMsg string, duration int64)
	LogScriptExecution(challengeName, scriptName, scriptType, command, status string, duration int64, output, errorOutput string, exitCode int)
}

// ChallengeConfig interface for accessing challenge configuration
type ChallengeConfig interface {
	GetName() string
	GetCwd() string
	GetScripts() map[string]ScriptValue
}

// ScriptValue interface for accessing script values
type ScriptValue interface {
	GetCommand() string
	HasInterval() bool
	GetInterval() time.Duration
}

// Manager manages script execution and lifecycle
type Manager struct {
	ctx               context.Context
	intervalScripts   map[string]map[string]context.CancelFunc
	intervalScriptsMu sync.RWMutex
	scriptMetrics     map[string]map[string]*watchertypes.ScriptMetrics
	scriptMetricsMu   sync.RWMutex
	challengeConfigs  map[string]ChallengeConfig
	configsMu         sync.RWMutex
	logger            ScriptLogger
}

// NewManager creates a new script manager
func NewManager(ctx context.Context, logger ScriptLogger) *Manager {
	return &Manager{
		ctx:              ctx,
		intervalScripts:  make(map[string]map[string]context.CancelFunc),
		scriptMetrics:    make(map[string]map[string]*watchertypes.ScriptMetrics),
		challengeConfigs: make(map[string]ChallengeConfig),
		logger:           logger,
	}
}

// RegisterChallenge registers a challenge configuration for script execution
func (m *Manager) RegisterChallenge(challenge ChallengeConfig) {
	m.configsMu.Lock()
	defer m.configsMu.Unlock()
	m.challengeConfigs[challenge.GetName()] = challenge
}

// UnregisterChallenge removes a challenge configuration
func (m *Manager) UnregisterChallenge(challengeName string) {
	m.configsMu.Lock()
	defer m.configsMu.Unlock()
	delete(m.challengeConfigs, challengeName)
}

// RunScriptWithIntervalSupport runs a script with proper interval script lifecycle management
func (m *Manager) RunScriptWithIntervalSupport(challenge ChallengeConfig, scriptName string) error {
	scripts := challenge.GetScripts()
	scriptValue, exists := scripts[scriptName]
	if !exists {
		return nil
	}

	command := scriptValue.GetCommand()
	if command == "" {
		return nil
	}

	// Check if script has an interval configured
	if scriptValue.HasInterval() {
		interval := scriptValue.GetInterval()
		log.InfoH2("Starting interval script '%s' with interval %v", scriptName, interval)
		log.InfoH3("Script command: %s", command)

		// Log script start
		if m.logger != nil {
			m.logger.LogToDatabase("INFO", "script", challenge.GetName(), scriptName,
				fmt.Sprintf("Starting interval script with interval %v", interval), "", 0)
		}

		// Use manager's interval script management
		m.StartIntervalScript(challenge.GetName(), scriptName, challenge, command, interval)
		return nil
	}

	// For non-interval scripts, stop any existing interval script with the same name
	m.StopIntervalScript(challenge.GetName(), scriptName)

	// Initialize metrics for one-time script if needed
	m.scriptMetricsMu.Lock()
	if m.scriptMetrics[challenge.GetName()] == nil {
		m.scriptMetrics[challenge.GetName()] = make(map[string]*watchertypes.ScriptMetrics)
	}
	if m.scriptMetrics[challenge.GetName()][scriptName] == nil {
		m.scriptMetrics[challenge.GetName()][scriptName] = &watchertypes.ScriptMetrics{
			IsInterval: false,
			Interval:   0,
		}
	} else {
		// Update existing metrics to mark as non-interval
		m.scriptMetrics[challenge.GetName()][scriptName].IsInterval = false
		m.scriptMetrics[challenge.GetName()][scriptName].Interval = 0
	}
	m.scriptMetricsMu.Unlock()

	// Log script execution start
	start := time.Now()
	if m.logger != nil {
		m.logger.LogScriptExecution(challenge.GetName(), scriptName, "one-time", command, "started", 0, "", "", 0)
	}

	// Run simple one-time script with timeout protection
	log.InfoH2("Running:\n%s", command)
	ctx, cancel := context.WithTimeout(context.Background(), DefaultScriptTimeout)
	defer cancel()

	err := RunShellWithContext(ctx, command, challenge.GetCwd())
	duration := time.Since(start)

	// Update metrics
	m.scriptMetricsMu.Lock()
	if m.scriptMetrics[challenge.GetName()] != nil && m.scriptMetrics[challenge.GetName()][scriptName] != nil {
		metrics := m.scriptMetrics[challenge.GetName()][scriptName]
		metrics.LastExecution = start
		metrics.ExecutionCount++
		metrics.LastDuration = duration
		metrics.TotalDuration += duration
		if err != nil {
			metrics.LastError = err
		} else {
			metrics.LastError = nil
		}
	}
	m.scriptMetricsMu.Unlock()

	// Log script completion
	if err != nil {
		if m.logger != nil {
			m.logger.LogToDatabase("ERROR", "script", challenge.GetName(), scriptName,
				"One-time script execution failed", err.Error(), duration.Milliseconds())
			m.logger.LogScriptExecution(challenge.GetName(), scriptName, "one-time", command, "failed", duration.Nanoseconds(), "", err.Error(), 1)
		}
	} else {
		if m.logger != nil {
			m.logger.LogToDatabase("INFO", "script", challenge.GetName(), scriptName,
				"One-time script execution completed successfully", "", duration.Milliseconds())
			m.logger.LogScriptExecution(challenge.GetName(), scriptName, "one-time", command, "completed", duration.Nanoseconds(), "", "", 0)
		}
	}

	return err
}

// GetMetrics returns script execution metrics for monitoring
func (m *Manager) GetMetrics() map[string]map[string]*watchertypes.ScriptMetrics {
	m.scriptMetricsMu.RLock()
	m.configsMu.RLock()
	defer m.scriptMetricsMu.RUnlock()
	defer m.configsMu.RUnlock()

	// Create a copy to avoid concurrent map access and enrich with interval information
	result := make(map[string]map[string]*watchertypes.ScriptMetrics)

	for challengeName, challengeMetrics := range m.scriptMetrics {
		result[challengeName] = make(map[string]*watchertypes.ScriptMetrics)

		// Get challenge config for interval information
		challengeConfig, hasConfig := m.challengeConfigs[challengeName]

		for scriptName, metrics := range challengeMetrics {
			// Create a copy of the metrics
			metricsCopy := &watchertypes.ScriptMetrics{
				LastExecution:  metrics.LastExecution,
				ExecutionCount: metrics.ExecutionCount,
				LastError:      metrics.LastError,
				LastDuration:   metrics.LastDuration,
				TotalDuration:  metrics.TotalDuration,
				IsInterval:     false,
				Interval:       0,
			}

			// Check if this script has interval configuration
			if hasConfig {
				scripts := challengeConfig.GetScripts()
				if scriptValue, exists := scripts[scriptName]; exists {
					if scriptValue.HasInterval() {
						metricsCopy.IsInterval = true
						metricsCopy.Interval = scriptValue.GetInterval()
					}
				}
			}

			result[challengeName][scriptName] = metricsCopy
		}
	}
	return result
}

// GetActiveIntervalScripts returns a list of currently running interval scripts
func (m *Manager) GetActiveIntervalScripts() map[string][]string {
	m.intervalScriptsMu.RLock()
	defer m.intervalScriptsMu.RUnlock()

	result := make(map[string][]string)
	for challengeName, scripts := range m.intervalScripts {
		result[challengeName] = make([]string, 0, len(scripts))
		for scriptName := range scripts {
			result[challengeName] = append(result[challengeName], scriptName)
		}
	}
	return result
}

// StopAllScriptsForChallenge stops all interval scripts for a challenge
func (m *Manager) StopAllScriptsForChallenge(challengeName string) {
	m.intervalScriptsMu.Lock()
	defer m.intervalScriptsMu.Unlock()

	if challengeScripts, exists := m.intervalScripts[challengeName]; exists {
		log.InfoH3("Stopping all interval scripts for challenge '%s'", challengeName)
		for scriptName, cancel := range challengeScripts {
			log.InfoH3("  - Stopping interval script '%s'", scriptName)
			cancel()
		}
		delete(m.intervalScripts, challengeName)
	}
}

// StopAllScripts stops all interval scripts
func (m *Manager) StopAllScripts(timeout time.Duration) {
	log.Info("Stopping all interval scripts with timeout %v...", timeout)

	m.intervalScriptsMu.Lock()
	defer m.intervalScriptsMu.Unlock()

	if len(m.intervalScripts) == 0 {
		return
	}

	// Cancel all scripts
	for challengeName := range m.intervalScripts {
		log.InfoH3("Stopping all interval scripts for challenge '%s'", challengeName)
		for scriptName, cancel := range m.intervalScripts[challengeName] {
			log.InfoH3("  - Stopping interval script '%s'", scriptName)
			cancel()
		}
	}

	// Clear all tracking
	m.intervalScripts = make(map[string]map[string]context.CancelFunc)

	// Give scripts time to finish
	if timeout > 0 {
		log.InfoH3("Waiting up to %v for scripts to finish...", timeout)
		time.Sleep(timeout)
	}
}
