package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/utils"
	"github.com/dimasma0305/gzcli/internal/log"
)

const (
	GZCTF_DIR   = ".gzctf"
	CONFIG_FILE = "conf.yaml"
)

// Config represents the application configuration
type Config struct {
	Url         string       `yaml:"url"`
	Creds       gzapi.Creds  `yaml:"creds"`
	Event       gzapi.Game   `yaml:"event"`
	Appsettings *AppSettings `yaml:"-"`
}

func GetConfig(api *gzapi.GZAPI, getCache func(string, interface{}) error, setCache func(string, interface{}) error, deleteCache func(string), createNewGame func(*Config, *gzapi.GZAPI) (*gzapi.Game, error)) (*Config, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	var config Config
	confPath := filepath.Join(dir, GZCTF_DIR, CONFIG_FILE)
	if err := utils.ParseYamlFromFile(confPath, &config); err != nil {
		return nil, err
	}

	// First, try to get cached config
	var configCache Config
	cacheErr := getCache("config", &configCache)

	// If we have cached game info, use it as the starting point
	if cacheErr == nil && configCache.Event.Id != 0 {
		config.Event.Id = configCache.Event.Id
		config.Event.PublicKey = configCache.Event.PublicKey
		log.Info("Using cached game ID: %d", config.Event.Id)
	}

	// Only interact with API if provided and we need to validate/create game
	if api != nil && api.Client != nil {
		// If we have a cached game ID, try to validate it exists
		if config.Event.Id != 0 {
			log.Info("Validating cached game ID %d exists on server...", config.Event.Id)
			games, err := api.GetGames()
			if err != nil {
				log.Error("Failed to get games for validation: %v", err)
				return nil, fmt.Errorf("API games fetch error: %w", err)
			}

			// Check if the cached game ID still exists
			gameExists := false
			for _, game := range games {
				if game.Id == config.Event.Id {
					gameExists = true
					// Update with current server data but keep the same ID
					config.Event.PublicKey = game.PublicKey
					log.Info("Cached game ID %d validated successfully", config.Event.Id)
					break
				}
			}

			// If cached game doesn't exist, clear cache and try to find by title
			if !gameExists {
				log.Info("Cached game ID %d not found on server, searching by title...", config.Event.Id)
				deleteCache("config")
				config.Event.Id = 0
				config.Event.PublicKey = ""
			}
		}

		// If we don't have a valid game ID, try to find by title or create new
		if config.Event.Id == 0 {
			game, err := api.GetGameByTitle(config.Event.Title)
			if err != nil {
				log.Info("Game '%s' not found by title, creating new game...", config.Event.Title)
				_, err = createNewGame(&config, api)
				if err != nil {
					return nil, fmt.Errorf("failed to create new game: %w", err)
				}
			} else {
				log.Info("Found existing game by title: %s (ID: %d)", game.Title, game.Id)
				config.Event.Id = game.Id
				config.Event.PublicKey = game.PublicKey
				// Update cache with found game
				if err := setCache("config", &config); err != nil {
					log.Error("Failed to update cache with found game: %v", err)
				}
			}
		}
	}

	config.Appsettings, err = GetAppSettings()
	if err != nil {
		return nil, fmt.Errorf("errror parsing appsettings.json: %s", err)
	}

	// Ensure the GZAPI client is set if provided to prevent nil pointer dereference
	if api != nil {
		config.Event.CS = api
	}

	return &config, nil
}

// GetAppSettingsField returns the Appsettings field
func (c *Config) GetAppSettingsField() *AppSettings {
	return c.Appsettings
}

// SetAppSettings sets the Appsettings field
func (c *Config) SetAppSettings(settings *AppSettings) {
	c.Appsettings = settings
}
