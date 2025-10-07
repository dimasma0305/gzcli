package cmd

import (
	"github.com/spf13/cobra"

	"github.com/dimasma0305/gzcli/internal/gzcli"
	"github.com/dimasma0305/gzcli/internal/log"
)

var (
	statusPidFile string
	statusLogFile string
	statusJSON    bool
)

var watchStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show file watcher status",
	Long:  `Display the current status of the file watcher daemon.`,
	Example: `  # Show status
  gzcli watch status

  # Show status in JSON format
  gzcli watch status --json`,
	Run: func(_ *cobra.Command, _ []string) {
		gz := gzcli.MustInit()

		watcher, err := gzcli.NewWatcher(gz)
		if err != nil {
			log.Fatal("Failed to create watcher: ", err)
		}

		pidFile := gzcli.DefaultWatcherConfig.PidFile
		if statusPidFile != "" {
			pidFile = statusPidFile
		}

		logFile := gzcli.DefaultWatcherConfig.LogFile
		if statusLogFile != "" {
			logFile = statusLogFile
		}

		if err := watcher.ShowStatus(pidFile, logFile, statusJSON); err != nil {
			log.Error("Failed to show status: %v", err)
		}
	},
}

func init() {
	watchCmd.AddCommand(watchStatusCmd)

	watchStatusCmd.Flags().StringVar(&statusPidFile, "pid-file", "", "Custom PID file location")
	watchStatusCmd.Flags().StringVar(&statusLogFile, "log-file", "", "Custom log file location")
	watchStatusCmd.Flags().BoolVar(&statusJSON, "json", false, "Output status in JSON format")
}
