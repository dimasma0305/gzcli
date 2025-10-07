package cmd

import (
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:     "watch",
	Aliases: []string{"w"},
	Short:   "File watcher operations for automatic challenge redeployment",
	Long: `Manage the file watcher daemon that automatically redeploys challenges when files change.

The watcher monitors challenge directories for changes and automatically:
  - Syncs updated challenges to the server
  - Restarts containers when needed
  - Executes custom scripts defined in challenge.yaml
  - Performs automatic git pull operations`,
	Example: `  # Start watcher daemon
  gzcli watch start

  # Check watcher status
  gzcli watch status

  # Stop watcher daemon
  gzcli watch stop

  # View watcher logs
  gzcli watch logs`,
}

func init() {
	rootCmd.AddCommand(watchCmd)
}
