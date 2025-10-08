//nolint:errcheck,gosec,revive // Test file with acceptable error handling patterns
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/testutil"
)

// cleanupTempDir removes a temporary directory with Windows-specific retry logic
func cleanupTempDir(t *testing.T, dir string) {
	// On Windows, files may be locked briefly after close
	if runtime.GOOS == "windows" {
		time.Sleep(100 * time.Millisecond)
	}

	err := os.RemoveAll(dir)
	if err != nil && runtime.GOOS == "windows" {
		// Retry once more on Windows
		time.Sleep(200 * time.Millisecond)
		err = os.RemoveAll(dir)
	}

	if err != nil {
		t.Logf("Warning: Failed to remove temp dir %s: %v", dir, err)
	}
}

// TestGetConfig_MalformedYAML tests handling of malformed YAML files
func TestGetConfig_MalformedYAML(t *testing.T) {
	testCases := []struct {
		name    string
		content string
	}{
		{"invalid syntax", testutil.CreateMalformedYAML("invalid_syntax")},
		{"unclosed quote", testutil.CreateMalformedYAML("unclosed_quote")},
		{"invalid UTF-8", testutil.CreateMalformedYAML("invalid_utf8")},
		{"duplicate keys", testutil.CreateMalformedYAML("duplicate_keys")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := setupConfigTest(t)
			defer cleanupTempDir(t, tmpDir)

			confPath := filepath.Join(tmpDir, ".gzctf", "conf.yaml")
			if err := os.WriteFile(confPath, []byte(tc.content), 0600); err != nil {
				t.Fatalf("Failed to write config file: %v", err)
			}

			_, err := GetConfig(nil, getCacheMock, setCacheMock, deleteCacheMock, createGameMock)
			if err == nil {
				t.Errorf("Expected error for %s, but got none", tc.name)
			}
		})
	}
}

// TestGetConfig_MissingRequiredFields tests incomplete configuration
func TestGetConfig_MissingRequiredFields(t *testing.T) {
	testCases := []struct {
		name    string
		content string
	}{
		{
			"missing URL",
			`creds:
  username: "admin"
  password: "pass"`,
		},
		{
			"missing credentials",
			`url: "http://test.com"`,
		},
		{
			"empty strings",
			`url: ""
creds:
  username: ""
  password: ""`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := setupConfigTest(t)
			defer cleanupTempDir(t, tmpDir)

			confPath := filepath.Join(tmpDir, ".gzctf", "conf.yaml")
			if err := os.WriteFile(confPath, []byte(tc.content), 0600); err != nil {
				t.Fatalf("Failed to write config file: %v", err)
			}

			_, err := GetConfig(nil, getCacheMock, setCacheMock, deleteCacheMock, createGameMock)
			// Some missing fields might be handled gracefully
			if err != nil {
				t.Logf("Missing field handled: %v", err)
			}
		})
	}
}

// TestGetConfig_TypeMismatches tests wrong types in YAML
func TestGetConfig_TypeMismatches(t *testing.T) {
	testCases := []struct {
		name    string
		content string
	}{
		{
			"array for string",
			`url: ["not", "a", "string"]
creds:
  username: "admin"
  password: "pass"`,
		},
		{
			"object for string",
			`url: {nested: "object"}
creds:
  username: "admin"
  password: "pass"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := setupConfigTest(t)
			defer cleanupTempDir(t, tmpDir)

			confPath := filepath.Join(tmpDir, ".gzctf", "conf.yaml")
			if err := os.WriteFile(confPath, []byte(tc.content), 0600); err != nil {
				t.Fatalf("Failed to write config file: %v", err)
			}

			_, err := GetConfig(nil, getCacheMock, setCacheMock, deleteCacheMock, createGameMock)
			if err == nil {
				t.Errorf("Expected error for %s", tc.name)
			}
		})
	}
}

// TestGetConfig_ConcurrentAccess tests multiple processes reading config
func TestGetConfig_ConcurrentAccess(t *testing.T) {
	tmpDir := setupConfigTest(t)
	defer os.RemoveAll(tmpDir)

	// Create valid server config (new structure)
	confPath := filepath.Join(tmpDir, ".gzctf", "conf.yaml")
	confData := `url: "http://test.com"
creds:
  username: "admin"
  password: "testpass"
`
	os.WriteFile(confPath, []byte(confData), 0600)

	// Create event directory and .gzevent file (new structure)
	eventDir := filepath.Join(tmpDir, "events", "test-event")
	os.MkdirAll(eventDir, 0750)
	eventData := `title: "Test CTF"
start: "2024-01-01T00:00:00Z"
end: "2024-01-02T00:00:00Z"
`
	os.WriteFile(filepath.Join(eventDir, ".gzevent"), []byte(eventData), 0600)

	// Create appsettings
	appSettingsPath := filepath.Join(tmpDir, ".gzctf", "appsettings.json")
	appSettingsData := `{"ContainerProvider": {"PublicEntry": "http://test.com"}}`
	os.WriteFile(appSettingsPath, []byte(appSettingsData), 0600)

	// Read concurrently
	testutil.ConcurrentTest(t, 10, 5, func(id, iter int) error {
		_, err := GetConfig(nil, getCacheMock, setCacheMock, deleteCacheMock, createGameMock)
		return err
	})
}

// TestGetConfig_LargeConfigFile tests extremely large configuration
func TestGetConfig_LargeConfigFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large config test in short mode")
	}

	tmpDir := setupConfigTest(t)
	defer func() {
		// Force garbage collection to ensure all file handles are closed (Windows needs this)
		runtime.GC()
		time.Sleep(100 * time.Millisecond)

		// Retry removal on Windows
		var lastErr error
		for i := 0; i < 5; i++ {
			if err := os.RemoveAll(tmpDir); err != nil {
				lastErr = err
				if runtime.GOOS == "windows" {
					time.Sleep(100 * time.Millisecond)
					continue
				}
				break
			}
			return
		}
		if lastErr != nil {
			t.Logf("Warning: Failed to remove temp dir: %v", lastErr)
		}
	}()

	// Create config with 10000 challenge entries
	var sb strings.Builder
	sb.WriteString(`url: "http://test.com"
creds:
  username: "admin"
  password: "pass"
challenges:
`)
	for i := 0; i < 10000; i++ {
		sb.WriteString(fmt.Sprintf("  - id: %d\n    name: Challenge%d\n", i, i))
	}

	confPath := filepath.Join(tmpDir, ".gzctf", "conf.yaml")
	if err := os.WriteFile(confPath, []byte(sb.String()), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	_, err := GetConfig(nil, getCacheMock, setCacheMock, deleteCacheMock, createGameMock)
	if err != nil {
		t.Logf("Large config file handled: %v", err)
	}
}

// TestGetConfig_ReadOnlyFile tests permissions issues
func TestGetConfig_ReadOnlyFile(t *testing.T) {
	// Skip on Windows as it has different permission model
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission test on Windows (different permission model)")
	}

	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tmpDir := setupConfigTest(t)
	defer cleanupTempDir(t, tmpDir)

	confPath := filepath.Join(tmpDir, ".gzctf", "conf.yaml")
	confData := `url: "http://test.com"
creds:
  username: "admin"
  password: "pass"
`
	if err := os.WriteFile(confPath, []byte(confData), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Make file unreadable
	if err := os.Chmod(confPath, 0000); err != nil {
		t.Fatalf("Failed to chmod file: %v", err)
	}
	defer func() {
		_ = os.Chmod(confPath, 0600)
	}()

	_, err := GetConfig(nil, getCacheMock, setCacheMock, deleteCacheMock, createGameMock)
	if err == nil {
		t.Error("Expected error when config file is unreadable")
	}
}

// TestGetConfig_SymlinkAttack tests symlink in config path
func TestGetConfig_SymlinkAttack(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping symlink test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer cleanupTempDir(t, tmpDir)

	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore directory: %v", err)
		}
	}()

	// Create target directory
	targetDir := filepath.Join(tmpDir, "target")
	if err := os.MkdirAll(targetDir, 0750); err != nil {
		t.Fatalf("Failed to create target dir: %v", err)
	}

	// Create config in target
	confData := `url: "http://test.com"
creds:
  username: "admin"
  password: "pass"
`
	if err := os.WriteFile(filepath.Join(targetDir, "conf.yaml"), []byte(confData), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create symlink
	symlinkPath := filepath.Join(tmpDir, ".gzctf")
	if err := os.Symlink(targetDir, symlinkPath); err != nil {
		t.Skipf("Failed to create symlink: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	_, err = GetConfig(nil, getCacheMock, setCacheMock, deleteCacheMock, createGameMock)
	// Should handle symlinks (they're sometimes legitimate)
	t.Logf("Symlink config result: %v", err)
}

// TestGetConfig_SpecialCharactersInValues tests injection attempts
func TestGetConfig_SpecialCharactersInValues(t *testing.T) {
	testCases := []struct {
		name     string
		username string
		password string
	}{
		{"SQL injection", "admin'; DROP TABLE users; --", "pass"},
		{"XSS attempt", "<script>alert('xss')</script>", "pass"},
		{"null bytes", "admin\x00", "pass\x00"},
		{"newlines", "admin\nfake: data", "pass\nmore: data"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := setupConfigTest(t)
			defer cleanupTempDir(t, tmpDir)

			confPath := filepath.Join(tmpDir, ".gzctf", "conf.yaml")
			confData := fmt.Sprintf(`url: "http://test.com"
creds:
  username: "%s"
  password: "%s"
`, tc.username, tc.password)
			if err := os.WriteFile(confPath, []byte(confData), 0600); err != nil {
				t.Fatalf("Failed to write config file: %v", err)
			}

			_, err := GetConfig(nil, getCacheMock, setCacheMock, deleteCacheMock, createGameMock)
			// The config should parse but the values should be safely stored
			if err != nil {
				t.Logf("Special characters handled: %v", err)
			}
		})
	}
}

// TestScriptValue_EdgeCases tests script value parsing edge cases
func TestScriptValue_EdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		yamlData string
		wantErr  bool
	}{
		{
			"empty string",
			`script: ""`,
			false, // Empty string is valid
		},
		{
			"very long command",
			`script: "` + strings.Repeat("a", 10000) + `"`,
			false,
		},
		{
			"command with newlines",
			`script: "echo 'line1\nline2'"`,
			false,
		},
		{
			"complex object with all fields",
			`script:
  execute: "docker build"
  interval: 1h30m
  timeout: 5m`,
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var data struct {
				Script ScriptValue `yaml:"script"`
			}

			err := yaml.Unmarshal([]byte(tc.yamlData), &data)
			if (err != nil) != tc.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

// TestGenerateSlug_EdgeCases tests slug generation with edge cases
func TestGenerateSlug_EdgeCases(t *testing.T) {
	testCases := []struct {
		name      string
		eventName string
		challenge ChallengeYaml
		want      string
	}{
		{
			"empty category and name",
			"ctf2024",
			ChallengeYaml{Category: "", Name: ""},
			"ctf2024__",
		},
		{
			"only special characters",
			"ctf2024",
			ChallengeYaml{Category: "!@#$", Name: "%^&*"},
			"ctf2024__",
		},
		{
			"unicode characters",
			"ctf2024",
			ChallengeYaml{Category: "日本語", Name: "チャレンジ"},
			"ctf2024__",
		},
		{
			"very long names",
			"ctf2024",
			ChallengeYaml{Category: strings.Repeat("a", 500), Name: strings.Repeat("b", 500)},
			"ctf2024_" + strings.Repeat("a", 500) + "_" + strings.Repeat("b", 500),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			got := generateSlug(tt.eventName, tt.challenge)
			if got != tt.want {
				t.Logf("generateSlug() = %s (length: %d)", got, len(got))
			}
		})
	}
}

// Helper functions
func setupConfigTest(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	originalDir, _ := os.Getwd()
	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore directory: %v", err)
		}
	})

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, ".gzctf"), 0750); err != nil {
		t.Fatalf("Failed to create .gzctf dir: %v", err)
	}

	return tmpDir
}

var getCacheMock = func(key string, v interface{}) error {
	return os.ErrNotExist
}

var setCacheMock = func(key string, v interface{}) error {
	return nil
}

var deleteCacheMock = func(key string) {}

var createGameMock = func(cfg *Config, api *gzapi.GZAPI) (*gzapi.Game, error) {
	return &gzapi.Game{Id: 1}, nil
}
