package cmd

import (
	"github.com/spf13/cobra"

	"github.com/dimasma0305/gzcli/internal/gzcli"
	"github.com/dimasma0305/gzcli/internal/log"
)

var structureCmd = &cobra.Command{
	Use:   "structure",
	Short: "Generate challenge directory structure",
	Long: `Generate directory structure for each challenge folder based on .structure template file.

This command reads the .structure file in the challenge directory and creates
the specified directory structure and placeholder files.`,
	Example: `  # Generate structure for all challenges
  gzcli structure`,
	Run: func(_ *cobra.Command, _ []string) {
		// Use event from flag if provided
		gz, err := gzcli.InitWithEvent(GetEventFlag())
		if err != nil {
			log.Error("Failed to initialize: %v", err)
			return
		}

		if err := gz.GenerateStructure(); err != nil {
			log.Fatal("Failed to generate structure: ", err)
		}

		log.Info("Challenge structures generated successfully!")
	},
}

func init() {
	rootCmd.AddCommand(structureCmd)
}
