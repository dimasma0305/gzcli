package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dimasma0305/gzcli/internal/gzcli"
	"github.com/dimasma0305/gzcli/internal/log"
)

var (
	statusPidFile    string
	statusLogFile    string
	statusJSON       bool
	statusEvent      string
	statusSocketPath string
)

var watchStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show file watcher status",
	Long:  `Display the current status of the file watcher daemon, optionally filtered by event.`,
	Example: `  # Show status for all events
  gzcli watch status

  # Show status for a specific event
  gzcli watch status --event ctf2024

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

		socketPath := gzcli.DefaultWatcherConfig.SocketPath
		if statusSocketPath != "" {
			socketPath = statusSocketPath
		}

		// If requesting status for a specific event, use socket command
		if statusEvent != "" {
			client := gzcli.NewWatcherClient(socketPath)
			response, err := client.SendCommand("status", map[string]interface{}{
				"event": statusEvent,
			})
			if err != nil {
				log.Fatal("Failed to communicate with watcher daemon: ", err)
			}

			if !response.Success {
				log.Fatal("Failed to get status: ", response.Error)
			}

			// Print the response
			if statusJSON {
				// Print raw JSON
				fmt.Printf("%+v\n", response.Data)
			} else {
				log.Info("Status for event '%s':", statusEvent)
				fmt.Printf("%+v\n", response.Data)
			}
			return
		}

		// Otherwise, use the default ShowStatus which shows daemon-level info
		if err := watcher.ShowStatus(pidFile, logFile, statusJSON); err != nil {
			log.Error("Failed to show status: %v", err)
		}
	},
}

func init() {
	watchCmd.AddCommand(watchStatusCmd)

	watchStatusCmd.Flags().StringVar(&statusEvent, "event", "", "Show status for a specific event")
	watchStatusCmd.Flags().StringVar(&statusPidFile, "pid-file", "", "Custom PID file location")
	watchStatusCmd.Flags().StringVar(&statusLogFile, "log-file", "", "Custom log file location")
	watchStatusCmd.Flags().StringVar(&statusSocketPath, "socket", "", "Custom socket file location")
	watchStatusCmd.Flags().BoolVar(&statusJSON, "json", false, "Output status in JSON format")

	// Register completion for --event flag
	_ = watchStatusCmd.RegisterFlagCompletionFunc("event", validEventNames)
}
