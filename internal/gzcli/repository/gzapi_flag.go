package repository

import (
	"fmt"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
)

// GZAPIFlagRepository implements FlagRepository using GZAPI
type GZAPIFlagRepository struct{}

// NewGZAPIFlagRepository creates a new GZAPI flag repository
func NewGZAPIFlagRepository() *GZAPIFlagRepository {
	return &GZAPIFlagRepository{}
}

// Create creates a new flag
func (r *GZAPIFlagRepository) Create(challengeID int, flag gzapi.CreateFlagForm) error {
	// This would need to be implemented based on the GZAPI create flag functionality
	// For now, return an error as this might not be directly supported
	return fmt.Errorf("create flag not implemented")
}

// Delete deletes a flag
func (r *GZAPIFlagRepository) Delete(flagID int) error {
	// This would need to be implemented based on the GZAPI delete flag functionality
	// For now, return an error as this might not be directly supported
	return fmt.Errorf("delete flag not implemented")
}

// List returns all flags for a challenge
func (r *GZAPIFlagRepository) List(challengeID int) ([]gzapi.Flag, error) {
	// This would need to be implemented based on the GZAPI list flags functionality
	return nil, fmt.Errorf("list flags not implemented")
}

// Exists checks if a flag exists
func (r *GZAPIFlagRepository) Exists(challengeID int, flag string) (bool, error) {
	flags, err := r.List(challengeID)
	if err != nil {
		return false, err
	}
	
	for _, f := range flags {
		if f.Flag == flag {
			return true, nil
		}
	}
	
	return false, nil
}