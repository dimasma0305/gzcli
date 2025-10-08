//nolint:revive // Config struct field names match YAML/API structure
package config

import (
	"fmt"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/log"
)

const (
	GZCTF_DIR   = ".gzctf"
	CONFIG_FILE = "conf.yaml"
)

// Config represents the combined application configuration (server + event)
type Config struct {
	Url         string       `yaml:"url"`
	Creds       gzapi.Creds  `yaml:"creds"`
	Event       gzapi.Game   `yaml:"event"`
	Appsettings *AppSettings `yaml:"-"`
	EventName   string       `yaml:"-"` // Current event name
}

// loadConfigFromCache loads cached config data (backward compatibility wrapper)
//
//nolint:unused // Kept for backward compatibility
func loadConfigFromCache(config *Config, getCache func(string, interface{}) error) {
	loadConfigFromCacheWithKey(config, getCache, "config")
}

// loadConfigFromCacheWithKey loads cached config data with a specific cache key
func loadConfigFromCacheWithKey(config *Config, getCache func(string, interface{}) error, cacheKey string) {
	var configCache Config
	cacheErr := getCache(cacheKey, &configCache)

	// If we have cached game info, use it as the starting point
	if cacheErr == nil && configCache.Event.Id != 0 {
		config.Event.Id = configCache.Event.Id
		config.Event.PublicKey = configCache.Event.PublicKey
		log.Info("Using cached game ID: %d", config.Event.Id)
	}
}

// validateCachedGame validates if a cached game ID still exists on the server (backward compatibility wrapper)
//
//nolint:unused // Kept for backward compatibility
func validateCachedGame(config *Config, api *gzapi.GZAPI, deleteCache func(string)) error {
	return validateCachedGameWithKey(config, api, deleteCache, "config")
}

// validateCachedGameWithKey validates if a cached game ID still exists on the server with a specific cache key
func validateCachedGameWithKey(config *Config, api *gzapi.GZAPI, deleteCache func(string), cacheKey string) error {
	if config.Event.Id == 0 {
		return nil
	}

	log.Info("Validating cached game ID %d exists on server...", config.Event.Id)
	games, err := api.GetGames()
	if err != nil {
		log.Error("Failed to get games for validation: %v", err)
		return fmt.Errorf("API games fetch error: %w", err)
	}

	// Check if the cached game ID still exists
	for _, game := range games {
		if game.Id == config.Event.Id {
			// Update with current server data but keep the same ID
			config.Event.PublicKey = game.PublicKey
			log.Info("Cached game ID %d validated successfully", config.Event.Id)
			return nil
		}
	}

	// If cached game doesn't exist, clear cache and try to find by title
	log.Info("Cached game ID %d not found on server, searching by title...", config.Event.Id)
	deleteCache(cacheKey)
	config.Event.Id = 0
	config.Event.PublicKey = ""

	return nil
}

// ensureGameExists ensures a game exists by title or creates a new one (backward compatibility wrapper)
//
//nolint:unused // Kept for backward compatibility
func ensureGameExists(config *Config, api *gzapi.GZAPI, setCache func(string, interface{}) error, createNewGame func(*Config, *gzapi.GZAPI) (*gzapi.Game, error)) error {
	return ensureGameExistsWithKey(config, api, setCache, createNewGame, "config")
}

// ensureGameExistsWithKey ensures a game exists by title or creates a new one with a specific cache key
func ensureGameExistsWithKey(config *Config, api *gzapi.GZAPI, setCache func(string, interface{}) error, createNewGame func(*Config, *gzapi.GZAPI) (*gzapi.Game, error), cacheKey string) error {
	if config.Event.Id != 0 {
		return nil
	}

	game, err := api.GetGameByTitle(config.Event.Title)
	if err != nil {
		log.Info("Game '%s' not found by title, creating new game...", config.Event.Title)
		_, err = createNewGame(config, api)
		if err != nil {
			return fmt.Errorf("failed to create new game: %w", err)
		}
		return nil
	}

	log.Info("Found existing game by title: %s (ID: %d)", game.Title, game.Id)
	config.Event.Id = game.Id
	config.Event.PublicKey = game.PublicKey

	// Update cache with found game
	if err := setCache(cacheKey, config); err != nil {
		log.Error("Failed to update cache with found game: %v", err)
	}

	return nil
}

func GetConfig(api *gzapi.GZAPI, getCache func(string, interface{}) error, setCache func(string, interface{}) error, deleteCache func(string), createNewGame func(*Config, *gzapi.GZAPI) (*gzapi.Game, error)) (*Config, error) {
	return GetConfigWithEvent(api, "", getCache, setCache, deleteCache, createNewGame)
}

// GetConfigWithEvent loads configuration for a specific event
// If eventName is empty, it will be auto-detected
func GetConfigWithEvent(api *gzapi.GZAPI, eventName string, getCache func(string, interface{}) error, setCache func(string, interface{}) error, deleteCache func(string), createNewGame func(*Config, *gzapi.GZAPI) (*gzapi.Game, error)) (*Config, error) {
	// Determine current event
	if eventName == "" {
		var err error
		eventName, err = GetCurrentEvent("")
		if err != nil {
			return nil, err
		}
	}

	// Load server config
	serverConfig, err := GetServerConfig()
	if err != nil {
		return nil, err
	}

	// Load event config
	eventConfig, err := GetEventConfig(eventName)
	if err != nil {
		return nil, err
	}

	// Merge into unified Config struct
	config := &Config{
		Url:       serverConfig.Url,
		Creds:     serverConfig.Creds,
		Event:     eventConfig.Game,
		EventName: eventName,
	}

	// Load cache for this specific event
	cacheKey := fmt.Sprintf("config-%s", eventName)
	loadConfigFromCacheWithKey(config, getCache, cacheKey)

	// Only interact with API if provided and we need to validate/create game
	if api != nil && api.Client != nil {
		if err := validateCachedGameWithKey(config, api, deleteCache, cacheKey); err != nil {
			return nil, err
		}

		if err := ensureGameExistsWithKey(config, api, setCache, createNewGame, cacheKey); err != nil {
			return nil, err
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

	return config, nil
}

// GetAppSettingsField returns the Appsettings field
func (c *Config) GetAppSettingsField() *AppSettings {
	return c.Appsettings
}

// SetAppSettings sets the Appsettings field
func (c *Config) SetAppSettings(settings *AppSettings) {
	c.Appsettings = settings
}
