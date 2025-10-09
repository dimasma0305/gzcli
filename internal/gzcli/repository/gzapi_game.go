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

// FindByTitle finds a game by its title
func (r *GZAPIGameRepository) FindByTitle(title string) (*gzapi.Game, error) {
	games, err := r.List()
	if err != nil {
		return nil, err
	}
	
	for _, game := range games {
		if game.Title == title {
			return game, nil
		}
	}
	
	return nil, fmt.Errorf("game with title %s not found", title)
}

// FindByID finds a game by its ID
func (r *GZAPIGameRepository) FindByID(id int) (*gzapi.Game, error) {
	games, err := r.List()
	if err != nil {
		return nil, err
	}
	
	for _, game := range games {
		if game.Id == id {
			return game, nil
		}
	}
	
	return nil, fmt.Errorf("game with ID %d not found", id)
}

// List returns all games
func (r *GZAPIGameRepository) List() ([]*gzapi.Game, error) {
	return r.api.GetGames()
}

// Create creates a new game
func (r *GZAPIGameRepository) Create(game gzapi.CreateGameForm) (*gzapi.Game, error) {
	return r.api.CreateGame(game)
}

// Update updates an existing game
func (r *GZAPIGameRepository) Update(game *gzapi.Game) (*gzapi.Game, error) {
	err := game.Update(game)
	if err != nil {
		return nil, err
	}
	return game, nil
}

// Delete deletes a game
func (r *GZAPIGameRepository) Delete(id int) error {
	// This would need to be implemented based on the GZAPI delete game functionality
	return fmt.Errorf("delete game not implemented")
}