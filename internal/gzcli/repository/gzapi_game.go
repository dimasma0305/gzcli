package repository

import (
	"context"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/errors"
	"github.com/dimasma0305/gzcli/internal/log"
)

// GZAPIGameRepository implements GameRepository using GZAPI
type GZAPIGameRepository struct {
	api *gzapi.GZAPI
}

// NewGZAPIGameRepository creates a new GZAPI game repository
func NewGZAPIGameRepository(api *gzapi.GZAPI) GameRepository {
	return &GZAPIGameRepository{
		api: api,
	}
}

// GetGames retrieves all games
func (r *GZAPIGameRepository) GetGames(ctx context.Context) ([]gzapi.Game, error) {
	games, err := r.api.GetGames()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get games")
	}

	return games, nil
}

// GetGame retrieves a specific game by ID
func (r *GZAPIGameRepository) GetGame(ctx context.Context, gameID int) (*gzapi.Game, error) {
	games, err := r.GetGames(ctx)
	if err != nil {
		return nil, err
	}

	for _, game := range games {
		if game.Id == gameID {
			return &game, nil
		}
	}

	return nil, errors.Wrapf(errors.ErrChallengeNotFound, "game %d not found", gameID)
}

// CreateGame creates a new game
func (r *GZAPIGameRepository) CreateGame(ctx context.Context, game *gzapi.Game) (*gzapi.Game, error) {
	createdGame, err := r.api.CreateGame(game)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create game: %s", game.Title)
	}

	log.Info("Successfully created game: %s (ID: %d)", createdGame.Title, createdGame.Id)
	return createdGame, nil
}

// UpdateGame updates an existing game
func (r *GZAPIGameRepository) UpdateGame(ctx context.Context, game *gzapi.Game) (*gzapi.Game, error) {
	updatedGame, err := r.api.UpdateGame(game)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to update game %d: %s", game.Id, game.Title)
	}

	log.Info("Successfully updated game: %s (ID: %d)", updatedGame.Title, updatedGame.Id)
	return updatedGame, nil
}

// DeleteGame deletes a game
func (r *GZAPIGameRepository) DeleteGame(ctx context.Context, gameID int) error {
	err := r.api.DeleteGame(gameID)
	if err != nil {
		return errors.Wrapf(err, "failed to delete game %d", gameID)
	}

	log.Info("Successfully deleted game %d", gameID)
	return nil
}