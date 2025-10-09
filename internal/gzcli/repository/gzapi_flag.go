package repository

import (
	"fmt"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
)

// GZAPIFlagRepository implements FlagRepository using GZAPI
type GZAPIFlagRepository struct {
	api *gzapi.GZAPI
}

// NewGZAPIFlagRepository creates a new GZAPI flag repository
func NewGZAPIFlagRepository(api *gzapi.GZAPI) *GZAPIFlagRepository {
	return &GZAPIFlagRepository{
		api: api,
	}
}

// Create creates a new flag
func (r *GZAPIFlagRepository) Create(challengeID int, flag gzapi.CreateFlagForm) error {
	// This would need to be implemented in the gzapi package
	// For now, return an error as this method doesn't exist yet
	return fmt.Errorf("Create not implemented yet")
}

// Delete deletes a flag
func (r *GZAPIFlagRepository) Delete(challengeID int, flagID int) error {
	// This would need to be implemented in the gzapi package
	// For now, return an error as this method doesn't exist yet
	return fmt.Errorf("Delete not implemented yet")
}

// List returns all flags for a challenge
func (r *GZAPIFlagRepository) List(challengeID int) ([]gzapi.Flag, error) {
	// This would need to be implemented in the gzapi package
	// For now, return an error as this method doesn't exist yet
	return nil, fmt.Errorf("List not implemented yet")
}