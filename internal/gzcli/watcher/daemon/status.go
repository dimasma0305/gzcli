package daemon

import (
	"encoding/json"
	"fmt"

	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/watchertypes"
	"github.com/dimasma0305/gzcli/internal/log"
)

// ShowStatus displays the watcher status
func ShowStatus(pidFile, logFile string, jsonOutput bool) error {
	daemonStatus := GetDaemonStatus(pidFile)
	isDaemon := daemonStatus["daemon"].(bool)
	daemonState := daemonStatus["status"].(string)

	log.Info("🔍 GZCTF Watcher Status")
	log.Info("==========================================")

	switch {
	case isDaemon && daemonState == "running":
		log.Info("🟢 Status: RUNNING (Daemon Mode)")
		if pid, ok := daemonStatus["pid"]; ok {
			log.Info("📄 Process ID: %v", pid)
		}
		log.Info("📄 PID File: %s", pidFile)
		log.Info("📝 Log File: %s", logFile)

		// For running daemon, try to get challenge info from status
		log.Info("")
		log.Info("📁 Configuration:")
		log.Info("   - Daemon Mode: Enabled")
		log.Info("   - PID File: %s", pidFile)
		log.Info("   - Log File: %s", logFile)
		log.Info("   - Git Pull: %v", watchertypes.DefaultWatcherConfig.GitPullEnabled)
		if watchertypes.DefaultWatcherConfig.GitPullEnabled {
			log.Info("   - Git Pull Interval: %v", watchertypes.DefaultWatcherConfig.GitPullInterval)
			log.Info("   - Git Repository: %s", watchertypes.DefaultWatcherConfig.GitRepository)
		}

		// Show recent log entries if available
		ShowRecentLogs(logFile)

	case daemonState == "dead":
		log.Info("🟡 Status: STOPPED (Stale PID file found)")
		log.Info("💬 A previous daemon process was running but is no longer active")
		log.Info("📄 Stale PID File: %s", pidFile)
		log.Info("🔧 Suggestion: Run 'gzcli --watch' to start a new daemon")

	case daemonState == "stopped":
		log.Info("⚫ Status: NOT RUNNING")
		log.Info("💬 No daemon is currently running")
		log.Info("📄 PID File: %s (not found)", pidFile)
		log.Info("🔧 Suggestion: Run 'gzcli --watch' to start the daemon")

	default:
		log.Info("🔴 Status: ERROR")
		if msg, ok := daemonStatus["message"]; ok {
			log.Info("💬 %s", msg)
		}
		log.Info("📄 PID File: %s", pidFile)
	}

	log.Info("")
	log.Info("🛠️  Available Commands:")
	log.Info("   - Start daemon: gzcli --watch")
	log.Info("   - Stop daemon:  gzcli --watch-stop")
	log.Info("   - Run foreground: gzcli --watch --watch-foreground")

	// Output JSON format if requested
	if jsonOutput {
		return outputStatusJSON(daemonStatus, pidFile, logFile, isDaemon, daemonState)
	}

	return nil
}

// outputStatusJSON outputs status in JSON format
func outputStatusJSON(daemonStatus map[string]interface{}, pidFile, logFile string, isDaemon bool, daemonState string) error {
	// Create a cleaner status object for JSON
	jsonStatus := map[string]interface{}{
		"daemon_running": isDaemon && daemonState == "running",
		"status":         daemonState,
		"pid_file":       pidFile,
		"log_file":       logFile,
	}

	if isDaemon && daemonState == "running" {
		if pid, ok := daemonStatus["pid"]; ok {
			jsonStatus["pid"] = pid
		}
	}

	if msg, ok := daemonStatus["message"]; ok {
		jsonStatus["message"] = msg
	}

	log.Info("")
	jsonData, err := json.MarshalIndent(jsonStatus, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal status to JSON: %w", err)
	}
	fmt.Println(string(jsonData))
	return nil
}
