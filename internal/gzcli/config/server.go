//nolint:revive // Config struct field names match YAML/API structure
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/utils"
)

// ServerConfig represents server-level configuration
type ServerConfig struct {
	Url   string      `yaml:"url"`
	Creds gzapi.Creds `yaml:"creds"`
}

// GetServerConfig reads server configuration from .gzctf/conf.yaml
func GetServerConfig() (*ServerConfig, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	confPath := filepath.Join(dir, GZCTF_DIR, CONFIG_FILE)
	var config ServerConfig
	if err := utils.ParseYamlFromFile(confPath, &config); err != nil {
		return nil, fmt.Errorf("failed to read server config %s: %w", confPath, err)
	}

	return &config, nil
}
