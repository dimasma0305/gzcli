package challenge

import (
	"fmt"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/log"
)

// FindCurrentGame searches for a game by its title in a slice of games and returns it.
// If found, it also associates the API client with the game object.
func FindCurrentGame(games []*gzapi.Game, title string, api *gzapi.GZAPI) *gzapi.Game {
	for _, game := range games {
		if game.Title == title {
			game.CS = api
			return game
		}
	}
	return nil
}

// CreateNewGame handles the creation of a new game on the server. It uses the local configuration
// to set up the game, uploads a poster, and updates the cache with the new game information.
func CreateNewGame(conf *config.Config, api *gzapi.GZAPI, createPosterFunc func(string, *gzapi.Game, *gzapi.GZAPI) (string, error), setCache func(string, interface{}) error) (*gzapi.Game, error) {
	log.Info("Create new game")
	event := gzapi.CreateGameForm{
		Title: conf.Event.Title,
		Start: conf.Event.Start.Time,
		End:   conf.Event.End.Time,
	}
	game, err := api.CreateGame(event)
	if err != nil {
		return nil, err
	}
	if conf.Event.Poster == "" {
		return nil, fmt.Errorf("poster is required")
	}

	poster, err := createPosterFunc(conf.Event.Poster, game, api)
	if err != nil {
		return nil, err
	}

	conf.Event.Id = game.Id
	conf.Event.PublicKey = game.PublicKey
	conf.Event.Poster = poster
	if err := game.Update(&conf.Event); err != nil {
		return nil, err
	}
	// Use event-specific cache key
	cacheKey := fmt.Sprintf("config-%s", conf.EventName)
	if err := setCache(cacheKey, conf); err != nil {
		return nil, err
	}
	return game, nil
}

// UpdateGameIfNeeded compares the local game configuration with the current game data on the
// server and performs an update if there are any differences.
func UpdateGameIfNeeded(conf *config.Config, currentGame *gzapi.Game, api *gzapi.GZAPI, createPosterFunc func(string, *gzapi.Game, *gzapi.GZAPI) (string, error), setCache func(string, interface{}) error) error {
	poster, err := createPosterFunc(conf.Event.Poster, currentGame, api)
	if err != nil {
		return err
	}
	conf.Event.Poster = poster
	if fmt.Sprintf("%v", conf.Event) != fmt.Sprintf("%v", *currentGame) {
		log.Info("Updated %s game", conf.Event.Title)

		conf.Event.Id = currentGame.Id
		conf.Event.PublicKey = currentGame.PublicKey

		if err := currentGame.Update(&conf.Event); err != nil {
			return err
		}
		// Use event-specific cache key
		cacheKey := fmt.Sprintf("config-%s", conf.EventName)
		if err := setCache(cacheKey, conf); err != nil {
			return err
		}
	}
	return nil
}
