package repository

import (
	"fmt"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
)

// GZAPIChallengeRepository implements ChallengeRepository using GZAPI
type GZAPIChallengeRepository struct {
	event *gzapi.Game
}

// NewGZAPIChallengeRepository creates a new GZAPI challenge repository
func NewGZAPIChallengeRepository(event *gzapi.Game) *GZAPIChallengeRepository {
	return &GZAPIChallengeRepository{
		event: event,
	}
}

// FindByName finds a challenge by its name
func (r *GZAPIChallengeRepository) FindByName(name string) (*gzapi.Challenge, error) {
	return r.event.GetChallenge(name)
}

// FindByID finds a challenge by its ID
func (r *GZAPIChallengeRepository) FindByID(id int) (*gzapi.Challenge, error) {
	challenges, err := r.event.GetChallenges()
	if err != nil {
		return nil, err
	}
	
	for _, challenge := range challenges {
		if challenge.Id == id {
			return &challenge, nil
		}
	}
	
	return nil, fmt.Errorf("challenge with ID %d not found", id)
}

// List returns all challenges
func (r *GZAPIChallengeRepository) List() ([]gzapi.Challenge, error) {
	return r.event.GetChallenges()
}

// Create creates a new challenge
func (r *GZAPIChallengeRepository) Create(challenge *gzapi.Challenge) (*gzapi.Challenge, error) {
	// This would need to be implemented based on the GZAPI create challenge functionality
	// For now, return an error as this might not be directly supported
	return nil, fmt.Errorf("create challenge not implemented")
}

// Update updates an existing challenge
func (r *GZAPIChallengeRepository) Update(challenge *gzapi.Challenge) (*gzapi.Challenge, error) {
	updated, err := challenge.Update(*challenge)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// Delete deletes a challenge
func (r *GZAPIChallengeRepository) Delete(id int) error {
	// This would need to be implemented based on the GZAPI delete challenge functionality
	return fmt.Errorf("delete challenge not implemented")
}

// Exists checks if a challenge exists by name
func (r *GZAPIChallengeRepository) Exists(name string) (bool, error) {
	challenges, err := r.List()
	if err != nil {
		return false, err
	}
	
	for _, challenge := range challenges {
		if challenge.Title == name {
			return true, nil
		}
	}
	
	return false, nil
}