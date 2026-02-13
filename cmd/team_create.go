package cmd

import (
	"github.com/spf13/cobra"

	"github.com/dimasma0305/gzcli/internal/gzcli"
	"github.com/dimasma0305/gzcli/internal/log"
)

var (
	createSendEmail         bool
	createEventID           int
	createInviteCode        string
	createForceInitMapping  bool
	createCommunicationType string
	createCommunicationLink string
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
  gzcli team create teams.csv --send-email

  # Create teams into specific event
  gzcli team create teams.csv --event-id 1 --invite-code "secret"`,
	Args: cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		csvFile := args[0]
		// Use event from flag if provided
		gz, err := gzcli.InitWithEvent(GetEventFlag())
		if err != nil {
			log.Error("Failed to initialize: %v", err)
			return
		}

		if err := gz.CreateTeams(csvFile, createSendEmail, createEventID, createInviteCode, createForceInitMapping, createCommunicationType, createCommunicationLink); err != nil {
			log.Fatal(err)
		}

		log.Info("Teams created successfully!")
		log.InfoH2("IMPORTANT: Do not change the account username and password.")
	},
}

func init() {
	teamCmd.AddCommand(teamCreateCmd)

	teamCreateCmd.Flags().BoolVar(&createSendEmail, "send-email", false, "Send registration emails to teams")
	teamCreateCmd.Flags().IntVar(&createEventID, "event-id", 0, "Specify the event ID to add teams to")
	teamCreateCmd.Flags().StringVar(&createInviteCode, "invite-code", "", "Specify the invite code for the event")
	teamCreateCmd.Flags().BoolVar(&createForceInitMapping, "force-init-mapping", false, "Force initialization of column mapping")
	teamCreateCmd.Flags().StringVar(&createCommunicationType, "communication-type", "", "Global communication type for all team emails (e.g. Discord, WhatsApp)")
	teamCreateCmd.Flags().StringVar(&createCommunicationLink, "communication-link", "", "Global communication link for all team emails")
}
