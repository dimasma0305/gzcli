package repository

import (
	"fmt"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
)

// GZAPIGameRepository implements GameRepository using GZAPI
type GZAPIGameRepository struct {
	api *gzapi.GZAPI
}

// NewGZAPIGameRepository creates a new GZAPI game repository
func NewGZAPIGameRepository(api *gzapi.GZAPI) *GZAPIGameRepository {
	return &GZAPIGameRepository{
		api: api,
	}
}

// FindByTitle finds a game by title
func (r *GZAPIGameRepository) FindByTitle(title string) (*gzapi.Game, error) {
	// This would need to be implemented in the gzapi package
	// For now, return an error as this method doesn't exist yet
	return nil, fmt.Errorf("FindByTitle not implemented yet")
}

// Create creates a new game
func (r *GZAPIGameRepository) Create(game *gzapi.Game) error {
	// This would need to be implemented in the gzapi package
	// For now, return an error as this method doesn't exist yet
	return fmt.Errorf("Create not implemented yet")
}

// Update updates an existing game
func (r *GZAPIGameRepository) Update(game *gzapi.Game) error {
	// This would need to be implemented in the gzapi package
	// For now, return an error as this method doesn't exist yet
	return fmt.Errorf("Update not implemented yet")
}

// Delete deletes a game
func (r *GZAPIGameRepository) Delete(id int) error {
	// This would need to be implemented in the gzapi package
	// For now, return an error as this method doesn't exist yet
	return fmt.Errorf("Delete not implemented yet")
}

// List returns all games
func (r *GZAPIGameRepository) List() ([]gzapi.Game, error) {
	// This would need to be implemented in the gzapi package
	// For now, return an error as this method doesn't exist yet
	return nil, fmt.Errorf("List not implemented yet")
}