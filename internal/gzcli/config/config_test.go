package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
)

func TestScriptValue_UnmarshalYAML_Simple(t *testing.T) {
	yamlData := `script: "echo hello"`

	var data struct {
		Script ScriptValue `yaml:"script"`
	}

	err := yaml.Unmarshal([]byte(yamlData), &data)
	if err != nil {
		t.Errorf("UnmarshalYAML() for simple script failed: %v", err)
	}

	if !data.Script.IsSimple() {
		t.Error("Expected simple script")
	}

	if data.Script.GetCommand() != "echo hello" {
		t.Errorf("Expected command 'echo hello', got %s", data.Script.GetCommand())
	}
}

func TestScriptValue_UnmarshalYAML_Complex(t *testing.T) {
	yamlData := `script:
  execute: "docker build"
  interval: 5m`

	var data struct {
		Script ScriptValue `yaml:"script"`
	}

	err := yaml.Unmarshal([]byte(yamlData), &data)
	if err != nil {
		t.Errorf("UnmarshalYAML() for complex script failed: %v", err)
	}

	if data.Script.IsSimple() {
		t.Error("Expected complex script")
	}

	if data.Script.GetCommand() != "docker build" {
		t.Errorf("Expected command 'docker build', got %s", data.Script.GetCommand())
	}

	if data.Script.GetInterval() != 5*time.Minute {
		t.Errorf("Expected interval 5m, got %v", data.Script.GetInterval())
	}

	if !data.Script.HasInterval() {
		t.Error("Expected HasInterval() to be true")
	}
}

func TestScriptValue_UnmarshalYAML_Invalid(t *testing.T) {
	yamlData := `script: [1, 2, 3]` // Invalid: array

	var data struct {
		Script ScriptValue `yaml:"script"`
	}

	err := yaml.Unmarshal([]byte(yamlData), &data)
	if err == nil {
		t.Error("Expected error for invalid script format")
	}

	if !strings.Contains(err.Error(), "script value must be") {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name      string
		challenge ChallengeYaml
		want      string
	}{
		{
			name: "simple slug",
			challenge: ChallengeYaml{
				Category: "Web",
				Name:     "Challenge1",
			},
			want: "web_challenge1",
		},
		{
			name: "with spaces",
			challenge: ChallengeYaml{
				Category: "Web Security",
				Name:     "SQL Injection",
			},
			want: "web_security_sql_injection",
		},
		{
			name: "with special characters",
			challenge: ChallengeYaml{
				Category: "Crypto",
				Name:     "RSA-2048",
			},
			want: "crypto_rsa2048",
		},
		{
			name: "uppercase to lowercase",
			challenge: ChallengeYaml{
				Category: "PWN",
				Name:     "Buffer Overflow",
			},
			want: "pwn_buffer_overflow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateSlug(tt.challenge)
			if got != tt.want {
				t.Errorf("generateSlug() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestGetConfig_ConfigFileNotFound(t *testing.T) {
	originalDir, _ := os.Getwd()
	tmpDir, err := os.MkdirTemp("", "config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("Failed to remove temp dir: %v", err)
		}
	}()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	getCache := func(key string, v interface{}) error {
		return os.ErrNotExist
	}

	setCache := func(key string, v interface{}) error {
		return nil
	}

	deleteCache := func(key string) {}

	createNewGame := func(cfg *Config, api *gzapi.GZAPI) (*gzapi.Game, error) {
		return nil, nil
	}

	_, err = GetConfig(nil, getCache, setCache, deleteCache, createNewGame)
	if err == nil {
		t.Error("Expected error when config file doesn't exist")
	}
}

func TestGetConfig_WithValidCache(t *testing.T) {
	originalDir, _ := os.Getwd()
	tmpDir, err := os.MkdirTemp("", "config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("Failed to remove temp dir: %v", err)
		}
	}()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Create .gzctf directory
	gzctfDir := filepath.Join(tmpDir, ".gzctf")
	if err := os.Mkdir(gzctfDir, 0755); err != nil {
		t.Fatalf("Failed to create .gzctf dir: %v", err)
	}

	// Create conf.yaml
	confPath := filepath.Join(gzctfDir, "conf.yaml")
	confData := `url: "http://test.com"
creds:
  username: "admin"
  password: "testpass"
event:
  title: "Test CTF"
  start: "2024-01-01T00:00:00Z"
  end: "2024-01-02T00:00:00Z"
`
	if err := os.WriteFile(confPath, []byte(confData), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create appsettings.json
	appSettingsPath := filepath.Join(gzctfDir, "appsettings.json")
	appSettingsData := `{
		"ContainerProvider": {
			"PublicEntry": "http://containers.test.com"
		},
		"EmailConfig": {
			"UserName": "test@example.com",
			"Password": "emailpass",
			"Smtp": {
				"Host": "smtp.example.com",
				"Port": 587
			}
		}
	}`
	if err := os.WriteFile(appSettingsPath, []byte(appSettingsData), 0600); err != nil {
		t.Fatalf("Failed to write appsettings file: %v", err)
	}

	cacheData := map[string]interface{}{
		"config": Config{
			Event: gzapi.Game{
				Id:        123,
				PublicKey: "cached-key",
			},
		},
	}

	getCache := func(key string, v interface{}) error {
		if data, ok := cacheData[key]; ok {
			if ptr, ok := v.(*Config); ok {
				if cached, ok := data.(Config); ok {
					*ptr = cached
					return nil
				}
			}
		}
		return os.ErrNotExist
	}

	setCache := func(key string, v interface{}) error {
		cacheData[key] = v
		return nil
	}

	deleteCache := func(key string) {
		delete(cacheData, key)
	}

	createNewGame := func(cfg *Config, api *gzapi.GZAPI) (*gzapi.Game, error) {
		return &gzapi.Game{Id: 456}, nil
	}

	config, err := GetConfig(nil, getCache, setCache, deleteCache, createNewGame)
	if err != nil {
		t.Errorf("GetConfig() failed: %v", err)
	}

	if config == nil {
		t.Fatal("Expected non-nil config")
	}

	// Should use cached game ID
	if config.Event.Id != 123 {
		t.Errorf("Expected cached game ID 123, got %d", config.Event.Id)
	}

	if config.Appsettings == nil {
		t.Error("Expected Appsettings to be loaded")
	}
}

func TestConfig_GetAppSettingsField(t *testing.T) {
	appSettings := &AppSettings{}
	appSettings.ContainerProvider.PublicEntry = "http://test.com"

	config := &Config{
		Appsettings: appSettings,
	}

	result := config.GetAppSettingsField()
	if result != appSettings {
		t.Error("GetAppSettingsField() returned different pointer")
	}
}

func TestConfig_SetAppSettings(t *testing.T) {
	config := &Config{}

	appSettings := &AppSettings{}
	appSettings.ContainerProvider.PublicEntry = "http://test.com"

	config.SetAppSettings(appSettings)

	if config.Appsettings != appSettings {
		t.Error("SetAppSettings() did not set the field correctly")
	}
}
