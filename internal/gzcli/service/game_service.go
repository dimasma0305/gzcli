package service

import (
	"fmt"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/repository"
	"github.com/dimasma0305/gzcli/internal/log"
)

// gameService implements GameService
type gameService struct {
	gameRepo      repository.GameRepository
	cache         repository.CacheRepository
	errorHandler  *ErrorHandler
	validator     *Validator
}

// NewGameService creates a new game service
func NewGameService(gameRepo repository.GameRepository, cache repository.CacheRepository) GameService {
	return &gameService{
		gameRepo:     gameRepo,
		cache:        cache,
		errorHandler: NewErrorHandler(),
		validator:    NewValidator(),
	}
}

// FindGame finds a game by title
func (s *gameService) FindGame(games []*gzapi.Game, title string) *gzapi.Game {
	for _, game := range games {
		if game.Title == title {
			return game
		}
	}
	return nil
}

// CreateGame creates a new game
func (s *gameService) CreateGame(conf *config.Config) (*gzapi.Game, error) {
	log.Info("Create new game")
	
	event := gzapi.CreateGameForm{
		Title: conf.Event.Title,
		Start: conf.Event.Start.Time,
		End:   conf.Event.End.Time,
	}
	
	game, err := s.gameRepo.Create(event)
	if err != nil {
		return nil, s.errorHandler.Wrap(err, "creating game")
	}
	
	if conf.Event.Poster == "" {
		return nil, fmt.Errorf("poster is required")
	}

	// Update the config with the new game data
	conf.Event.Id = game.Id
	conf.Event.PublicKey = game.PublicKey
	
	// Update the game with the full configuration
	updatedGame, err := s.gameRepo.Update(&conf.Event)
	if err != nil {
		return nil, s.errorHandler.Wrap(err, "updating game")
	}
	
	// Cache the updated config
	if err := s.cache.Set("config", conf); err != nil {
		log.Error("Failed to cache config: %v", err)
		// Don't return error as this is not critical
	}
	
	return updatedGame, nil
}

// UpdateGameIfNeeded updates a game if needed
func (s *gameService) UpdateGameIfNeeded(conf *config.Config, currentGame *gzapi.Game, api *gzapi.GZAPI, createPosterFunc func(string, *gzapi.Game, *gzapi.GZAPI) (string, error), setCache func(string, interface{}) error) error {
	log.Info("Updating game configuration...")
	
	// This would need to be implemented based on the existing updateGameIfNeeded logic
	// For now, just return nil as a placeholder
	return nil
}

// List returns all games
func (s *gameService) List() ([]*gzapi.Game, error) {
	return s.gameRepo.List()
}