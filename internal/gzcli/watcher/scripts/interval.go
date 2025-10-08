package scripts

import (
	"context"
	"time"

	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/watchertypes"
	"github.com/dimasma0305/gzcli/internal/log"
)

// StartIntervalScript starts an interval script for a challenge with proper tracking and validation
func (m *Manager) StartIntervalScript(challengeName, scriptName string, challenge ChallengeConfig, command string, interval time.Duration) {
	// Validate interval before starting
	if !ValidateInterval(interval, scriptName) {
		log.Error("Invalid interval for script '%s' in challenge '%s', skipping", scriptName, challengeName)
		return
	}

	m.intervalScriptsMu.Lock()
	defer m.intervalScriptsMu.Unlock()

	// Initialize map for challenge if it doesn't exist
	if m.intervalScripts[challengeName] == nil {
		m.intervalScripts[challengeName] = make(map[string]context.CancelFunc)
	}

	// Stop existing interval script if running
	if cancel, exists := m.intervalScripts[challengeName][scriptName]; exists {
		log.InfoH3("Stopping existing interval script '%s' for challenge '%s'", scriptName, challengeName)
		cancel()
	}

	// Initialize metrics if needed
	m.scriptMetricsMu.Lock()
	if m.scriptMetrics[challengeName] == nil {
		m.scriptMetrics[challengeName] = make(map[string]*watchertypes.ScriptMetrics)
	}
	if m.scriptMetrics[challengeName][scriptName] == nil {
		m.scriptMetrics[challengeName][scriptName] = &watchertypes.ScriptMetrics{
			IsInterval: true,
			Interval:   interval,
		}
	} else {
		// Update existing metrics with interval info
		m.scriptMetrics[challengeName][scriptName].IsInterval = true
		m.scriptMetrics[challengeName][scriptName].Interval = interval
	}
	m.scriptMetricsMu.Unlock()

	// Create new context for this interval script
	ctx, cancel := context.WithCancel(m.ctx)
	m.intervalScripts[challengeName][scriptName] = cancel

	// Start the interval script in a goroutine
	go m.runIntervalScript(ctx, challengeName, scriptName, command, interval, challenge.GetCwd())
}

// updateScriptMetricsStart updates metrics at the start of execution
func (m *Manager) updateScriptMetricsStart(challengeName, scriptName string, start time.Time) {
	m.scriptMetricsMu.Lock()
	defer m.scriptMetricsMu.Unlock()

	if m.scriptMetrics[challengeName] != nil && m.scriptMetrics[challengeName][scriptName] != nil {
		m.scriptMetrics[challengeName][scriptName].LastExecution = start
		m.scriptMetrics[challengeName][scriptName].ExecutionCount++
	}
}

// updateScriptMetricsEnd updates metrics at the end of execution
func (m *Manager) updateScriptMetricsEnd(challengeName, scriptName string, duration time.Duration, err error) {
	m.scriptMetricsMu.Lock()
	defer m.scriptMetricsMu.Unlock()

	if m.scriptMetrics[challengeName] != nil && m.scriptMetrics[challengeName][scriptName] != nil {
		metrics := m.scriptMetrics[challengeName][scriptName]
		metrics.LastError = err
		metrics.LastDuration = duration
		metrics.TotalDuration += duration
	}
}

// logScriptStart logs the start of a script execution
func (m *Manager) logScriptStart(challengeName, scriptName, command string) {
	if m.logger != nil {
		m.logger.LogScriptExecution(challengeName, scriptName, "interval", command, "started", 0, "", "", 0)
	}
}

// logScriptStop logs when a script is stopped
func (m *Manager) logScriptStop(challengeName, scriptName, command string) {
	if m.logger != nil {
		m.logger.LogScriptExecution(challengeName, scriptName, "interval", command, "stopped", 0, "", "Context cancelled", 0)
	}
}

// logScriptExecution logs a script execution attempt
func (m *Manager) logScriptExecution(challengeName, scriptName, command string) {
	if m.logger != nil {
		m.logger.LogScriptExecution(challengeName, scriptName, "interval", command, "executing", 0, "", "", 0)
	}
}

// logScriptCompletion logs the completion of a script execution
func (m *Manager) logScriptCompletion(challengeName, scriptName, command, output, errorOutput string, duration time.Duration, exitCode int, success bool) {
	status := "failed"
	if success {
		status = "completed"
	}

	if m.logger != nil {
		m.logger.LogScriptExecution(challengeName, scriptName, "interval", command, status,
			duration.Nanoseconds(), output, errorOutput, exitCode)
	}
}

// executeIntervalScriptOnce executes an interval script once and returns the result
func (m *Manager) executeIntervalScriptOnce(ctx context.Context, challengeName, scriptName, command, cwd string) (time.Duration, error) {
	start := time.Now()
	m.logScriptExecution(challengeName, scriptName, command)
	m.updateScriptMetricsStart(challengeName, scriptName, start)

	err := RunShellForInterval(ctx, command, cwd, DefaultScriptTimeout)
	duration := time.Since(start)

	return duration, err
}

// runIntervalScript runs an interval script with proper integration and database logging
func (m *Manager) runIntervalScript(ctx context.Context, challengeName, scriptName, command string, interval time.Duration, cwd string) {
	// Validate interval
	if !ValidateInterval(interval, scriptName) {
		log.Error("Invalid interval for script '%s' in challenge '%s', skipping", scriptName, challengeName)
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.InfoH3("Started interval script '%s' for challenge '%s' with interval %v", scriptName, challengeName, interval)
	m.logScriptStart(challengeName, scriptName, command)

	for {
		select {
		case <-ctx.Done():
			log.InfoH3("Stopped interval script '%s' for challenge '%s' (context cancelled)", scriptName, challengeName)
			m.logScriptStop(challengeName, scriptName, command)
			return
		case <-ticker.C:
			log.InfoH3("Executing interval script '%s' for challenge '%s'", scriptName, challengeName)

			duration, err := m.executeIntervalScriptOnce(ctx, challengeName, scriptName, command, cwd)

			exitCode := 0
			success := true
			errorOutput := ""
			output := ""

			if err != nil {
				log.Error("Interval script '%s' failed for challenge '%s' after %v: %v", scriptName, challengeName, duration, err)
				success = false
				exitCode = 1
				errorOutput = err.Error()
			} else {
				log.InfoH3("Interval script '%s' completed successfully for challenge '%s' in %v", scriptName, challengeName, duration)
			}

			m.updateScriptMetricsEnd(challengeName, scriptName, duration, err)
			m.logScriptCompletion(challengeName, scriptName, command, output, errorOutput, duration, exitCode, success)
		}
	}
}

// StopIntervalScript stops a specific interval script for a challenge
func (m *Manager) StopIntervalScript(challengeName, scriptName string) {
	m.intervalScriptsMu.Lock()
	defer m.intervalScriptsMu.Unlock()

	if challengeScripts, exists := m.intervalScripts[challengeName]; exists {
		if cancel, exists := challengeScripts[scriptName]; exists {
			log.InfoH3("Stopping interval script '%s' for challenge '%s'", scriptName, challengeName)
			cancel()
			delete(challengeScripts, scriptName)

			// Clean up empty challenge map
			if len(challengeScripts) == 0 {
				delete(m.intervalScripts, challengeName)
			}
		}
	}
}
