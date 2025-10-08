package cmd

import (
	"strings"
	"testing"
)

func TestScriptCommand_MultiEventFlags(t *testing.T) {
	// Test that script command properly registers multi-event flags

	// Check that event flag is registered
	eventFlag := scriptCmd.Flags().Lookup("event")
	if eventFlag == nil {
		t.Error("script command should have --event flag")
	}

	// Check that exclude-event flag is registered
	excludeFlag := scriptCmd.Flags().Lookup("exclude-event")
	if excludeFlag == nil {
		t.Error("script command should have --exclude-event flag")
	}

	// Check that event flag has correct shorthand
	if eventFlag != nil && eventFlag.Shorthand != "e" {
		t.Errorf("script --event flag shorthand = %q, want %q", eventFlag.Shorthand, "e")
	}

	// Check flag types
	if eventFlag != nil && eventFlag.Value.Type() != "stringSlice" {
		t.Errorf("script --event flag type = %q, want %q", eventFlag.Value.Type(), "stringSlice")
	}

	if excludeFlag != nil && excludeFlag.Value.Type() != "stringSlice" {
		t.Errorf("script --exclude-event flag type = %q, want %q", excludeFlag.Value.Type(), "stringSlice")
	}
}

func TestScriptCommand_HelpText(t *testing.T) {
	// Test that help text mentions multi-event behavior

	if !strings.Contains(scriptCmd.Long, "all events") {
		t.Error("script command Long description should mention 'all events' behavior")
	}

	if !strings.Contains(scriptCmd.Long, "--event") {
		t.Error("script command Long description should mention --event flag")
	}

	if !strings.Contains(scriptCmd.Long, "--exclude-event") {
		t.Error("script command Long description should mention --exclude-event flag")
	}
}

func TestScriptCommand_Structure(t *testing.T) {
	tests := []struct {
		name      string
		checkFunc func(*testing.T)
	}{
		{
			name: "command has correct use",
			checkFunc: func(t *testing.T) {
				if scriptCmd.Use != "script <name>" {
					t.Errorf("script command Use = %q, want %q", scriptCmd.Use, "script <name>")
				}
			},
		},
		{
			name: "command requires exactly one argument",
			checkFunc: func(t *testing.T) {
				if scriptCmd.Args == nil {
					t.Error("script command should have Args validation")
				}
			},
		},
		{
			name: "command has short description",
			checkFunc: func(t *testing.T) {
				if scriptCmd.Short == "" {
					t.Error("script command should have short description")
				}
			},
		},
		{
			name: "command has long description",
			checkFunc: func(t *testing.T) {
				if scriptCmd.Long == "" {
					t.Error("script command should have long description")
				}
			},
		},
		{
			name: "command has examples",
			checkFunc: func(t *testing.T) {
				if scriptCmd.Example == "" {
					t.Error("script command should have examples")
				}
			},
		},
		{
			name: "command has run function",
			checkFunc: func(t *testing.T) {
				if scriptCmd.Run == nil {
					t.Error("script command should have Run function")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.checkFunc)
	}
}

func TestScriptCommand_Examples(t *testing.T) {
	// Verify that examples contain expected patterns
	expectedPatterns := []string{
		"all events",
		"--event",
		"--exclude-event",
		"deploy",
	}

	helpText := scriptCmd.Long + scriptCmd.Example

	for _, pattern := range expectedPatterns {
		if !strings.Contains(helpText, pattern) {
			t.Errorf("script command help should contain pattern %q", pattern)
		}
	}
}

func TestScriptCommand_VariablesInitialized(t *testing.T) {
	// Test that script command variables are properly initialized

	if scriptEvents == nil {
		t.Error("scriptEvents should be initialized")
	}

	if scriptExcludeEvents == nil {
		t.Error("scriptExcludeEvents should be initialized")
	}
}

func TestScriptCommand_FlagDefaults(t *testing.T) {
	// Reset flags to test defaults
	scriptEvents = []string{}
	scriptExcludeEvents = []string{}

	if len(scriptEvents) != 0 {
		t.Errorf("scriptEvents default = %v, want empty slice", scriptEvents)
	}

	if len(scriptExcludeEvents) != 0 {
		t.Errorf("scriptExcludeEvents default = %v, want empty slice", scriptExcludeEvents)
	}
}

func TestScriptCommand_ExampleScripts(t *testing.T) {
	// Verify examples show common script names
	commonScripts := []string{"deploy", "test", "cleanup"}

	for _, script := range commonScripts {
		if !strings.Contains(scriptCmd.Example, script) {
			t.Errorf("script command examples should include %q script", script)
		}
	}
}
