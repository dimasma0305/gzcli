package cmd

import (
	"github.com/spf13/cobra"

	"github.com/dimasma0305/gzcli/internal/gzcli"
	"github.com/dimasma0305/gzcli/internal/log"
)

var (
	createSendEmail bool
)

var teamCreateCmd = &cobra.Command{
	Use:   "create <csv-file>",
	Short: "Create teams from a CSV file",
	Long: `Create teams from a CSV file containing team information.

The CSV file should have the following format:
  RealName,Email,TeamName

Example:
  John Doe,john@example.com,TeamAlpha
  Jane Smith,jane@example.com,TeamBeta`,
	Example: `  # Create teams from CSV
  gzcli team create teams.csv

  # Create teams and send registration emails
  gzcli team create teams.csv --send-email`,
	Args: cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		csvFile := args[0]
		// Use event from flag if provided
		gz, err := gzcli.InitWithEvent(GetEventFlag())
		if err != nil {
			log.Error("Failed to initialize: %v", err)
			return
		}

		if err := gz.CreateTeams(csvFile, createSendEmail); err != nil {
			log.Fatal(err)
		}

		log.Info("Teams created successfully!")
	},
}

func init() {
	teamCmd.AddCommand(teamCreateCmd)

	teamCreateCmd.Flags().BoolVar(&createSendEmail, "send-email", false, "Send registration emails to teams")
}
