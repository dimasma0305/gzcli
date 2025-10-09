package service

import (
	"fmt"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/repository"
	"github.com/dimasma0305/gzcli/internal/log"
)

// GameService handles game business logic
type GameService struct {
	gameRepo repository.GameRepository
	cacheRepo repository.CacheRepository
	api      *gzapi.GZAPI
}

// GameServiceConfig holds configuration for GameService
type GameServiceConfig struct {
	GameRepo repository.GameRepository
	Cache    repository.CacheRepository
	API      *gzapi.GZAPI
}

// NewGameService creates a new GameService
func NewGameService(config GameServiceConfig) *GameService {
	return &GameService{
		gameRepo:  config.GameRepo,
		cacheRepo: config.Cache,
		api:       config.API,
	}
}

// FindGame finds a game by title
func (s *GameService) FindGame(games []gzapi.Game, title string) *gzapi.Game {
	for _, game := range games {
		if game.Title == title {
			game.CS = s.api
			return &game
		}
	}
	return nil
}

// Create creates a new game
func (s *GameService) Create(game *gzapi.Game) error {
	return s.gameRepo.Create(game)
}

// Update updates an existing game
func (s *GameService) Update(game *gzapi.Game) error {
	return s.gameRepo.Update(game)
}

// UpdateGameIfNeeded updates a game if needed
func (s *GameService) UpdateGameIfNeeded(config *config.Config, currentGame *gzapi.Game, createPosterFunc func(string, *gzapi.Game, *gzapi.GZAPI) (string, error), setCache func(string, interface{}) error) error {
	poster, err := createPosterFunc(config.Event.Poster, currentGame, s.api)
	if err != nil {
		return err
	}
	config.Event.Poster = poster
	if fmt.Sprintf("%v", config.Event) != fmt.Sprintf("%v", *currentGame) {
		log.Info("Updated %s game", config.Event.Title)

		config.Event.Id = currentGame.Id
		config.Event.PublicKey = currentGame.PublicKey

		if err := currentGame.Update(&config.Event); err != nil {
			return err
		}
		if err := setCache("config", config); err != nil {
			return err
		}
	}
	return nil
}

// List returns all games
func (s *GameService) List() ([]gzapi.Game, error) {
	return s.gameRepo.List()
}