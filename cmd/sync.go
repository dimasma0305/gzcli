package cmd

import (
	"github.com/spf13/cobra"

	"github.com/dimasma0305/gzcli/internal/gzcli"
)

var (
	syncUpdateGame bool
)

var syncCmd = &cobra.Command{
	Use:     "sync",
	Aliases: []string{"s"},
	Short:   "Synchronize CTF challenges with the server",
	Long: `Synchronize local challenge configurations with the GZ::CTF server.

This command:
  - Reads challenge configurations from local directories
  - Creates or updates challenges on the server
  - Uploads attachments and container images
  - Syncs challenge visibility and scoring`,
	Example: `  # Sync challenges
  gzcli sync

  # Sync and update game configuration
  gzcli sync --update-game`,
	Run: func(_ *cobra.Command, _ []string) {
		// Use event from flag if provided
		gz, err := gzcli.InitWithEvent(GetEventFlag())
		if err != nil {
			panic(err)
		}
		gz.UpdateGame = syncUpdateGame
		gz.MustSync()
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().BoolVar(&syncUpdateGame, "update-game", false, "Update game configuration during sync")
}
