package cmd

import (
	"github.com/spf13/cobra"

	"github.com/dimasma0305/gzcli/internal/gzcli"
	"github.com/dimasma0305/gzcli/internal/log"
)

var (
	stopPidFile    string
	stopEvent      string
	stopSocketPath string
)

var watchStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the file watcher daemon or a specific event",
	Long:  `Stop the running file watcher daemon, or stop watching a specific event while keeping the daemon running.`,
	Example: `  # Stop the entire watcher daemon
  gzcli watch stop

  # Stop watching a specific event (daemon continues running other events)
  gzcli watch stop --event ctf2024

  # Stop with custom PID file
  gzcli watch stop --pid-file /custom/path/watcher.pid`,
	Run: func(_ *cobra.Command, _ []string) {
		gz := gzcli.MustInit()

		watcher, err := gzcli.NewWatcher(gz)
		if err != nil {
			log.Fatal("Failed to create watcher: ", err)
		}

		pidFile := gzcli.DefaultWatcherConfig.PidFile
		if stopPidFile != "" {
			pidFile = stopPidFile
		}

		socketPath := gzcli.DefaultWatcherConfig.SocketPath
		if stopSocketPath != "" {
			socketPath = stopSocketPath
		}

		// If a specific event is specified, stop only that event via socket
		if stopEvent != "" {
			log.Info("🛑 Stopping event watcher for: %s", stopEvent)
			client := gzcli.NewWatcherClient(socketPath)

			// Send stop event command via socket
			response, err := client.SendCommand("stop_event", map[string]interface{}{
				"event": stopEvent,
			})
			if err != nil {
				log.Fatal("Failed to communicate with watcher daemon: ", err)
			}

			if response.Success {
				log.Info("✅ Event watcher for '%s' stopped successfully", stopEvent)
			} else {
				log.Fatal("Failed to stop event watcher: ", response.Error)
			}
			return
		}

		// Otherwise, stop the entire daemon
		log.Info("🛑 Stopping GZCTF Watcher daemon...")
		if err := watcher.StopDaemon(pidFile); err != nil {
			log.Fatal("Failed to stop daemon: ", err)
		}
	},
}

func init() {
	watchCmd.AddCommand(watchStopCmd)

	watchStopCmd.Flags().StringVar(&stopEvent, "event", "", "Stop watching a specific event (keeps daemon running)")
	watchStopCmd.Flags().StringVar(&stopPidFile, "pid-file", "", "Custom PID file location")
	watchStopCmd.Flags().StringVar(&stopSocketPath, "socket", "", "Custom socket file location")

	// Register completion for --event flag
	_ = watchStopCmd.RegisterFlagCompletionFunc("event", validEventNames)
}
