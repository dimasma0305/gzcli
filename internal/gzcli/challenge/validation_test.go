package challenge

import (
	"strings"
	"testing"
	"time"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
)

func TestIsGoodChallenge_Valid(t *testing.T) {
	tests := []struct {
		name      string
		challenge config.ChallengeYaml
	}{
		{
			name: "valid static attachment challenge",
			challenge: config.ChallengeYaml{
				Name:        "Test Challenge",
				Author:      "test-author",
				Description: "Test description",
				Type:        "StaticAttachment",
				Value:       100,
				Flags:       []string{"FLAG{test}"},
			},
		},
		{
			name: "valid static container challenge",
			challenge: config.ChallengeYaml{
				Name:        "Container Challenge",
				Author:      "test-author",
				Description: "Container test",
				Type:        "StaticContainer",
				Value:       200,
				Flags:       []string{"FLAG{container_test}"},
			},
		},
		{
			name: "valid dynamic container challenge",
			challenge: config.ChallengeYaml{
				Name:        "Dynamic Challenge",
				Author:      "test-author",
				Description: "Dynamic test",
				Type:        "DynamicContainer",
				Value:       500,
				Container: config.Container{
					FlagTemplate: "FLAG{[TEAM_HASH]}",
				},
			},
		},
		{
			name: "valid dynamic attachment challenge",
			challenge: config.ChallengeYaml{
				Name:        "Dynamic Attachment",
				Author:      "test-author",
				Description: "Dynamic attachment test",
				Type:        "DynamicAttachment",
				Value:       300,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IsGoodChallenge(tt.challenge)
			if err != nil {
				t.Errorf("IsGoodChallenge() error = %v, want nil", err)
			}
		})
	}
}

func TestIsGoodChallenge_Invalid(t *testing.T) {
	tests := []struct {
		name          string
		challenge     config.ChallengeYaml
		expectedError string
	}{
		{
			name: "missing name",
			challenge: config.ChallengeYaml{
				Author:      "test-author",
				Description: "Test",
				Type:        "StaticAttachment",
				Value:       100,
				Flags:       []string{"FLAG{test}"},
			},
			expectedError: "missing name",
		},
		{
			name: "missing author",
			challenge: config.ChallengeYaml{
				Name:        "Test",
				Description: "Test",
				Type:        "StaticAttachment",
				Value:       100,
				Flags:       []string{"FLAG{test}"},
			},
			expectedError: "missing author",
		},
		{
			name: "invalid type",
			challenge: config.ChallengeYaml{
				Name:        "Test",
				Author:      "test-author",
				Description: "Test",
				Type:        "InvalidType",
				Value:       100,
				Flags:       []string{"FLAG{test}"},
			},
			expectedError: "invalid type",
		},
		{
			name: "negative value",
			challenge: config.ChallengeYaml{
				Name:        "Test",
				Author:      "test-author",
				Description: "Test",
				Type:        "StaticAttachment",
				Value:       -100,
				Flags:       []string{"FLAG{test}"},
			},
			expectedError: "negative value",
		},
		{
			name: "static attachment missing flags",
			challenge: config.ChallengeYaml{
				Name:        "Test",
				Author:      "test-author",
				Description: "Test",
				Type:        "StaticAttachment",
				Value:       100,
				Flags:       []string{},
			},
			expectedError: "missing flags",
		},
		{
			name: "static container missing flags",
			challenge: config.ChallengeYaml{
				Name:        "Test",
				Author:      "test-author",
				Description: "Test",
				Type:        "StaticContainer",
				Value:       100,
			},
			expectedError: "missing flags",
		},
		{
			name: "dynamic container missing flag template",
			challenge: config.ChallengeYaml{
				Name:        "Test",
				Author:      "test-author",
				Description: "Test",
				Type:        "DynamicContainer",
				Value:       100,
				Container:   config.Container{},
			},
			expectedError: "missing flag template",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IsGoodChallenge(tt.challenge)
			if err == nil {
				t.Errorf("IsGoodChallenge() expected error containing %q, got nil", tt.expectedError)
				return
			}
			if !strings.Contains(err.Error(), "invalid challenge") {
				t.Errorf("IsGoodChallenge() error = %v, expected error containing 'invalid challenge'", err)
			}
		})
	}
}

func TestValidateChallenges_NoDuplicates(t *testing.T) {
	challenges := []config.ChallengeYaml{
		{
			Name:        "Challenge 1",
			Author:      "author1",
			Description: "desc1",
			Type:        "StaticAttachment",
			Value:       100,
			Flags:       []string{"FLAG{1}"},
			Cwd:         "/path/1",
		},
		{
			Name:        "Challenge 2",
			Author:      "author2",
			Description: "desc2",
			Type:        "StaticAttachment",
			Value:       200,
			Flags:       []string{"FLAG{2}"},
			Cwd:         "/path/2",
		},
	}

	err := ValidateChallenges(challenges)
	if err != nil {
		t.Errorf("ValidateChallenges() error = %v, want nil", err)
	}
}

func TestValidateChallenges_WithDuplicates(t *testing.T) {
	challenges := []config.ChallengeYaml{
		{
			Name:        "Duplicate Challenge",
			Author:      "author1",
			Description: "desc1",
			Type:        "StaticAttachment",
			Value:       100,
			Flags:       []string{"FLAG{1}"},
			Cwd:         "/path/1",
		},
		{
			Name:        "Duplicate Challenge",
			Author:      "author2",
			Description: "desc2",
			Type:        "StaticAttachment",
			Value:       200,
			Flags:       []string{"FLAG{2}"},
			Cwd:         "/path/2",
		},
	}

	err := ValidateChallenges(challenges)
	if err == nil {
		t.Error("ValidateChallenges() expected error for duplicate names, got nil")
		return
	}

	if !strings.Contains(err.Error(), "multiple challenges with the same name") {
		t.Errorf("ValidateChallenges() error = %v, expected error about duplicate names", err)
	}

	if !strings.Contains(err.Error(), "Duplicate Challenge") {
		t.Errorf("ValidateChallenges() error = %v, expected error to contain challenge name", err)
	}
}

func TestValidateChallenges_MultipleDuplicates(t *testing.T) {
	challenges := []config.ChallengeYaml{
		{
			Name:   "Challenge A",
			Author: "author1",
			Type:   "StaticAttachment",
			Value:  100,
			Flags:  []string{"FLAG{1}"},
			Cwd:    "/path/1",
		},
		{
			Name:   "Challenge A",
			Author: "author2",
			Type:   "StaticAttachment",
			Value:  200,
			Flags:  []string{"FLAG{2}"},
			Cwd:    "/path/2",
		},
		{
			Name:   "Challenge B",
			Author: "author3",
			Type:   "StaticAttachment",
			Value:  300,
			Flags:  []string{"FLAG{3}"},
			Cwd:    "/path/3",
		},
		{
			Name:   "Challenge B",
			Author: "author4",
			Type:   "StaticAttachment",
			Value:  400,
			Flags:  []string{"FLAG{4}"},
			Cwd:    "/path/4",
		},
	}

	err := ValidateChallenges(challenges)
	if err == nil {
		t.Error("ValidateChallenges() expected error for multiple duplicate names, got nil")
		return
	}

	errorMsg := err.Error()
	if !strings.Contains(errorMsg, "Challenge A") {
		t.Errorf("Expected error to contain 'Challenge A', got: %v", err)
	}
	if !strings.Contains(errorMsg, "Challenge B") {
		t.Errorf("Expected error to contain 'Challenge B', got: %v", err)
	}
}

func TestValidateChallenges_InvalidChallenge(t *testing.T) {
	challenges := []config.ChallengeYaml{
		{
			Name:        "Valid Challenge",
			Author:      "author1",
			Description: "desc1",
			Type:        "StaticAttachment",
			Value:       100,
			Flags:       []string{"FLAG{1}"},
			Cwd:         "/path/1",
		},
		{
			// Missing name - invalid
			Author:      "author2",
			Description: "desc2",
			Type:        "StaticAttachment",
			Value:       200,
			Flags:       []string{"FLAG{2}"},
			Cwd:         "/path/2",
		},
	}

	err := ValidateChallenges(challenges)
	if err == nil {
		t.Error("ValidateChallenges() expected error for invalid challenge, got nil")
	}
}

func TestValidateInterval(t *testing.T) {
	tests := []struct {
		name       string
		interval   time.Duration
		scriptName string
		want       bool
	}{
		{
			name:       "valid interval - 1 minute",
			interval:   1 * time.Minute,
			scriptName: "test_script",
			want:       true,
		},
		{
			name:       "valid interval - 1 hour",
			interval:   1 * time.Hour,
			scriptName: "test_script",
			want:       true,
		},
		{
			name:       "valid interval - 12 hours",
			interval:   12 * time.Hour,
			scriptName: "test_script",
			want:       true,
		},
		{
			name:       "minimum valid interval - 30 seconds",
			interval:   30 * time.Second,
			scriptName: "test_script",
			want:       true,
		},
		{
			name:       "maximum valid interval - 24 hours",
			interval:   24 * time.Hour,
			scriptName: "test_script",
			want:       true,
		},
		{
			name:       "too short - 10 seconds",
			interval:   10 * time.Second,
			scriptName: "test_script",
			want:       false,
		},
		{
			name:       "too short - 29 seconds",
			interval:   29 * time.Second,
			scriptName: "test_script",
			want:       false,
		},
		{
			name:       "too long - 25 hours",
			interval:   25 * time.Hour,
			scriptName: "test_script",
			want:       false,
		},
		{
			name:       "too long - 48 hours",
			interval:   48 * time.Hour,
			scriptName: "test_script",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateInterval(tt.interval, tt.scriptName)
			if got != tt.want {
				t.Errorf("ValidateInterval(%v, %q) = %v, want %v", tt.interval, tt.scriptName, got, tt.want)
			}
		})
	}
}

func TestValidateChallenges_EmptyList(t *testing.T) {
	challenges := []config.ChallengeYaml{}

	err := ValidateChallenges(challenges)
	if err != nil {
		t.Errorf("ValidateChallenges() with empty list error = %v, want nil", err)
	}
}

func TestIsGoodChallenge_AllTypes(t *testing.T) {
	types := []string{
		"StaticAttachment",
		"StaticContainer",
		"DynamicAttachment",
		"DynamicContainer",
	}

	for _, challengeType := range types {
		t.Run(challengeType, func(t *testing.T) {
			challenge := config.ChallengeYaml{
				Name:        "Test " + challengeType,
				Author:      "test-author",
				Description: "Test",
				Type:        challengeType,
				Value:       100,
			}

			// Add flags for static types
			if challengeType == "StaticAttachment" || challengeType == "StaticContainer" {
				challenge.Flags = []string{"FLAG{test}"}
			}

			// Add flag template for dynamic container
			if challengeType == "DynamicContainer" {
				challenge.Container.FlagTemplate = "FLAG{[TEAM_HASH]}"
			}

			err := IsGoodChallenge(challenge)
			if err != nil {
				t.Errorf("IsGoodChallenge() for type %s error = %v, want nil", challengeType, err)
			}
		})
	}
}

func TestIsGoodChallenge_ZeroValue(t *testing.T) {
	challenge := config.ChallengeYaml{
		Name:        "Zero Value Challenge",
		Author:      "test-author",
		Description: "Test with zero value",
		Type:        "StaticAttachment",
		Value:       0, // Zero is valid
		Flags:       []string{"FLAG{test}"},
	}

	err := IsGoodChallenge(challenge)
	if err != nil {
		t.Errorf("IsGoodChallenge() with zero value error = %v, want nil", err)
	}
}
