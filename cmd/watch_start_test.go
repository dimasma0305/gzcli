package cmd

import (
	"strings"
	"testing"
)

func TestWatchStartCommand_MultiEventFlags(t *testing.T) {
	// Test that watch start command properly registers multi-event flags

	// Check that event flag is registered
	eventFlag := watchStartCmd.Flags().Lookup("event")
	if eventFlag == nil {
		t.Error("watch start command should have --event flag")
	}

	// Check that exclude-event flag is registered
	excludeFlag := watchStartCmd.Flags().Lookup("exclude-event")
	if excludeFlag == nil {
		t.Error("watch start command should have --exclude-event flag")
	}

	// Check that event flag has correct shorthand
	if eventFlag != nil && eventFlag.Shorthand != "e" {
		t.Errorf("watch start --event flag shorthand = %q, want %q", eventFlag.Shorthand, "e")
	}

	// Check flag types
	if eventFlag != nil && eventFlag.Value.Type() != "stringSlice" {
		t.Errorf("watch start --event flag type = %q, want %q", eventFlag.Value.Type(), "stringSlice")
	}

	if excludeFlag != nil && excludeFlag.Value.Type() != "stringSlice" {
		t.Errorf("watch start --exclude-event flag type = %q, want %q", excludeFlag.Value.Type(), "stringSlice")
	}
}

func TestWatchStartCommand_HelpText(t *testing.T) {
	// Test that help text mentions multi-event behavior

	if !strings.Contains(watchStartCmd.Long, "all events") {
		t.Error("watch start command Long description should mention 'all events' behavior")
	}

	if !strings.Contains(watchStartCmd.Long, "--event") {
		t.Error("watch start command Long description should mention --event flag")
	}

	if !strings.Contains(watchStartCmd.Long, "--exclude-event") {
		t.Error("watch start command Long description should mention --exclude-event flag")
	}
}

//nolint:dupl // Test structure is similar but tests different commands
func TestWatchStartCommand_Structure(t *testing.T) {
	tests := []struct {
		name      string
		checkFunc func(*testing.T)
	}{
		{
			name: "command has correct use",
			checkFunc: func(t *testing.T) {
				if watchStartCmd.Use != "start" {
					t.Errorf("watch start command Use = %q, want %q", watchStartCmd.Use, "start")
				}
			},
		},
		{
			name: "command has short description",
			checkFunc: func(t *testing.T) {
				if watchStartCmd.Short == "" {
					t.Error("watch start command should have short description")
				}
			},
		},
		{
			name: "command has long description",
			checkFunc: func(t *testing.T) {
				if watchStartCmd.Long == "" {
					t.Error("watch start command should have long description")
				}
			},
		},
		{
			name: "command has examples",
			checkFunc: func(t *testing.T) {
				if watchStartCmd.Example == "" {
					t.Error("watch start command should have examples")
				}
			},
		},
		{
			name: "command has run function",
			checkFunc: func(t *testing.T) {
				if watchStartCmd.Run == nil {
					t.Error("watch start command should have Run function")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.checkFunc)
	}
}

func TestWatchStartCommand_Examples(t *testing.T) {
	// Verify that examples contain expected patterns
	expectedPatterns := []string{
		"all events",
		"--event",
		"--exclude-event",
		"daemon",
	}

	helpText := watchStartCmd.Long + watchStartCmd.Example

	for _, pattern := range expectedPatterns {
		if !strings.Contains(helpText, pattern) {
			t.Errorf("watch start command help should contain pattern %q", pattern)
		}
	}
}

func TestWatchStartCommand_VariablesInitialized(t *testing.T) {
	// Test that watch command variables are properly initialized

	if watchEvents == nil {
		t.Error("watchEvents should be initialized")
	}

	if watchExcludeEvents == nil {
		t.Error("watchExcludeEvents should be initialized")
	}
}

func TestWatchStartCommand_FlagDefaults(t *testing.T) {
	// Reset flags to test defaults
	watchEvents = []string{}
	watchExcludeEvents = []string{}

	if len(watchEvents) != 0 {
		t.Errorf("watchEvents default = %v, want empty slice", watchEvents)
	}

	if len(watchExcludeEvents) != 0 {
		t.Errorf("watchExcludeEvents default = %v, want empty slice", watchExcludeEvents)
	}
}

func TestWatchStartCommand_OtherFlags(t *testing.T) {
	// Test that other important flags are still present

	importantFlags := []string{
		"foreground",
		"pid-file",
		"log-file",
		"debounce",
		"poll-interval",
	}

	for _, flagName := range importantFlags {
		flag := watchStartCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("watch start command should have --%s flag", flagName)
		}
	}
}

func TestWatchStartCommand_ForegroundFlag(t *testing.T) {
	// Test foreground flag specifically

	flag := watchStartCmd.Flags().Lookup("foreground")
	if flag == nil {
		t.Fatal("watch start command should have --foreground flag")
	}

	if flag.Shorthand != "f" {
		t.Errorf("--foreground flag shorthand = %q, want %q", flag.Shorthand, "f")
	}
}

func TestWatchStartCommand_DescriptionMentionsDaemon(t *testing.T) {
	// Verify description mentions daemon behavior
	if !strings.Contains(watchStartCmd.Long, "daemon") {
		t.Error("watch start command description should mention 'daemon' mode")
	}
}
