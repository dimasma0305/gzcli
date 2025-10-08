package socket

import (
	"fmt"
	"time"
)

// StreamLiveLogs streams database logs in real-time
func (c *Client) StreamLiveLogs(limit int, interval time.Duration) error {
	fmt.Printf("📡 Live Database Logs (refreshing every %v)\n", interval)
	fmt.Println("==========================================")
	fmt.Println("Press Ctrl+C to stop streaming")
	fmt.Println()

	var lastLogID int64

	for {
		// Get recent logs
		response, err := c.GetLogs(limit)
		if err != nil {
			fmt.Printf("❌ Error getting logs: %v\n", err)
			time.Sleep(interval)
			continue
		}

		if !response.Success {
			fmt.Printf("❌ Failed to get logs: %s\n", response.Error)
			time.Sleep(interval)
			continue
		}

		// Process and display new logs
		if data, ok := response.Data["logs"].([]interface{}); ok && len(data) > 0 {
			newLogs := []interface{}{}

			// Filter for new logs only
			for _, logInterface := range data {
				if logMap, ok := logInterface.(map[string]interface{}); ok {
					if idFloat, ok := logMap["id"].(float64); ok {
						logID := int64(idFloat)
						if logID > lastLogID {
							newLogs = append(newLogs, logInterface)
							if logID > lastLogID {
								lastLogID = logID
							}
						}
					}
				}
			}

			// Display new logs (reverse order to show newest first)
			if len(newLogs) > 0 {
				for i := len(newLogs) - 1; i >= 0; i-- {
					logInterface := newLogs[i]
					if logMap, ok := logInterface.(map[string]interface{}); ok {
						displayLogEntry(logMap)
					}
				}
			}
		}

		time.Sleep(interval)
	}
}

// formatTimestamp formats a timestamp interface value
func formatTimestamp(ts interface{}) string {
	t, ok := ts.(string)
	if !ok {
		return ""
	}
	if parsed, err := time.Parse("2006-01-02T15:04:05Z", t); err == nil {
		return parsed.Format("15:04:05")
	}
	return t
}

// displayLogEntry formats and displays a single log entry
func displayLogEntry(logMap map[string]interface{}) {
	timestamp, level, component, challenge, message := extractLogData(logMap)

	script := ""
	if sc, ok := logMap["script"].(string); ok && sc != "" {
		script = fmt.Sprintf("/%s", sc)
	}

	levelIcon := getLevelIcon(level)
	fmt.Printf("[%s] %s %s %s%s %s\n", timestamp, levelIcon, component, challenge, script, message)
}

// printStatusInfo prints basic status information
func printStatusInfo(data map[string]interface{}) {
	if status, ok := data["status"].(string); ok && status == "running" {
		fmt.Println("🟢 Status: RUNNING")
	} else {
		fmt.Println("🔴 Status: UNKNOWN")
	}

	if challenges, ok := data["watched_challenges"].(float64); ok {
		fmt.Printf("📁 Watched Challenges: %.0f\n", challenges)
	}
}

// printFeatureStatus prints database and socket status
func printFeatureStatus(data map[string]interface{}) {
	if dbEnabled, ok := data["database_enabled"].(bool); ok {
		status := "DISABLED"
		if dbEnabled {
			status = "ENABLED"
		}
		fmt.Printf("🗄️  Database: %s\n", status)
	}

	if socketEnabled, ok := data["socket_enabled"].(bool); ok {
		status := "DISABLED"
		if socketEnabled {
			status = "ENABLED"
		}
		fmt.Printf("🔌 Socket Server: %s\n", status)
	}
}

// printActiveScripts prints active interval scripts
func printActiveScripts(data map[string]interface{}) {
	activeScripts, ok := data["active_scripts"].(map[string]interface{})
	if !ok || len(activeScripts) == 0 {
		return
	}

	fmt.Println("\n🔄 Active Interval Scripts:")
	for challengeName, scriptsInterface := range activeScripts {
		scripts, ok := scriptsInterface.([]interface{})
		if !ok || len(scripts) == 0 {
			continue
		}

		fmt.Printf("  📦 %s:\n", challengeName)
		for _, scriptInterface := range scripts {
			if script, ok := scriptInterface.(string); ok {
				fmt.Printf("    - %s\n", script)
			}
		}
	}
}

// printAvailableCommands prints the list of available commands
func printAvailableCommands() {
	fmt.Println("\n🛠️  Available Commands:")
	fmt.Println("   gzcli watcher-client status")
	fmt.Println("   gzcli watcher-client list")
	fmt.Println("   gzcli watcher-client logs [--watcher-limit N]")
	fmt.Println("   gzcli watcher-client live-logs [--watcher-limit N] [--watcher-interval 2s]")
	fmt.Println("   gzcli watcher-client metrics")
	fmt.Println("   gzcli watcher-client executions [--watcher-challenge NAME]")
	fmt.Println("   gzcli watcher-client stop-script --watcher-challenge NAME --watcher-script SCRIPT")
	fmt.Println("   gzcli watcher-client restart --watcher-challenge NAME")
}

// PrintStatus prints a formatted status report
func (c *Client) PrintStatus() error {
	response, err := c.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("status request failed: %s", response.Error)
	}

	fmt.Println("🔍 GZCTF Watcher Status")
	fmt.Println("==========================================")

	printStatusInfo(response.Data)
	printFeatureStatus(response.Data)
	printActiveScripts(response.Data)
	printAvailableCommands()

	return nil
}

// PrintChallenges prints a formatted list of challenges
func (c *Client) PrintChallenges() error {
	response, err := c.ListChallenges()
	if err != nil {
		return fmt.Errorf("failed to list challenges: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("list challenges request failed: %s", response.Error)
	}

	fmt.Println("📁 Watched Challenges")
	fmt.Println("==========================================")

	if data, ok := response.Data["challenges"].([]interface{}); ok {
		if len(data) == 0 {
			fmt.Println("No challenges are currently being watched.")
			return nil
		}

		for i, challengeInterface := range data {
			if challenge, ok := challengeInterface.(map[string]interface{}); ok {
				name := "unknown"
				if n, ok := challenge["name"].(string); ok {
					name = n
				}

				watching := false
				if w, ok := challenge["watching"].(bool); ok {
					watching = w
				}

				directory := ""
				if d, ok := challenge["directory"].(string); ok {
					directory = d
				}

				status := "🔴"
				if watching {
					status = "🟢"
				}

				fmt.Printf("%d. %s %s\n", i+1, status, name)
				if directory != "" {
					fmt.Printf("   📂 %s\n", directory)
				}
			}
		}
	}

	return nil
}

// extractLogData extracts log data from a log map
func extractLogData(logMap map[string]interface{}) (timestamp, level, component, challenge, message string) {
	timestamp = formatTimestamp(logMap["timestamp"])
	level, _ = logMap["level"].(string)
	component, _ = logMap["component"].(string)

	if ch, ok := logMap["challenge"].(string); ok && ch != "" {
		challenge = fmt.Sprintf("[%s]", ch)
	}

	message, _ = logMap["message"].(string)
	return
}

// getLevelIcon returns the appropriate icon for a log level
func getLevelIcon(level string) string {
	switch level {
	case "ERROR":
		return "❌"
	case "WARN":
		return "⚠️"
	case "INFO":
		return "ℹ️"
	case "DEBUG":
		return "🔍"
	default:
		return "ℹ️"
	}
}

// PrintLogs prints formatted recent logs
func (c *Client) PrintLogs(limit int) error {
	response, err := c.GetLogs(limit)
	if err != nil {
		return fmt.Errorf("failed to get logs: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("get logs request failed: %s", response.Error)
	}

	fmt.Printf("📋 Recent Logs (last %d entries)\n", limit)
	fmt.Println("==========================================")

	data, ok := response.Data["logs"].([]interface{})
	if !ok {
		return nil
	}

	if len(data) == 0 {
		fmt.Println("No logs available.")
		return nil
	}

	for _, logInterface := range data {
		if logMap, ok := logInterface.(map[string]interface{}); ok {
			timestamp, level, component, challenge, message := extractLogData(logMap)
			levelIcon := getLevelIcon(level)
			fmt.Printf("[%s] %s %s %s %s\n", timestamp, levelIcon, component, challenge, message)
		}
	}

	return nil
}

// formatDuration formats a duration in nanoseconds to a human-readable string
func formatDuration(durationNs float64) string {
	if durationNs == 0 {
		return ""
	}
	switch {
	case durationNs >= 1000000000: // >= 1 second
		return fmt.Sprintf("%.1fs", durationNs/1000000000)
	case durationNs >= 1000000: // >= 1 millisecond
		return fmt.Sprintf("%.0fms", durationNs/1000000)
	default:
		return fmt.Sprintf("%.0fμs", durationNs/1000)
	}
}

// formatInterval formats an interval duration
func formatInterval(intervalNs float64) string {
	if intervalNs == 0 {
		return ""
	}

	intervalDuration := time.Duration(intervalNs)
	switch {
	case intervalDuration >= time.Hour:
		return fmt.Sprintf(" [interval: %.0fh]", intervalDuration.Hours())
	case intervalDuration >= time.Minute:
		return fmt.Sprintf(" [interval: %.0fm]", intervalDuration.Minutes())
	default:
		return fmt.Sprintf(" [interval: %.0fs]", intervalDuration.Seconds())
	}
}

// formatLastExecution formats a last execution timestamp
func formatLastExecution(lastExecStr string) string {
	if lastExecStr == "" {
		return ""
	}
	if parsed, err := time.Parse("2006-01-02T15:04:05Z", lastExecStr); err == nil {
		return parsed.Format("2006-01-02 15:04:05")
	}
	return lastExecStr
}

// extractScriptMetrics extracts script metrics from a map
func extractScriptMetrics(scriptMetrics map[string]interface{}) (execCount float64, lastExecution, lastDuration string, isInterval bool, interval string) {
	execCount, _ = scriptMetrics["ExecutionCount"].(float64)

	if le, ok := scriptMetrics["LastExecution"].(string); ok {
		lastExecution = formatLastExecution(le)
	}

	if ld, ok := scriptMetrics["LastDuration"].(float64); ok {
		lastDuration = formatDuration(ld)
	}

	isInterval, _ = scriptMetrics["is_interval"].(bool)

	if isInterval {
		if iv, ok := scriptMetrics["interval"].(float64); ok && iv > 0 {
			interval = formatInterval(iv)
		}
	}

	return
}

// printScriptMetric prints metrics for a single script
func printScriptMetric(scriptName string, execCount float64, lastExecution, lastDuration string, isInterval bool, interval string) {
	scriptType := "▶️ "
	if isInterval {
		scriptType = "🔄 "
	}

	fmt.Printf("     %s%s: %.0f executions", scriptType, scriptName, execCount)

	if lastExecution != "" && lastExecution != "0001-01-01 00:00:00" {
		fmt.Printf(", last: %s", lastExecution)
	}

	if lastDuration != "" {
		fmt.Printf(" (%s)", lastDuration)
	}

	if interval != "" {
		fmt.Printf("%s", interval)
	}

	fmt.Println()
}

// PrintMetrics prints formatted script metrics
func (c *Client) PrintMetrics() error {
	response, err := c.GetMetrics()
	if err != nil {
		return fmt.Errorf("failed to get metrics: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("get metrics request failed: %s", response.Error)
	}

	fmt.Println("📊 Script Execution Metrics")
	fmt.Println("==========================================")

	data, ok := response.Data["metrics"].(map[string]interface{})
	if !ok {
		return nil
	}

	if len(data) == 0 {
		fmt.Println("No metrics available.")
		return nil
	}

	for challengeName, challengeInterface := range data {
		challengeMetrics, ok := challengeInterface.(map[string]interface{})
		if !ok {
			continue
		}

		fmt.Printf("\n📦 Challenge: %s\n", challengeName)
		fmt.Println("   Scripts:")

		for scriptName, scriptInterface := range challengeMetrics {
			scriptMetrics, ok := scriptInterface.(map[string]interface{})
			if !ok {
				continue
			}

			execCount, lastExecution, lastDuration, isInterval, interval := extractScriptMetrics(scriptMetrics)
			printScriptMetric(scriptName, execCount, lastExecution, lastDuration, isInterval, interval)
		}
	}

	return nil
}
