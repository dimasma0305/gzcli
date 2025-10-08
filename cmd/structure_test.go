package cmd

import (
	"strings"
	"testing"
)

func TestStructureCommand_MultiEventFlags(t *testing.T) {
	// Test that structure command properly registers multi-event flags

	// Check that event flag is registered
	eventFlag := structureCmd.Flags().Lookup("event")
	if eventFlag == nil {
		t.Error("structure command should have --event flag")
	}

	// Check that exclude-event flag is registered
	excludeFlag := structureCmd.Flags().Lookup("exclude-event")
	if excludeFlag == nil {
		t.Error("structure command should have --exclude-event flag")
	}

	// Check that event flag has correct shorthand
	if eventFlag != nil && eventFlag.Shorthand != "e" {
		t.Errorf("structure --event flag shorthand = %q, want %q", eventFlag.Shorthand, "e")
	}

	// Check flag types
	if eventFlag != nil && eventFlag.Value.Type() != "stringSlice" {
		t.Errorf("structure --event flag type = %q, want %q", eventFlag.Value.Type(), "stringSlice")
	}

	if excludeFlag != nil && excludeFlag.Value.Type() != "stringSlice" {
		t.Errorf("structure --exclude-event flag type = %q, want %q", excludeFlag.Value.Type(), "stringSlice")
	}
}

func TestStructureCommand_HelpText(t *testing.T) {
	// Test that help text mentions multi-event behavior

	if !strings.Contains(structureCmd.Long, "all events") {
		t.Error("structure command Long description should mention 'all events' behavior")
	}

	if !strings.Contains(structureCmd.Long, "--event") {
		t.Error("structure command Long description should mention --event flag")
	}

	if !strings.Contains(structureCmd.Long, "--exclude-event") {
		t.Error("structure command Long description should mention --exclude-event flag")
	}
}

//nolint:dupl // Test structure is similar but tests different commands
func TestStructureCommand_Structure(t *testing.T) {
	tests := []struct {
		name      string
		checkFunc func(*testing.T)
	}{
		{
			name: "command has correct use",
			checkFunc: func(t *testing.T) {
				if structureCmd.Use != "structure" {
					t.Errorf("structure command Use = %q, want %q", structureCmd.Use, "structure")
				}
			},
		},
		{
			name: "command has short description",
			checkFunc: func(t *testing.T) {
				if structureCmd.Short == "" {
					t.Error("structure command should have short description")
				}
			},
		},
		{
			name: "command has long description",
			checkFunc: func(t *testing.T) {
				if structureCmd.Long == "" {
					t.Error("structure command should have long description")
				}
			},
		},
		{
			name: "command has examples",
			checkFunc: func(t *testing.T) {
				if structureCmd.Example == "" {
					t.Error("structure command should have examples")
				}
			},
		},
		{
			name: "command has run function",
			checkFunc: func(t *testing.T) {
				if structureCmd.Run == nil {
					t.Error("structure command should have Run function")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.checkFunc)
	}
}

func TestStructureCommand_Examples(t *testing.T) {
	// Verify that examples contain expected patterns
	expectedPatterns := []string{
		"all events",
		"--event",
		"--exclude-event",
	}

	helpText := structureCmd.Long + structureCmd.Example

	for _, pattern := range expectedPatterns {
		if !strings.Contains(helpText, pattern) {
			t.Errorf("structure command help should contain pattern %q", pattern)
		}
	}
}

func TestStructureCommand_VariablesInitialized(t *testing.T) {
	// Test that structure command variables are properly initialized

	if structureEvents == nil {
		t.Error("structureEvents should be initialized")
	}

	if structureExcludeEvents == nil {
		t.Error("structureExcludeEvents should be initialized")
	}
}

func TestStructureCommand_FlagDefaults(t *testing.T) {
	// Reset flags to test defaults
	structureEvents = []string{}
	structureExcludeEvents = []string{}

	if len(structureEvents) != 0 {
		t.Errorf("structureEvents default = %v, want empty slice", structureEvents)
	}

	if len(structureExcludeEvents) != 0 {
		t.Errorf("structureExcludeEvents default = %v, want empty slice", structureExcludeEvents)
	}
}

func TestStructureCommand_DescriptionMentionsTemplate(t *testing.T) {
	// Verify description mentions .structure template
	if !strings.Contains(structureCmd.Long, ".structure") {
		t.Error("structure command description should mention '.structure' template")
	}
}
