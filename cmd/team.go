package cmd

import (
	"github.com/spf13/cobra"
)

var teamCmd = &cobra.Command{
	Use:     "team",
	Aliases: []string{"t"},
	Short:   "Team management operations",
	Long: `Manage teams for your CTF including:
  - Creating teams from CSV files
  - Sending registration emails
  - Registering teams to games
  - Deleting teams and users`,
	Example: `  # Create teams from CSV
  gzcli team create teams.csv

  # Create teams and send emails
  gzcli team create teams.csv --send-email

  # Register teams to a game
  gzcli team register teams.csv --game "My CTF" --division "Open"

  # Delete all teams and users
  gzcli team delete --all`,
}

func init() {
	rootCmd.AddCommand(teamCmd)
}
