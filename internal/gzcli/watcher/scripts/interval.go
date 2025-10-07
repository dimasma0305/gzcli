package scripts

import (
	"context"
	"time"

	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/types"
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
		m.scriptMetrics[challengeName] = make(map[string]*types.ScriptMetrics)
	}
	if m.scriptMetrics[challengeName][scriptName] == nil {
		m.scriptMetrics[challengeName][scriptName] = &types.ScriptMetrics{
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

	// Log initial start to database
	if m.logger != nil {
		m.logger.LogScriptExecution(challengeName, scriptName, "interval", command, "started", 0, "", "", 0)
	}

	for {
		select {
		case <-ctx.Done():
			log.InfoH3("Stopped interval script '%s' for challenge '%s' (context cancelled)", scriptName, challengeName)
			if m.logger != nil {
				m.logger.LogScriptExecution(challengeName, scriptName, "interval", command, "stopped", 0, "", "Context cancelled", 0)
			}
			return
		case <-ticker.C:
			log.InfoH3("Executing interval script '%s' for challenge '%s'", scriptName, challengeName)

			// Log execution start
			start := time.Now()
			if m.logger != nil {
				m.logger.LogScriptExecution(challengeName, scriptName, "interval", command, "executing", 0, "", "", 0)
			}

			// Update metrics
			m.scriptMetricsMu.Lock()
			if m.scriptMetrics[challengeName] != nil && m.scriptMetrics[challengeName][scriptName] != nil {
				m.scriptMetrics[challengeName][scriptName].LastExecution = start
				m.scriptMetrics[challengeName][scriptName].ExecutionCount++
			}
			m.scriptMetricsMu.Unlock()

			// Execute the script with context-aware execution and proper timeout
			var exitCode int = 0
			var success bool = false
			var errorOutput string = ""
			var output string = ""

			err := RunShellForInterval(ctx, command, cwd, DefaultScriptTimeout)
			duration := time.Since(start)

			if err != nil {
				log.Error("Interval script '%s' failed for challenge '%s' after %v: %v", scriptName, challengeName, duration, err)
				success = false
				exitCode = 1
				errorOutput = err.Error()

				// Update metrics with error
				m.scriptMetricsMu.Lock()
				if m.scriptMetrics[challengeName] != nil && m.scriptMetrics[challengeName][scriptName] != nil {
					metrics := m.scriptMetrics[challengeName][scriptName]
					metrics.LastError = err
					metrics.LastDuration = duration
					metrics.TotalDuration += duration
				}
				m.scriptMetricsMu.Unlock()
			} else {
				log.InfoH3("Interval script '%s' completed successfully for challenge '%s' in %v", scriptName, challengeName, duration)
				success = true
				exitCode = 0

				// Update metrics with success
				m.scriptMetricsMu.Lock()
				if m.scriptMetrics[challengeName] != nil && m.scriptMetrics[challengeName][scriptName] != nil {
					metrics := m.scriptMetrics[challengeName][scriptName]
					metrics.LastError = nil
					metrics.LastDuration = duration
					metrics.TotalDuration += duration
				}
				m.scriptMetricsMu.Unlock()
			}

			// Log execution completion to database
			status := "failed"
			if success {
				status = "completed"
			}

			if m.logger != nil {
				m.logger.LogScriptExecution(challengeName, scriptName, "interval", command, status,
					duration.Nanoseconds(), output, errorOutput, exitCode)
			}
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
