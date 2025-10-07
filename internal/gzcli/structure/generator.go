// Package structure provides utilities for generating challenge directory structures
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
	// Read the .structure file
	_, err := os.ReadDir(".structure")
	if err != nil {
		return fmt.Errorf(".structure dir doesn't exist: %w", err)
	}

	// Iterate over each challenge in the challenges slice
	for _, challenge := range challenges {
		// Construct the challenge path using the challenge data
		if err := template.TemplateToDestination(".structure", challenge, challenge.GetCwd()); err != nil {
			log.Error("Failed to copy .structure to %s: %v", challenge.GetCwd(), err)
			continue
		}
		log.Info("Successfully copied .structure to %s", challenge.GetCwd())
	}

	return nil
}
