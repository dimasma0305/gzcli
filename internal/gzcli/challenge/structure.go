//nolint:revive // Exported functions follow project conventions
package challenge

import (
	"fmt"
	"os"

	"github.com/dimasma0305/gzcli/internal/log"
	"github.com/dimasma0305/gzcli/internal/template"
)

func GenStructure(challenges []ChallengeYaml) error {
	// Read the .structure file
	_, err := os.ReadDir(".structure")
	if err != nil {
		return fmt.Errorf(".structure dir doesn't exist: %w", err)
	}

	// Iterate over each challenge in the challenges slice
	for _, challenge := range challenges {
		// Construct the challenge path using the challenge data
		if err := template.TemplateToDestination(".structure", challenge, challenge.Cwd); err != nil {
			log.Error("Failed to copy .structure to %s: %v", challenge.Cwd, err)
			continue
		}
		log.Info("Successfully copied .structure to %s", challenge.Cwd)
	}

	return nil
}
