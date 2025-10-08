package cmd

import (
	"github.com/spf13/cobra"

	"github.com/dimasma0305/gzcli/internal/gzcli"
	"github.com/dimasma0305/gzcli/internal/log"
)

var (
	structureEvents        []string
	structureExcludeEvents []string
)

var structureCmd = &cobra.Command{
	Use:   "structure",
	Short: "Generate challenge directory structure",
	Long: `Generate directory structure for each challenge folder based on .structure template file.

This command reads the .structure file in the challenge directory and creates
the specified directory structure and placeholder files.

By default, generates structures for all events. Use --event to specify specific events,
or --exclude-event to exclude certain events.`,
	Example: `  # Generate structure for all events
  gzcli structure

  # Generate structure for specific events
  gzcli structure --event ctf2024 --event ctf2025

  # Generate structure for all except practice event
  gzcli structure --exclude-event practice`,
	Run: func(_ *cobra.Command, _ []string) {
		// Resolve which events to generate structure for
		events, err := ResolveTargetEvents(structureEvents, structureExcludeEvents)
		if err != nil {
			log.Error("Failed to resolve target events: %v", err)
			return
		}

		// Track results
		successCount := 0
		failureCount := 0
		var failedEvents []string

		log.Info("Generating structure for %d event(s): %v", len(events), events)

		// Generate structure for each event
		for _, eventName := range events {
			log.InfoH2("[%s] Generating challenge structures...", eventName)

			gz, err := gzcli.InitWithEvent(eventName)
			if err != nil {
				log.Error("[%s] Failed to initialize: %v", eventName, err)
				failureCount++
				failedEvents = append(failedEvents, eventName)
				continue
			}

			if err := gz.GenerateStructure(); err != nil {
				log.Error("[%s] Failed to generate structure: %v", eventName, err)
				failureCount++
				failedEvents = append(failedEvents, eventName)
			} else {
				log.Info("[%s] Challenge structures generated successfully", eventName)
				successCount++
			}
		}

		// Display summary
		log.Info("Structure Generation Summary: %d succeeded, %d failed", successCount, failureCount)
		if failureCount > 0 {
			log.Error("Failed events: %v", failedEvents)
			log.Fatal("Some events failed to generate structure")
		}
	},
}

func init() {
	rootCmd.AddCommand(structureCmd)

	structureCmd.Flags().StringSliceVarP(&structureEvents, "event", "e", []string{}, "Specific event(s) to generate structure for (can be specified multiple times)")
	structureCmd.Flags().StringSliceVar(&structureExcludeEvents, "exclude-event", []string{}, "Event(s) to exclude from structure generation (can be specified multiple times)")
}
