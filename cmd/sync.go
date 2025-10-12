package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/dimasma0305/gzcli/internal/gzcli"
	"github.com/dimasma0305/gzcli/internal/log"
)

var (
	syncUpdateGame    bool
	syncEvents        []string
	syncExcludeEvents []string
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
  - Syncs challenge visibility and scoring

By default, syncs all events. Use --event to specify specific events,
or --exclude-event to exclude certain events.`,
	Example: `  # Sync all events
  gzcli sync

  # Sync specific events
  gzcli sync --event ctf2024 --event ctf2025

  # Sync all except specific events
  gzcli sync --exclude-event practice

  # Sync and update game configuration
  gzcli sync --update-game`,
	Run: func(_ *cobra.Command, _ []string) {
		// Resolve which events to sync
		events, err := ResolveTargetEvents(syncEvents, syncExcludeEvents)
		if err != nil {
			log.Error("Failed to resolve target events: %v", err)
			os.Exit(1)
		}

		// Track results
		successCount := 0
		failureCount := 0
		type failedEvent struct {
			name string
			err  error
		}
		var failedEvents []failedEvent

		log.Info("Syncing %d event(s): %v", len(events), events)

		// Sync each event
		for _, eventName := range events {
			log.InfoH2("[%s] Starting sync...", eventName)

			gz, err := gzcli.InitWithEvent(eventName)
			if err != nil {
				log.Error("[%s] Failed to initialize: %v", eventName, err)
				failureCount++
				failedEvents = append(failedEvents, failedEvent{name: eventName, err: err})
				continue
			}

			gz.UpdateGame = syncUpdateGame
			if err := gz.Sync(); err != nil {
				log.Error("[%s] Sync failed: %v", eventName, err)
				failureCount++
				failedEvents = append(failedEvents, failedEvent{name: eventName, err: err})
			} else {
				log.Info("[%s] Sync completed successfully", eventName)
				successCount++
			}
		}

		// Display summary
		log.InfoH2("Sync Summary: %d succeeded, %d failed", successCount, failureCount)
		if failureCount > 0 {
			log.Error("Failed events:")
			for _, fe := range failedEvents {
				log.Error("  - %s: %v", fe.name, fe.err)
			}
			log.Error("\nPlease check:")
			log.Error("  1. Event directories exist in events/")
			log.Error("  2. Each event has a valid .gzevent configuration file")
			log.Error("  3. Server is accessible and credentials are correct")
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().BoolVar(&syncUpdateGame, "update-game", false, "Update game configuration during sync")
	syncCmd.Flags().StringSliceVarP(&syncEvents, "event", "e", []string{}, "Specific event(s) to sync (can be specified multiple times)")
	syncCmd.Flags().StringSliceVar(&syncExcludeEvents, "exclude-event", []string{}, "Event(s) to exclude from sync (can be specified multiple times)")
}
