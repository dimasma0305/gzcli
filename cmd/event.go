package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/log"
)

var eventCmd = &cobra.Command{
	Use:   "event",
	Short: "Manage CTF events",
	Long: `Manage multiple CTF events in your workspace.

Events are stored in the events/ directory, each with their own configuration
and challenges. You can switch between events, list available events, and
create new ones.`,
	Example: `  # List all events
  gzcli event list

  # Switch to a specific event
  gzcli event switch ctf2024

  # Show current event
  gzcli event current

  # Create a new event
  gzcli event create ctf2025`,
}

var eventListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available events",
	Long:  `List all events in the events/ directory that have a valid .gzevent configuration file.`,
	Run: func(_ *cobra.Command, _ []string) {
		events, err := config.ListEvents()
		if err != nil {
			log.Error("Failed to list events: %v", err)
			return
		}

		if len(events) == 0 {
			log.Info("No events found. Run 'gzcli event create <name>' to create one")
			return
		}

		// Get current event (if set)
		currentEvent, _ := config.GetCurrentEvent("")

		log.Info("Available events:")
		for _, event := range events {
			if event == currentEvent {
				log.Info("  • %s (current)", event)
			} else {
				log.Info("  • %s", event)
			}
		}
	},
}

var eventCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the current active event",
	Long:  `Display which event is currently active based on flags, environment variables, or default settings.`,
	Run: func(_ *cobra.Command, _ []string) {
		currentEvent, err := config.GetCurrentEvent(GetEventFlag())
		if err != nil {
			log.Error("Failed to determine current event: %v", err)
			log.Info("Use 'gzcli event switch <name>' to set a default event")
			return
		}

		log.Info("Current event: %s", currentEvent)

		// Show how it was determined
		if GetEventFlag() != "" {
			log.Info("(set via --event flag)")
		} else if envEvent := config.GetEnvEvent(); envEvent != "" {
			log.Info("(set via GZCLI_EVENT environment variable)")
		} else {
			log.Info("(auto-detected or set as default)")
		}
	},
}

var eventSwitchCmd = &cobra.Command{
	Use:   "switch [event-name]",
	Short: "Switch to a different event as the default",
	Long: `Set a specific event as the default event for all commands.
This creates/updates the .gzcli/current-event file.`,
	Args: cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		eventName := args[0]

		if err := config.SetCurrentEvent(eventName); err != nil {
			log.Error("Failed to switch event: %v", err)
			return
		}

		log.Info("✅ Switched to event: %s", eventName)
	},
}

var eventCreateCmd = &cobra.Command{
	Use:   "create [event-name]",
	Short: "Create a new event",
	Long: `Create a new event directory with a .gzevent configuration file.

This command will:
  • Create events/[name]/ directory
  • Create a template .gzevent file
  • Initialize challenge category directories
  • Optionally initialize a git repository`,
	Args: cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		eventName := args[0]

		log.Info("Creating new event: %s", eventName)
		log.Error("Event creation not yet implemented. Use 'gzcli init' for now")
		fmt.Printf("TODO: Implement event creation for: %s\n", eventName)
	},
}

func init() {
	rootCmd.AddCommand(eventCmd)
	eventCmd.AddCommand(eventListCmd)
	eventCmd.AddCommand(eventCurrentCmd)
	eventCmd.AddCommand(eventSwitchCmd)
	eventCmd.AddCommand(eventCreateCmd)
}
