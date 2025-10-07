package cmd

import (
	"github.com/spf13/cobra"

	"github.com/dimasma0305/gzcli/internal/gzcli"
	"github.com/dimasma0305/gzcli/internal/log"
)

var (
	logsFile string
)

var watchLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Follow and display watcher logs in real-time",
	Long:  `Stream the file watcher daemon log file in real-time (like tail -f).`,
	Example: `  # View logs
  gzcli watch logs

  # View custom log file
  gzcli watch logs --log-file /custom/path/watcher.log`,
	Run: func(_ *cobra.Command, _ []string) {
		gz := gzcli.MustInit()

		watcher, err := gzcli.NewWatcher(gz)
		if err != nil {
			log.Fatal("Failed to create watcher: ", err)
		}

		logFile := gzcli.DefaultWatcherConfig.LogFile
		if logsFile != "" {
			logFile = logsFile
		}

		log.Info("📋 Following GZCTF Watcher logs: %s", logFile)
		log.Info("Press Ctrl+C to stop following logs")
		log.Info("==========================================")

		if err := watcher.FollowLogs(logFile); err != nil {
			log.Fatal("Failed to follow logs: ", err)
		}
	},
}

func init() {
	watchCmd.AddCommand(watchLogsCmd)

	watchLogsCmd.Flags().StringVar(&logsFile, "log-file", "", "Custom log file location")
}
