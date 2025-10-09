package repository

import (
	"context"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/errors"
	"github.com/dimasma0305/gzcli/internal/log"
)

// GZAPIFlagRepository implements FlagRepository using GZAPI
type GZAPIFlagRepository struct {
	api *gzapi.GZAPI
}

// NewGZAPIFlagRepository creates a new GZAPI flag repository
func NewGZAPIFlagRepository(api *gzapi.GZAPI) FlagRepository {
	return &GZAPIFlagRepository{
		api: api,
	}
}

// CreateFlag creates a new flag for a challenge
func (r *GZAPIFlagRepository) CreateFlag(ctx context.Context, challengeID int, flag *gzapi.Flag) (*gzapi.Flag, error) {
	createdFlag, err := r.api.CreateFlag(challengeID, flag)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create flag for challenge %d", challengeID)
	}

	log.Info("Successfully created flag for challenge %d", challengeID)
	return createdFlag, nil
}

// UpdateFlag updates an existing flag
func (r *GZAPIFlagRepository) UpdateFlag(ctx context.Context, challengeID int, flag *gzapi.Flag) (*gzapi.Flag, error) {
	updatedFlag, err := r.api.UpdateFlag(challengeID, flag)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to update flag %d for challenge %d", flag.Id, challengeID)
	}

	log.Info("Successfully updated flag %d for challenge %d", flag.Id, challengeID)
	return updatedFlag, nil
}

// DeleteFlag deletes a flag
func (r *GZAPIFlagRepository) DeleteFlag(ctx context.Context, challengeID int, flagID int) error {
	err := r.api.DeleteFlag(challengeID, flagID)
	if err != nil {
		return errors.Wrapf(err, "failed to delete flag %d for challenge %d", flagID, challengeID)
	}

	log.Info("Successfully deleted flag %d for challenge %d", flagID, challengeID)
	return nil
}

// GetFlags retrieves all flags for a challenge
func (r *GZAPIFlagRepository) GetFlags(ctx context.Context, challengeID int) ([]gzapi.Flag, error) {
	flags, err := r.api.GetFlags(challengeID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get flags for challenge %d", challengeID)
	}

	return flags, nil
}