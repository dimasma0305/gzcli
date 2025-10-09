package repository

import (
	"fmt"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
)

// GZAPIChallengeRepository implements ChallengeRepository using GZAPI
type GZAPIChallengeRepository struct {
	api *gzapi.GZAPI
}

// NewGZAPIChallengeRepository creates a new GZAPI challenge repository
func NewGZAPIChallengeRepository(api *gzapi.GZAPI) *GZAPIChallengeRepository {
	return &GZAPIChallengeRepository{
		api: api,
	}
}

// FindByID finds a challenge by ID
func (r *GZAPIChallengeRepository) FindByID(id int) (*gzapi.Challenge, error) {
	// This would need to be implemented in the gzapi package
	// For now, return an error as this method doesn't exist yet
	return nil, fmt.Errorf("FindByID not implemented yet")
}

// FindByName finds a challenge by name
func (r *GZAPIChallengeRepository) FindByName(name string) (*gzapi.Challenge, error) {
	// This would need to be implemented in the gzapi package
	// For now, return an error as this method doesn't exist yet
	return nil, fmt.Errorf("FindByName not implemented yet")
}

// Create creates a new challenge
func (r *GZAPIChallengeRepository) Create(challenge *gzapi.Challenge) error {
	// This would need to be implemented in the gzapi package
	// For now, return an error as this method doesn't exist yet
	return fmt.Errorf("Create not implemented yet")
}

// Update updates an existing challenge
func (r *GZAPIChallengeRepository) Update(challenge *gzapi.Challenge) error {
	// This would need to be implemented in the gzapi package
	// For now, return an error as this method doesn't exist yet
	return fmt.Errorf("Update not implemented yet")
}

// Delete deletes a challenge
func (r *GZAPIChallengeRepository) Delete(id int) error {
	// This would need to be implemented in the gzapi package
	// For now, return an error as this method doesn't exist yet
	return fmt.Errorf("Delete not implemented yet")
}

// List returns all challenges
func (r *GZAPIChallengeRepository) List() ([]gzapi.Challenge, error) {
	// This would need to be implemented in the gzapi package
	// For now, return an error as this method doesn't exist yet
	return nil, fmt.Errorf("List not implemented yet")
}