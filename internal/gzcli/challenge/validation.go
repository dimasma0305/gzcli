package challenge

import (
	"fmt"
	"strings"
	"time"

	"github.com/dimasma0305/gzcli/internal/log"
)

var validTypes = map[string]struct{}{
	"StaticAttachment":  {},
	"StaticContainer":   {},
	"DynamicAttachment": {},
	"DynamicContainer":  {},
}

// Interval validation constants
const (
	MinInterval = 30 * time.Second
	MaxInterval = 24 * time.Hour
)

func IsGoodChallenge(challenge ChallengeYaml) error {
	var errors []string

	if challenge.Name == "" {
		errors = append(errors, "missing name")
	}
	if challenge.Author == "" {
		errors = append(errors, "missing author")
	}
	if _, valid := validTypes[challenge.Type]; !valid {
		errors = append(errors, fmt.Sprintf("invalid type: %s", challenge.Type))
	}
	if challenge.Value < 0 {
		errors = append(errors, "negative value")
	}

	switch {
	case len(challenge.Flags) == 0 && (challenge.Type == "StaticAttachment" || challenge.Type == "StaticContainer"):
		errors = append(errors, "missing flags for static challenge")
	case challenge.Type == "DynamicContainer" && challenge.Container.FlagTemplate == "":
		errors = append(errors, "missing flag template for dynamic container")
	}

	if len(errors) > 0 {
		log.Error("Validation errors for %s:", challenge.Name)
		for _, e := range errors {
			log.Error("  - %s", e)
		}
		return fmt.Errorf("invalid challenge: %s", challenge.Name)
	}

	return nil
}

func ValidateChallenges(challengesConf []ChallengeYaml) error {
	// Track seen names and duplicate occurrences
	seenNames := make(map[string]int, len(challengesConf))
	var duplicates []string

	// First pass: count occurrences
	for _, challengeConf := range challengesConf {
		seenNames[challengeConf.Name]++
	}

	// Collect names with duplicates
	for name, count := range seenNames {
		if count > 1 {
			duplicates = append(duplicates, name)
		}
	}

	// Return all duplicates at once
	if len(duplicates) > 0 {
		return fmt.Errorf("multiple challenges with the same name found:\n  - %s",
			strings.Join(duplicates, "\n  - "))
	}

	// Existing validation logic
	for _, challengeConf := range challengesConf {
		if challengeConf.Type == "" {
			challengeConf.Type = "StaticAttachments"
		}
		log.Info("Validating %s challenge...", challengeConf.Cwd)
		if err := IsGoodChallenge(challengeConf); err != nil {
			return fmt.Errorf("invalid challenge %q: %w", challengeConf.Name, err)
		}
		log.Info("Challenge %s is valid.", challengeConf.Cwd)
	}

	return nil
}

// ValidateInterval validates that an interval is within acceptable bounds
func ValidateInterval(interval time.Duration, scriptName string) bool {
	if interval < MinInterval {
		log.Error("Interval %v too short for script '%s', minimum is %v", interval, scriptName, MinInterval)
		return false
	}
	if interval > MaxInterval {
		log.Error("Interval %v too long for script '%s', maximum is %v", interval, scriptName, MaxInterval)
		return false
	}
	return true
}
