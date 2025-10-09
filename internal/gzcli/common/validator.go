package common

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/errors"
)

// Validator provides validation utilities
type Validator struct{}

// NewValidator creates a new Validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateChallenge validates a challenge configuration
func (v *Validator) ValidateChallenge(challenge config.ChallengeYaml) error {
	var validationErrors []string

	// Validate required fields
	if challenge.Name == "" {
		validationErrors = append(validationErrors, "name is required")
	}

	if challenge.Author == "" {
		validationErrors = append(validationErrors, "author is required")
	}

	if challenge.Type == "" {
		validationErrors = append(validationErrors, "type is required")
	}

	// Validate challenge type
	validTypes := map[string]struct{}{
		"StaticAttachment":  {},
		"StaticContainer":   {},
		"DynamicAttachment": {},
		"DynamicContainer":  {},
	}

	if _, valid := validTypes[challenge.Type]; !valid {
		validationErrors = append(validationErrors, fmt.Sprintf("invalid type: %s", challenge.Type))
	}

	// Validate value
	if challenge.Value < 0 {
		validationErrors = append(validationErrors, "value must be non-negative")
	}

	// Validate flags based on type
	if err := v.validateFlags(challenge); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate container configuration
	if err := v.validateContainer(challenge); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate hints
	if err := v.validateHints(challenge.Hints); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate scripts
	if err := v.validateScripts(challenge.Scripts); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	if len(validationErrors) > 0 {
		return errors.Wrapf(errors.ErrValidationFailed, "validation failed: %s", strings.Join(validationErrors, "; "))
	}

	return nil
}

// validateFlags validates flags based on challenge type
func (v *Validator) validateFlags(challenge config.ChallengeYaml) error {
	switch challenge.Type {
	case "StaticAttachment", "StaticContainer":
		if len(challenge.Flags) == 0 {
			return fmt.Errorf("flags are required for %s challenges", challenge.Type)
		}
		
		// Validate flag format
		for i, flag := range challenge.Flags {
			if !v.isValidFlagFormat(flag) {
				return fmt.Errorf("invalid flag format at index %d: %s", i, flag)
			}
		}
		
	case "DynamicContainer":
		if challenge.Container.FlagTemplate == "" {
			return fmt.Errorf("flag template is required for %s challenges", challenge.Type)
		}
	}

	return nil
}

// validateContainer validates container configuration
func (v *Validator) validateContainer(challenge config.ChallengeYaml) error {
	if challenge.Type == "DynamicContainer" || challenge.Type == "StaticContainer" {
		if challenge.Container.Image == "" {
			return fmt.Errorf("container image is required for %s challenges", challenge.Type)
		}
		
		// Validate memory limit
		if challenge.Container.MemoryLimit > 0 && challenge.Container.MemoryLimit < 64 {
			return fmt.Errorf("memory limit must be at least 64MB")
		}
		
		// Validate CPU count
		if challenge.Container.CPUCount > 0 && challenge.Container.CPUCount > 32 {
			return fmt.Errorf("CPU count cannot exceed 32")
		}
	}

	return nil
}

// validateHints validates hints
func (v *Validator) validateHints(hints []string) error {
	for i, hint := range hints {
		if strings.TrimSpace(hint) == "" {
			return fmt.Errorf("hint at index %d cannot be empty", i)
		}
		
		if len(hint) > 1000 {
			return fmt.Errorf("hint at index %d is too long (max 1000 characters)", i)
		}
	}

	return nil
}

// validateScripts validates scripts
func (v *Validator) validateScripts(scripts map[string]config.ScriptConfig) error {
	for name, script := range scripts {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("script name cannot be empty")
		}
		
		if script.Interval > 0 {
			if script.Interval < 30*time.Second {
				return fmt.Errorf("script %s interval must be at least 30 seconds", name)
			}
			if script.Interval > 24*time.Hour {
				return fmt.Errorf("script %s interval cannot exceed 24 hours", name)
			}
		}
	}

	return nil
}

// isValidFlagFormat checks if a flag has a valid format
func (v *Validator) isValidFlagFormat(flag string) bool {
	// Basic flag format validation - adjust regex as needed
	flagRegex := regexp.MustCompile(`^[A-Za-z0-9_{}]+$`)
	return flagRegex.MatchString(flag) && len(flag) >= 5
}

// ValidateConfig validates application configuration
func (v *Validator) ValidateConfig(config *config.Config) error {
	var validationErrors []string

	if config.Url == "" {
		validationErrors = append(validationErrors, "URL is required")
	}

	if config.Creds.Username == "" {
		validationErrors = append(validationErrors, "username is required")
	}

	if config.Creds.Password == "" {
		validationErrors = append(validationErrors, "password is required")
	}

	if config.Event.Title == "" {
		validationErrors = append(validationErrors, "event title is required")
	}

	if len(validationErrors) > 0 {
		return errors.Wrapf(errors.ErrInvalidConfig, "config validation failed: %s", strings.Join(validationErrors, "; "))
	}

	return nil
}