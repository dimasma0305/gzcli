package repository

import (
	"context"
	"fmt"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/errors"
)

// GZAPIChallengeRepository implements ChallengeRepository using GZAPI
type GZAPIChallengeRepository struct {
	api *gzapi.GZAPI
}

// NewGZAPIChallengeRepository creates a new GZAPI challenge repository
func NewGZAPIChallengeRepository(api *gzapi.GZAPI) ChallengeRepository {
	return &GZAPIChallengeRepository{
		api: api,
	}
}

// GetChallenges retrieves all challenges for a game
func (r *GZAPIChallengeRepository) GetChallenges(ctx context.Context, gameID int) ([]gzapi.Challenge, error) {
	// Create a temporary game object to use the existing API
	game := &gzapi.Game{Id: gameID, CS: r.api}
	challenges, err := game.GetChallenges()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get challenges for game %d", gameID)
	}
	return challenges, nil
}

// GetChallenge retrieves a specific challenge
func (r *GZAPIChallengeRepository) GetChallenge(ctx context.Context, gameID int, challengeID int) (*gzapi.Challenge, error) {
	challenges, err := r.GetChallenges(ctx, gameID)
	if err != nil {
		return nil, err
	}

	for _, challenge := range challenges {
		if challenge.Id == challengeID {
			return &challenge, nil
		}
	}

	return nil, errors.Wrapf(errors.ErrChallengeNotFound, "challenge %d not found in game %d", challengeID, gameID)
}

// CreateChallenge creates a new challenge
func (r *GZAPIChallengeRepository) CreateChallenge(ctx context.Context, gameID int, challenge *gzapi.Challenge) (*gzapi.Challenge, error) {
	// This would need to be implemented in the GZAPI package
	// For now, return an error indicating it's not implemented
	return nil, fmt.Errorf("create challenge not implemented in GZAPI")
}

// UpdateChallenge updates an existing challenge
func (r *GZAPIChallengeRepository) UpdateChallenge(ctx context.Context, gameID int, challenge *gzapi.Challenge) (*gzapi.Challenge, error) {
	// This would need to be implemented in the GZAPI package
	// For now, return an error indicating it's not implemented
	return nil, fmt.Errorf("update challenge not implemented in GZAPI")
}

// DeleteChallenge deletes a challenge
func (r *GZAPIChallengeRepository) DeleteChallenge(ctx context.Context, gameID int, challengeID int) error {
	// This would need to be implemented in the GZAPI package
	// For now, return an error indicating it's not implemented
	return fmt.Errorf("delete challenge not implemented in GZAPI")
}