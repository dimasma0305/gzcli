package cmd

import (
	"strconv"
	"strings"
	"time"

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
	eventCreateTitle string
	eventCreateStart string
	eventCreateEnd   string
)

var eventCreateCmd = &cobra.Command{
	Use:   "create [event-name]",
	Short: "Create a new event",
	Long: `Create a new event directory with a .gzevent configuration file.

This command will:
  • Create events/[name]/ directory
  • Create a template .gzevent file with provided details
  • Initialize challenge category directories
  • Auto-set as current event if it's the only event`,
	Example: `  # Create event with required flags
  gzcli event create myctf2024 --title "My CTF 2024" --start 2024-12-01T00:00:00Z --end 2024-12-03T00:00:00Z

  # Use TAB to autocomplete event names and dates
  gzcli event create <TAB>`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: eventNameCompletion,
	Run: func(cmd *cobra.Command, args []string) {
		eventName := args[0]

		// Validate required flags
		if eventCreateTitle == "" {
			log.Error("--title flag is required")
			_ = cmd.Usage()
			return
		}
		if eventCreateStart == "" {
			log.Error("--start flag is required")
			_ = cmd.Usage()
			return
		}
		if eventCreateEnd == "" {
			log.Error("--end flag is required")
			_ = cmd.Usage()
			return
		}

		log.Info("Creating new event: %s", eventName)

		// Prepare event info from flags
		eventInfo := map[string]string{
			"title": eventCreateTitle,
			"start": eventCreateStart,
			"end":   eventCreateEnd,
		}

		// Create the event structure
		// Note: Template errors for example files are expected and can be ignored
		// (they contain {{.slug}}, {{.host}} etc. that are meant to be filled in later)
		if errors := other.EventTemplate(".", eventName, eventInfo); errors != nil {
			// Only fail if we have real errors (not template processing errors)
			hasRealErrors := false
			for _, err := range errors {
				if err != nil {
					// Skip template processing errors for example files
					errStr := err.Error()
					if !containsAny(errStr, []string{"template processing error", ".example/", ".structure/"}) {
						log.Error("%s", err)
						hasRealErrors = true
					}
				}
			}
			if hasRealErrors {
				return
			}
		}

		log.Info("✅ Event '%s' created successfully!", eventName)

		// Auto-set as current if this is the only event
		events, err := config.ListEvents()
		if err == nil && len(events) == 1 {
			if err := config.SetCurrentEvent(eventName); err != nil {
				log.Error("Failed to set as current event: %v", err)
			} else {
				log.Info("✅ Set as current event (auto-detected as only event)")
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

// containsAny checks if the string contains any of the substrings
func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

// dateCompletionContext holds context for date completion
type dateCompletionContext struct {
	isEndFlag bool
	startDate time.Time
	now       time.Time
}

// getDateCompletionContext extracts context from command flags
func getDateCompletionContext(cmd *cobra.Command) dateCompletionContext {
	ctx := dateCompletionContext{
		now: time.Now(),
	}

	if flagValue, _ := cmd.Flags().GetString("start"); flagValue != "" {
		ctx.isEndFlag = true
		if parsedStart, err := time.Parse(time.RFC3339, flagValue); err == nil {
			ctx.startDate = parsedStart
		}
	}

	return ctx
}

// completeYear suggests year values (YYYY-)
func completeYear(ctx dateCompletionContext) ([]string, cobra.ShellCompDirective) {
	currentYear := ctx.now.Year()
	minYear := currentYear

	if ctx.isEndFlag && !ctx.startDate.IsZero() {
		minYear = ctx.startDate.Year()
	}

	suggestions := []string{}
	for year := minYear; year <= currentYear+2; year++ {
		suggestions = append(suggestions, strconv.Itoa(year)+"-")
	}
	return suggestions, cobra.ShellCompDirectiveNoSpace
}

// completeMonth suggests month values (YYYY-MM-)
func completeMonth(toComplete string, ctx dateCompletionContext) ([]string, cobra.ShellCompDirective) {
	yearStr := toComplete[:4]
	year, _ := strconv.Atoi(yearStr)
	currentYear := ctx.now.Year()
	currentMonth := int(ctx.now.Month())

	// Determine minimum month
	minMonth := 1
	if !ctx.isEndFlag && year == currentYear {
		minMonth = currentMonth
	} else if ctx.isEndFlag && !ctx.startDate.IsZero() && year == ctx.startDate.Year() {
		minMonth = int(ctx.startDate.Month())
	}

	suggestions := []string{}
	for month := minMonth; month <= 12; month++ {
		suggestions = append(suggestions, yearStr+"-"+padZero(month)+"-")
	}
	return suggestions, cobra.ShellCompDirectiveNoSpace
}

// completeDay suggests day values (YYYY-MM-DDT)
func completeDay(toComplete string, ctx dateCompletionContext) ([]string, cobra.ShellCompDirective) {
	prefix := toComplete[:8]
	yearStr := toComplete[:4]
	monthStr := toComplete[5:7]
	year, _ := strconv.Atoi(yearStr)
	month, _ := strconv.Atoi(monthStr)

	minDay := getMinDay(year, month, ctx)
	maxDay := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC).Day()

	suggestions := suggestDays(prefix, minDay, maxDay)
	return suggestions, cobra.ShellCompDirectiveNoSpace
}

// getMinDay calculates minimum valid day
func getMinDay(year, month int, ctx dateCompletionContext) int {
	minDay := 1

	if !ctx.isEndFlag && year == ctx.now.Year() && month == int(ctx.now.Month()) {
		minDay = ctx.now.Day()
	} else if ctx.isEndFlag && !ctx.startDate.IsZero() &&
		year == ctx.startDate.Year() && month == int(ctx.startDate.Month()) {
		minDay = ctx.startDate.Day() + 1
	}

	return minDay
}

// suggestDays generates day suggestions
func suggestDays(prefix string, minDay, maxDay int) []string {
	suggestions := []string{}
	commonDays := []int{1, 2, 3, 5, 10, 15, 20, 25, 28}

	for _, day := range commonDays {
		if day >= minDay && day <= maxDay {
			suggestions = append(suggestions, prefix+padZero(day)+"T")
		}
	}

	// Always include last day of month if valid
	if maxDay >= minDay && maxDay > 28 {
		suggestions = append(suggestions, prefix+padZero(maxDay)+"T")
	}

	// If no common days, suggest next valid days
	if len(suggestions) == 0 {
		for day := minDay; day <= maxDay && len(suggestions) < 10; day++ {
			suggestions = append(suggestions, prefix+padZero(day)+"T")
		}
	}

	return suggestions
}

// completeHour suggests hour values (YYYY-MM-DDTHH:)
func completeHour(toComplete string) ([]string, cobra.ShellCompDirective) {
	prefix := toComplete[:11]
	return []string{
		prefix + "00:", prefix + "06:", prefix + "08:",
		prefix + "09:", prefix + "10:", prefix + "12:",
		prefix + "14:", prefix + "16:", prefix + "18:",
		prefix + "20:", prefix + "22:", prefix + "23:",
	}, cobra.ShellCompDirectiveNoSpace
}

// completeMinute suggests minute values (YYYY-MM-DDTHH:MM:)
func completeMinute(toComplete string) ([]string, cobra.ShellCompDirective) {
	prefix := toComplete[:14]
	return []string{
		prefix + "00:", prefix + "15:", prefix + "30:", prefix + "45:",
	}, cobra.ShellCompDirectiveNoSpace
}

// completeSecondAndTimezone suggests second and timezone values
func completeSecondAndTimezone(toComplete string) ([]string, cobra.ShellCompDirective) {
	prefix := toComplete[:17]
	return []string{
		prefix + "00Z",      // UTC
		prefix + "00+00:00", // UTC explicit
		prefix + "00+07:00", // GMT+7 (Jakarta, Bangkok)
		prefix + "00+08:00", // GMT+8 (Singapore, Manila)
		prefix + "00+09:00", // GMT+9 (Tokyo, Seoul)
		prefix + "00-05:00", // GMT-5 (US EST)
		prefix + "00-08:00", // GMT-8 (US PST)
	}, cobra.ShellCompDirectiveNoFileComp
}

// completeFallback provides full example dates
func completeFallback(ctx dateCompletionContext) ([]string, cobra.ShellCompDirective) {
	suggestions := []string{}

	if ctx.isEndFlag && !ctx.startDate.IsZero() {
		// Suggest dates 1, 2, 3 days after start
		for i := 1; i <= 3; i++ {
			endDate := ctx.startDate.AddDate(0, 0, i)
			suggestions = append(suggestions, endDate.Format(time.RFC3339))
		}
	} else {
		// Suggest dates from today
		suggestions = []string{
			ctx.now.Format(time.RFC3339),
			ctx.now.AddDate(0, 0, 7).Format(time.RFC3339),
			ctx.now.AddDate(0, 1, 0).Format(time.RFC3339),
		}
	}

	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

// dateCompletion provides intelligent autocomplete for RFC3339 date format
func dateCompletion(cmd *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	length := len(toComplete)
	ctx := getDateCompletionContext(cmd)

	// Step 1: Year (YYYY-)
	if length == 0 {
		return completeYear(ctx)
	}

	// Step 2: Month (YYYY-MM-)
	if length == 5 && strings.Count(toComplete, "-") == 1 {
		return completeMonth(toComplete, ctx)
	}

	// Step 3: Day (YYYY-MM-DDT)
	if length == 8 && strings.Count(toComplete, "-") == 2 {
		return completeDay(toComplete, ctx)
	}

	// Step 4: Hour (YYYY-MM-DDTHH:)
	if length == 11 && strings.Contains(toComplete, "T") && !strings.Contains(toComplete, ":") {
		return completeHour(toComplete)
	}

	// Step 5: Minute (YYYY-MM-DDTHH:MM:)
	if length == 14 && strings.Count(toComplete, ":") == 1 {
		return completeMinute(toComplete)
	}

	// Step 6: Second + Timezone (YYYY-MM-DDTHH:MM:SS+/-HH:MM or Z)
	if length == 17 && strings.Count(toComplete, ":") == 2 {
		return completeSecondAndTimezone(toComplete)
	}

	// Fallback: provide full example dates based on context
	return completeFallback(ctx)
}

// padZero adds leading zero to single digit numbers
func padZero(n int) string {
	if n < 10 {
		return "0" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}

// eventNameCompletion provides autocomplete suggestions for event names and flags
func eventNameCompletion(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	// If event name is already provided, suggest flags
	if len(args) >= 1 {
		// Return flag suggestions
		flags := []string{
			"--title\tEvent title (required)",
			"--start\tStart date in RFC3339 format (required)",
			"--end\tEnd date in RFC3339 format (required)",
		}
		return flags, cobra.ShellCompDirectiveNoFileComp
	}

	now := time.Now()
	currentYear := now.Year()

	// Suggest common event naming patterns
	suggestions := []string{
		"ctf" + strconv.Itoa(currentYear) + "\tCTF " + strconv.Itoa(currentYear),
		"ctf" + strconv.Itoa(currentYear+1) + "\tCTF " + strconv.Itoa(currentYear+1),
		strings.ToLower(now.Month().String()) + "ctf" + strconv.Itoa(currentYear) + "\t" + now.Month().String() + " CTF " + strconv.Itoa(currentYear),
		"winterctf" + strconv.Itoa(currentYear) + "\tWinter CTF " + strconv.Itoa(currentYear),
		"summerctf" + strconv.Itoa(currentYear) + "\tSummer CTF " + strconv.Itoa(currentYear),
		"springctf" + strconv.Itoa(currentYear+1) + "\tSpring CTF " + strconv.Itoa(currentYear+1),
		"practice\tPractice environment",
		"training\tTraining environment",
	}

	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

func init() {
	rootCmd.AddCommand(eventCmd)
	eventCmd.AddCommand(eventListCmd)
	eventCmd.AddCommand(eventCurrentCmd)
	eventCmd.AddCommand(eventSwitchCmd)
	eventCmd.AddCommand(eventCreateCmd)

	// Add flags for event create command
	eventCreateCmd.Flags().StringVar(&eventCreateTitle, "title", "", "Event title (required)")
	eventCreateCmd.Flags().StringVar(&eventCreateStart, "start", "", "Start date in RFC3339 format, e.g., 2024-12-01T00:00:00Z (required)")
	eventCreateCmd.Flags().StringVar(&eventCreateEnd, "end", "", "End date in RFC3339 format, e.g., 2024-12-03T00:00:00Z (required)")

	// Add intelligent shell completion for date flags
	_ = eventCreateCmd.RegisterFlagCompletionFunc("start", dateCompletion)
	_ = eventCreateCmd.RegisterFlagCompletionFunc("end", dateCompletion)
}
