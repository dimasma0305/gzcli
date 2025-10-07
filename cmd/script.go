package cmd

import (
	"github.com/spf13/cobra"

	"github.com/dimasma0305/gzcli/internal/gzcli"
	"github.com/dimasma0305/gzcli/internal/log"
)

var scriptCmd = &cobra.Command{
	Use:   "script <name>",
	Short: "Execute a custom script defined in challenge configurations",
	Long: `Execute a custom script across all challenges that define it.

Scripts are defined in challenge.yaml files under the 'scripts' section.
This command will run the specified script for all challenges that have it defined.`,
	Example: `  # Run the 'deploy' script
  gzcli script deploy

  # Run the 'test' script
  gzcli script test

  # Run the 'cleanup' script
  gzcli script cleanup`,
	Args: cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		scriptName := args[0]

		log.Info("Running script: %s", scriptName)
		if err := gzcli.RunScripts(scriptName); err != nil {
			log.Fatal("Script execution failed: ", err)
		}

		log.Info("Script '%s' executed successfully!", scriptName)
	},
}

func init() {
	rootCmd.AddCommand(scriptCmd)
}
