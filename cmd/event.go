package cmd

import (
	"github.com/spf13/cobra"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/log"
	"github.com/dimasma0305/gzcli/internal/template/other"
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
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: validEventNames,
	Run: func(_ *cobra.Command, args []string) {
		eventName := args[0]

		if err := config.SetCurrentEvent(eventName); err != nil {
			log.Error("Failed to switch event: %v", err)
			return
		}

		log.Info("✅ Switched to event: %s", eventName)
	},
}

var (
	eventCreateTitle      string
	eventCreateStart      string
	eventCreateEnd        string
	eventCreateSetCurrent bool
)

var eventCreateCmd = &cobra.Command{
	Use:   "create [event-name]",
	Short: "Create a new event",
	Long: `Create a new event directory with a .gzevent configuration file.

This command will:
  • Create events/[name]/ directory
  • Create a template .gzevent file
  • Initialize challenge category directories
  • Optionally set as the current event`,
	Args: cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		eventName := args[0]

		log.Info("Creating new event: %s", eventName)

		// Prepare event info from flags or prompt user
		eventInfo := map[string]string{
			"title": eventCreateTitle,
			"start": eventCreateStart,
			"end":   eventCreateEnd,
		}

		// Create the event structure
		if errors := other.EventTemplate(".", eventName, eventInfo); errors != nil {
			for _, err := range errors {
				if err != nil {
					log.Error("%s", err)
				}
			}
			return
		}

		log.Info("✅ Event '%s' created successfully!", eventName)

		// Set as current event if flag is set or if it's the only event
		shouldSetCurrent := eventCreateSetCurrent
		if !shouldSetCurrent {
			// Auto-set as current if this is the only event
			events, err := config.ListEvents()
			if err == nil && len(events) == 1 {
				shouldSetCurrent = true
			}
		}

		if shouldSetCurrent {
			if err := config.SetCurrentEvent(eventName); err != nil {
				log.Error("Failed to set as current event: %v", err)
			} else {
				log.Info("✅ Set as current event")
			}
		} else {
			log.Info("Run 'gzcli event switch %s' to set it as the current event", eventName)
		}

		log.Info("\nNext steps:")
		log.Info("  1. Review the event configuration: events/%s/.gzevent", eventName)
		log.Info("  2. Add challenges to category directories")
		log.Info("  3. Run 'gzcli structure' to generate challenge structure")
	},
}

func init() {
	rootCmd.AddCommand(eventCmd)
	eventCmd.AddCommand(eventListCmd)
	eventCmd.AddCommand(eventCurrentCmd)
	eventCmd.AddCommand(eventSwitchCmd)
	eventCmd.AddCommand(eventCreateCmd)

	// Add flags for event create command
	eventCreateCmd.Flags().StringVar(&eventCreateTitle, "title", "", "Event title (default: prompt user)")
	eventCreateCmd.Flags().StringVar(&eventCreateStart, "start", "", "Start date in RFC3339 format (default: prompt user)")
	eventCreateCmd.Flags().StringVar(&eventCreateEnd, "end", "", "End date in RFC3339 format (default: prompt user)")
	eventCreateCmd.Flags().BoolVar(&eventCreateSetCurrent, "set-current", false, "Set as current event (default: auto if only event)")
}
