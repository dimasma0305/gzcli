// Package structure provides utilities for generating challenge directory structures.
//
// This package helps maintain consistent directory layouts across challenges by
// copying template structures from a .structure directory to challenge directories.
//
// Example usage:
//
//	challenges := []ChallengeData{
//	    &Challenge{cwd: "./challenges/web/xss"},
//	    &Challenge{cwd: "./challenges/crypto/rsa"},
//	}
//
//	if err := structure.GenerateStructure(challenges); err != nil {
//	    log.Fatalf("Failed to generate structures: %v", err)
//	}
package structure

import (
	"fmt"
	"os"

	"github.com/dimasma0305/gzcli/internal/log"
	"github.com/dimasma0305/gzcli/internal/template"
)

// ChallengeData interface for accessing challenge data needed for structure generation
type ChallengeData interface {
	GetCwd() string
}

// GenerateStructure generates challenge structure from template
func GenerateStructure(challenges []ChallengeData) error {
	// Validate input
	if len(challenges) == 0 {
		return fmt.Errorf("no challenges provided")
	}

	// Read the .structure file
	_, err := os.ReadDir(".structure")
	if err != nil {
		return fmt.Errorf(".structure dir doesn't exist: %w", err)
	}

	// Iterate over each challenge in the challenges slice
	for _, challenge := range challenges {
		if challenge == nil {
			log.Error("Nil challenge encountered, skipping")
			continue
		}

		cwd := challenge.GetCwd()
		if cwd == "" {
			log.Error("Challenge has empty working directory, skipping")
			continue
		}

		// Construct the challenge path using the challenge data
		if err := template.TemplateToDestination(".structure", challenge, cwd); err != nil {
			log.Error("Failed to copy .structure to %s: %v", cwd, err)
			continue
		}
		log.Info("Successfully copied .structure to %s", cwd)
	}

	return nil
}
