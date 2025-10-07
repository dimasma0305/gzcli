package cmd

import (
	"github.com/spf13/cobra"

	"github.com/dimasma0305/gzcli/internal/gzcli"
	"github.com/dimasma0305/gzcli/internal/log"
)

var (
	stopPidFile string
)

var watchStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the file watcher daemon",
	Long:  `Stop the running file watcher daemon.`,
	Example: `  # Stop the watcher
  gzcli watch stop

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

		log.Info("🛑 Stopping GZCTF Watcher daemon...")
		if err := watcher.StopDaemon(pidFile); err != nil {
			log.Fatal("Failed to stop daemon: ", err)
		}
	},
}

func init() {
	watchCmd.AddCommand(watchStopCmd)

	watchStopCmd.Flags().StringVar(&stopPidFile, "pid-file", "", "Custom PID file location")
}
