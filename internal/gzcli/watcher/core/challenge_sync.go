package core

import (
	"context"
	"fmt"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/container"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/service"
	"github.com/dimasma0305/gzcli/internal/log"
)

// ChallengeSync handles challenge synchronization operations
type ChallengeSync struct {
	eventName string
	container *container.Container
}

// NewChallengeSync creates a new challenge sync handler
func NewChallengeSync(eventName string, container *container.Container) *ChallengeSync {
	return &ChallengeSync{
		eventName: eventName,
		container: container,
	}
}

// SyncChallenge syncs a challenge configuration with the API
func (cs *ChallengeSync) SyncChallenge(ctx context.Context, challengeConf config.ChallengeYaml, challenges []gzapi.Challenge) error {
	log.InfoH3("[%s] Syncing challenge: %s", cs.eventName, challengeConf.Name)

	// Get challenge service from container
	challengeSvc := cs.container.ChallengeService()

	// Check if challenge exists
	existingChallenge := cs.findExistingChallenge(challenges, challengeConf.Name)
	
	if existingChallenge == nil {
		// Create new challenge
		return cs.createNewChallenge(ctx, challengeConf, challengeSvc)
	}

	// Update existing challenge
	return cs.updateExistingChallenge(ctx, challengeConf, existingChallenge, challengeSvc)
}

// SyncToExistingChallenge syncs changes to an existing challenge
func (cs *ChallengeSync) SyncToExistingChallenge(ctx context.Context, challengeConf config.ChallengeYaml, existingChallenge *gzapi.Challenge, challenges []gzapi.Challenge) error {
	log.InfoH3("[%s] Syncing to existing challenge: %s", cs.eventName, challengeConf.Name)

	// Set the existing challenge data
	existingChallenge.CS = cs.container.API

	// Use service layer to sync challenge with existing data
	challengeSvc := cs.container.ChallengeService()
	return challengeSvc.Sync(ctx, challengeConf)
}

// findExistingChallenge finds an existing challenge by name
func (cs *ChallengeSync) findExistingChallenge(challenges []gzapi.Challenge, name string) *gzapi.Challenge {
	for _, challenge := range challenges {
		if challenge.Title == name {
			return &challenge
		}
	}
	return nil
}

// createNewChallenge creates a new challenge
func (cs *ChallengeSync) createNewChallenge(ctx context.Context, challengeConf config.ChallengeYaml, challengeSvc *service.ChallengeService) error {
	log.InfoH3("[%s] Creating new challenge: %s", cs.eventName, challengeConf.Name)
	
	if err := challengeSvc.Sync(ctx, challengeConf); err != nil {
		return fmt.Errorf("failed to create challenge %s: %w", challengeConf.Name, err)
	}

	log.InfoH3("[%s] Successfully created challenge: %s", cs.eventName, challengeConf.Name)
	return nil
}

// updateExistingChallenge updates an existing challenge
func (cs *ChallengeSync) updateExistingChallenge(ctx context.Context, challengeConf config.ChallengeYaml, existingChallenge *gzapi.Challenge, challengeSvc *service.ChallengeService) error {
	log.InfoH3("[%s] Updating existing challenge: %s", cs.eventName, challengeConf.Name)
	
	if err := challengeSvc.Sync(ctx, challengeConf); err != nil {
		return fmt.Errorf("failed to update challenge %s: %w", challengeConf.Name, err)
	}

	log.InfoH3("[%s] Successfully updated challenge: %s", cs.eventName, challengeConf.Name)
	return nil
}