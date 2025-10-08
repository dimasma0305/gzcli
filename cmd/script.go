package cmd

import (
	"github.com/spf13/cobra"

	"github.com/dimasma0305/gzcli/internal/gzcli"
	"github.com/dimasma0305/gzcli/internal/log"
)

var (
	scriptEvents        []string
	scriptExcludeEvents []string
)

var scriptCmd = &cobra.Command{
	Use:   "script <name>",
	Short: "Execute a custom script defined in challenge configurations",
	Long: `Execute a custom script across all challenges that define it.

Scripts are defined in challenge.yaml files under the 'scripts' section.
This command will run the specified script for all challenges that have it defined.

By default, runs scripts for all events. Use --event to specify specific events,
or --exclude-event to exclude certain events.`,
	Example: `  # Run the 'deploy' script for all events
  gzcli script deploy

  # Run the 'test' script for specific events
  gzcli script test --event ctf2024 --event ctf2025

  # Run the 'cleanup' script for all except practice event
  gzcli script cleanup --exclude-event practice`,
	Args: cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		scriptName := args[0]

		// Resolve which events to run scripts for
		events, err := ResolveTargetEvents(scriptEvents, scriptExcludeEvents)
		if err != nil {
			log.Error("Failed to resolve target events: %v", err)
			log.Fatal(err)
		}

		// Track results
		successCount := 0
		failureCount := 0
		var failedEvents []string

		log.Info("Running script '%s' for %d event(s): %v", scriptName, len(events), events)

		// Run script for each event
		for _, eventName := range events {
			log.InfoH2("[%s] Running script '%s'...", eventName, scriptName)

			if err := gzcli.RunScripts(scriptName, eventName); err != nil {
				log.Error("[%s] Script execution failed: %v", eventName, err)
				failureCount++
				failedEvents = append(failedEvents, eventName)
			} else {
				log.Info("[%s] Script '%s' executed successfully", eventName, scriptName)
				successCount++
			}
		}

		// Display summary
		log.Info("Script Execution Summary: %d succeeded, %d failed", successCount, failureCount)
		if failureCount > 0 {
			log.Error("Failed events: %v", failedEvents)
			log.Fatal("Some events failed to execute script")
		}
	},
}

func init() {
	rootCmd.AddCommand(scriptCmd)

	scriptCmd.Flags().StringSliceVarP(&scriptEvents, "event", "e", []string{}, "Specific event(s) to run script for (can be specified multiple times)")
	scriptCmd.Flags().StringSliceVar(&scriptExcludeEvents, "exclude-event", []string{}, "Event(s) to exclude from script execution (can be specified multiple times)")
}
