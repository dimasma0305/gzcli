package server

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvFile(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()

	// Create test .env file
	envContent := `# Test environment file
PUBLIC_PORT=3000
WS_PUBLIC_PORT=3001
BACKEND_PORT=4000
# Comment line
QUOTED_VAR="value with spaces"
SINGLE_QUOTED='another value'
EMPTY_VAR=
`
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte(envContent), 0600); err != nil {
		t.Fatalf("Failed to write test env file: %v", err)
	}

	// Load the env file
	envVars := loadEnvFile(envPath)

	// Test expected values
	tests := []struct {
		key      string
		expected string
	}{
		{"PUBLIC_PORT", "3000"},
		{"WS_PUBLIC_PORT", "3001"},
		{"BACKEND_PORT", "4000"},
		{"QUOTED_VAR", "value with spaces"},
		{"SINGLE_QUOTED", "another value"},
	}

	for _, tt := range tests {
		if got := envVars[tt.key]; got != tt.expected {
			t.Errorf("envVars[%s] = %q, want %q", tt.key, got, tt.expected)
		}
	}
}

func TestExpandEnvVarsWithMap(t *testing.T) {
	envMap := map[string]string{
		"PUBLIC_PORT": "3000",
		"WS_PORT":     "3001",
	}

	// Set a system env var
	os.Setenv("SYSTEM_PORT", "8080")
	defer os.Unsetenv("SYSTEM_PORT")

	tests := []struct {
		input    string
		expected string
	}{
		{"${PUBLIC_PORT}:8080", "3000:8080"},
		{"${WS_PORT}:80", "3001:80"},
		{"$PUBLIC_PORT:80", "3000:80"},
		{"${SYSTEM_PORT}:3000", "8080:3000"}, // Falls back to system env
		{"${UNDEFINED}:80", ":80"},           // Undefined var becomes empty
		{"3000:8080", "3000:8080"},           // No vars to expand
	}

	for _, tt := range tests {
		got := expandEnvVarsWithMap(tt.input, envMap)
		if got != tt.expected {
			t.Errorf("expandEnvVarsWithMap(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestParseComposePorts_WithEnvFile(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()

	// Create .env file
	envContent := `PUBLIC_PORT=3000
WS_PUBLIC_PORT=3001
BACKEND_PORT=4000
`
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte(envContent), 0600); err != nil {
		t.Fatalf("Failed to write .env file: %v", err)
	}

	// Create custom env file
	customEnvContent := `API_PORT=5000
`
	customEnvPath := filepath.Join(tmpDir, "custom.env")
	if err := os.WriteFile(customEnvPath, []byte(customEnvContent), 0600); err != nil {
		t.Fatalf("Failed to write custom.env file: %v", err)
	}

	// Create docker-compose.yml
	composeContent := `version: '3.8'
services:
  web:
    image: nginx
    ports:
      - "${PUBLIC_PORT}:8080"
      - "${WS_PUBLIC_PORT}:8081"
    expose:
      - "${BACKEND_PORT}"

  api:
    image: node
    env_file:
      - custom.env
    ports:
      - "${API_PORT}:3000"
`
	composePath := filepath.Join(tmpDir, "docker-compose.yml")
	if err := os.WriteFile(composePath, []byte(composeContent), 0600); err != nil {
		t.Fatalf("Failed to write docker-compose.yml: %v", err)
	}

	// Parse ports
	parser := NewPortParser()
	ports := parser.ParsePorts("compose", composePath, tmpDir)

	// Expected ports (order may vary due to map iteration)
	expected := []string{
		"3000:8080", // web - PUBLIC_PORT
		"3001:8081", // web - WS_PUBLIC_PORT
		"*:4000",    // web - BACKEND_PORT (expose)
		"5000:3000", // api - API_PORT (from custom.env)
	}

	if len(ports) != len(expected) {
		t.Fatalf("Expected %d ports, got %d: %v", len(expected), len(ports), ports)
	}

	// Use map for order-independent comparison
	portMap := make(map[string]bool)
	for _, p := range ports {
		portMap[p] = true
	}

	for _, exp := range expected {
		if !portMap[exp] {
			t.Errorf("Missing expected port %q. Got: %v", exp, ports)
		}
	}
}
