package service

import (
	"context"
	"fmt"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/repository"
	"github.com/dimasma0305/gzcli/internal/gzcli/errors"
	"github.com/dimasma0305/gzcli/internal/log"
)

// GameServiceConfig holds configuration for GameService
type GameServiceConfig struct {
	Cache repository.CacheRepository
	GameRepo repository.GameRepository
	API   *gzapi.GZAPI
}

// GameService handles game business logic
type GameService struct {
	cache    repository.CacheRepository
	gameRepo repository.GameRepository
	api      *gzapi.GZAPI
}

// NewGameService creates a new GameService
func NewGameService(cache repository.CacheRepository, gameRepo repository.GameRepository, api *gzapi.GZAPI) *GameService {
	return &GameService{
		cache:    cache,
		gameRepo: gameRepo,
		api:      api,
	}
}

// FindGame finds a game by title
func (s *GameService) FindGame(ctx context.Context, games []gzapi.Game, title string) *gzapi.Game {
	for _, game := range games {
		if game.Title == title {
			return &game
		}
	}
	return nil
}

// UpdateGameIfNeeded updates a game if needed
func (s *GameService) UpdateGameIfNeeded(ctx context.Context, conf *config.Config, currentGame *gzapi.Game, createPosterIfNotExistOrDifferent func(*gzapi.Game, *gzapi.GZAPI) error, setCache func(string, interface{}) error) error {
	log.Info("Checking if game needs updating...")

	// Check if game configuration has changed
	if s.gameNeedsUpdate(conf, currentGame) {
		log.Info("Game configuration has changed, updating...")

		// Update game via API
		updatedGame, err := s.updateGame(ctx, currentGame, conf)
		if err != nil {
			return errors.Wrap(err, "failed to update game")
		}

		// Handle poster creation if needed
		if createPosterIfNotExistOrDifferent != nil {
			if err := createPosterIfNotExistOrDifferent(updatedGame, s.api); err != nil {
				log.Warn("Failed to create/update poster: %v", err)
			}
		}

		// Update cache
		if setCache != nil {
			if err := setCache("config", conf); err != nil {
				log.Warn("Failed to update config cache: %v", err)
			}
		}

		log.Info("Game updated successfully")
	} else {
		log.Info("Game is up to date")
	}

	return nil
}

// gameNeedsUpdate checks if a game needs updating
func (s *GameService) gameNeedsUpdate(conf *config.Config, currentGame *gzapi.Game) bool {
	// Compare key fields that might change
	return conf.Event.Title != currentGame.Title ||
		conf.Event.Description != currentGame.Description ||
		conf.Event.Value != currentGame.Value
}

// updateGame updates a game via the API
func (s *GameService) updateGame(ctx context.Context, currentGame *gzapi.Game, conf *config.Config) (*gzapi.Game, error) {
	// This would implement the actual game update logic
	// For now, return the current game
	log.Debug("Game update logic not fully implemented")
	return currentGame, nil
}

// CreateGame creates a new game
func (s *GameService) CreateGame(ctx context.Context, game *gzapi.Game) (*gzapi.Game, error) {
	if s.gameRepo == nil {
		return nil, errors.New("game repository not available")
	}
	
	createdGame, err := s.gameRepo.CreateGame(ctx, game)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create game")
	}
	
	return createdGame, nil
}

// GetGames retrieves all games
func (s *GameService) GetGames(ctx context.Context) ([]gzapi.Game, error) {
	if s.gameRepo == nil {
		return []gzapi.Game{}, nil
	}
	
	games, err := s.gameRepo.GetGames(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get games")
	}
	
	return games, nil
}