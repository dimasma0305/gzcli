package gzcli

import (
	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
)

// ConfigManager handles configuration management for gzcli
type ConfigManager struct{}

// NewConfigManager creates a new config manager
func NewConfigManager() *ConfigManager {
	return &ConfigManager{}
}

// Config is a compatibility wrapper that allows lowercase appsettings field for watcher.go
type Config struct {
	Url         string       `yaml:"url"` //nolint:revive // Field name required for watcher.go compatibility
	Creds       gzapi.Creds  `yaml:"creds"`
	Event       gzapi.Game   `yaml:"event"`
	appsettings *AppSettings `yaml:"-"`
	Appsettings *AppSettings `yaml:"-"` // Public field for external access
}

// ToConfigPackage converts to config.Config
func (c *Config) ToConfigPackage() *config.Config {
	settings := c.Appsettings
	if settings == nil {
		settings = c.appsettings
	}
	return &config.Config{
		Url:         c.Url,
		Creds:       c.Creds,
		Event:       c.Event,
		Appsettings: settings,
	}
}

// FromConfigPackage converts from config.Config
func FromConfigPackage(conf *config.Config) *Config {
	settings := conf.Appsettings
	return &Config{
		Url:         conf.Url,
		Creds:       conf.Creds,
		Event:       conf.Event,
		appsettings: settings,
		Appsettings: settings,
	}
}

// SetAppSettings sets both appsettings fields
func (c *Config) SetAppSettings(settings *AppSettings) {
	c.appsettings = settings
	c.Appsettings = settings
}

// GetAppSettingsField returns the settings
func (c *Config) GetAppSettingsField() *AppSettings {
	if c.Appsettings != nil {
		return c.Appsettings
	}
	return c.appsettings
}