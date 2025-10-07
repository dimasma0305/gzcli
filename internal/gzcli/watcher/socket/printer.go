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

	var lastLogID int64 = 0

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

// displayLogEntry formats and displays a single log entry
func displayLogEntry(logMap map[string]interface{}) {
	timestamp := ""
	if t, ok := logMap["timestamp"].(string); ok {
		if parsed, err := time.Parse("2006-01-02T15:04:05Z", t); err == nil {
			timestamp = parsed.Format("15:04:05")
		} else {
			timestamp = t
		}
	}

	level := ""
	if l, ok := logMap["level"].(string); ok {
		level = l
	}

	component := ""
	if comp, ok := logMap["component"].(string); ok {
		component = comp
	}

	challenge := ""
	if ch, ok := logMap["challenge"].(string); ok && ch != "" {
		challenge = fmt.Sprintf("[%s]", ch)
	}

	script := ""
	if sc, ok := logMap["script"].(string); ok && sc != "" {
		script = fmt.Sprintf("/%s", sc)
	}

	message := ""
	if m, ok := logMap["message"].(string); ok {
		message = m
	}

	levelIcon := "ℹ️"
	switch level {
	case "ERROR":
		levelIcon = "❌"
	case "WARN":
		levelIcon = "⚠️"
	case "INFO":
		levelIcon = "ℹ️"
	case "DEBUG":
		levelIcon = "🔍"
	}

	fmt.Printf("[%s] %s %s %s%s %s\n", timestamp, levelIcon, component, challenge, script, message)
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

	if data, ok := response.Data["status"].(string); ok && data == "running" {
		fmt.Println("🟢 Status: RUNNING")
	} else {
		fmt.Println("🔴 Status: UNKNOWN")
	}

	if challenges, ok := response.Data["watched_challenges"].(float64); ok {
		fmt.Printf("📁 Watched Challenges: %.0f\n", challenges)
	}

	if dbEnabled, ok := response.Data["database_enabled"].(bool); ok {
		if dbEnabled {
			fmt.Println("🗄️  Database: ENABLED")
		} else {
			fmt.Println("🗄️  Database: DISABLED")
		}
	}

	if socketEnabled, ok := response.Data["socket_enabled"].(bool); ok {
		if socketEnabled {
			fmt.Println("🔌 Socket Server: ENABLED")
		} else {
			fmt.Println("🔌 Socket Server: DISABLED")
		}
	}

	if activeScripts, ok := response.Data["active_scripts"].(map[string]interface{}); ok && len(activeScripts) > 0 {
		fmt.Println("\n🔄 Active Interval Scripts:")
		for challengeName, scriptsInterface := range activeScripts {
			if scripts, ok := scriptsInterface.([]interface{}); ok && len(scripts) > 0 {
				fmt.Printf("  📦 %s:\n", challengeName)
				for _, scriptInterface := range scripts {
					if script, ok := scriptInterface.(string); ok {
						fmt.Printf("    - %s\n", script)
					}
				}
			}
		}
	}

	fmt.Println("\n🛠️  Available Commands:")
	fmt.Println("   gzcli watcher-client status")
	fmt.Println("   gzcli watcher-client list")
	fmt.Println("   gzcli watcher-client logs [--watcher-limit N]")
	fmt.Println("   gzcli watcher-client live-logs [--watcher-limit N] [--watcher-interval 2s]")
	fmt.Println("   gzcli watcher-client metrics")
	fmt.Println("   gzcli watcher-client executions [--watcher-challenge NAME]")
	fmt.Println("   gzcli watcher-client stop-script --watcher-challenge NAME --watcher-script SCRIPT")
	fmt.Println("   gzcli watcher-client restart --watcher-challenge NAME")

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

	if data, ok := response.Data["logs"].([]interface{}); ok {
		if len(data) == 0 {
			fmt.Println("No logs available.")
			return nil
		}

		for _, logInterface := range data {
			if logMap, ok := logInterface.(map[string]interface{}); ok {
				timestamp := ""
				if t, ok := logMap["timestamp"].(string); ok {
					if parsed, err := time.Parse("2006-01-02T15:04:05Z", t); err == nil {
						timestamp = parsed.Format("15:04:05")
					} else {
						timestamp = t
					}
				}

				level := ""
				if l, ok := logMap["level"].(string); ok {
					level = l
				}

				component := ""
				if cmp, ok := logMap["component"].(string); ok {
					component = cmp
				}

				challenge := ""
				if ch, ok := logMap["challenge"].(string); ok && ch != "" {
					challenge = fmt.Sprintf("[%s]", ch)
				}

				message := ""
				if m, ok := logMap["message"].(string); ok {
					message = m
				}

				levelIcon := "ℹ️"
				switch level {
				case "ERROR":
					levelIcon = "❌"
				case "WARN":
					levelIcon = "⚠️"
				case "INFO":
					levelIcon = "ℹ️"
				case "DEBUG":
					levelIcon = "🔍"
				}

				fmt.Printf("[%s] %s %s %s %s\n", timestamp, levelIcon, component, challenge, message)
			}
		}
	}

	return nil
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

	if data, ok := response.Data["metrics"].(map[string]interface{}); ok {
		if len(data) == 0 {
			fmt.Println("No metrics available.")
			return nil
		}

		for challengeName, challengeInterface := range data {
			if challengeMetrics, ok := challengeInterface.(map[string]interface{}); ok {
				fmt.Printf("\n📦 Challenge: %s\n", challengeName)
				fmt.Println("   Scripts:")

				for scriptName, scriptInterface := range challengeMetrics {
					if scriptMetrics, ok := scriptInterface.(map[string]interface{}); ok {
						execCount := float64(0)
						if ec, ok := scriptMetrics["ExecutionCount"].(float64); ok {
							execCount = ec
						}

						lastExecution := ""
						if le, ok := scriptMetrics["LastExecution"].(string); ok {
							if parsed, err := time.Parse("2006-01-02T15:04:05Z", le); err == nil {
								lastExecution = parsed.Format("2006-01-02 15:04:05")
							} else {
								lastExecution = le
							}
						}

						lastDuration := ""
						if ld, ok := scriptMetrics["LastDuration"].(float64); ok {
							switch {
							case ld >= 1000000000: // >= 1 second
								lastDuration = fmt.Sprintf("%.1fs", ld/1000000000)
							case ld >= 1000000: // >= 1 millisecond
								lastDuration = fmt.Sprintf("%.0fms", ld/1000000)
							case ld > 0:
								lastDuration = fmt.Sprintf("%.0fμs", ld/1000)
							}
						}

						// Check if this is an interval script
						isInterval := false
						if ii, ok := scriptMetrics["is_interval"].(bool); ok {
							isInterval = ii
						}

						interval := ""
						if isInterval {
							if iv, ok := scriptMetrics["interval"].(float64); ok && iv > 0 {
								intervalDuration := time.Duration(iv)
								switch {
								case intervalDuration >= time.Hour:
									interval = fmt.Sprintf(" [interval: %.0fh]", intervalDuration.Hours())
								case intervalDuration >= time.Minute:
									interval = fmt.Sprintf(" [interval: %.0fm]", intervalDuration.Minutes())
								default:
									interval = fmt.Sprintf(" [interval: %.0fs]", intervalDuration.Seconds())
								}
							}
						}

						// Create script type indicator
						scriptType := ""
						if isInterval {
							scriptType = "🔄 "
						} else {
							scriptType = "▶️ "
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
				}
			}
		}
	}

	return nil
}
