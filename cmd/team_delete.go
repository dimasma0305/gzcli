package cmd

import (
	"github.com/dimasma0305/gzcli/internal/gzcli"
	"github.com/dimasma0305/gzcli/internal/log"
	"github.com/spf13/cobra"
)

var (
	deleteAll bool
)

var teamDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete teams and users",
	Long:  `Delete all teams and users from the CTF platform.`,
	Example: `  # Delete all teams and users
  gzcli team delete --all`,
	Run: func(cmd *cobra.Command, args []string) {
		if !deleteAll {
			log.Error("Please specify --all flag to confirm deletion")
			cmd.Help()
			return
		}

		gz := gzcli.MustInit()

		if err := gz.DeleteAllUser(); err != nil {
			log.Fatal("User deletion failed: ", err)
		}

		log.Info("All teams and users deleted successfully!")
	},
}

func init() {
	teamCmd.AddCommand(teamDeleteCmd)

	teamDeleteCmd.Flags().BoolVar(&deleteAll, "all", false, "Confirm deletion of all teams and users")
}
