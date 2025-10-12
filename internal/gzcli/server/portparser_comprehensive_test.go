package server

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadEnvFile_EdgeCases tests edge cases in .env file parsing
func TestLoadEnvFile_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected map[string]string
	}{
		{
			name: "malformed lines ignored",
			content: `VALID_KEY=value
INVALID LINE WITHOUT EQUALS
ANOTHER_VALID=123`,
			expected: map[string]string{
				"VALID_KEY":     "value",
				"ANOTHER_VALID": "123",
			},
		},
		{
			name: "keys with spaces are parsed",
			content: `VALID_KEY=value
SPACED KEY = value with spaces
NORMAL=test`,
			expected: map[string]string{
				"VALID_KEY":  "value",
				"SPACED KEY": "value with spaces",
				"NORMAL":     "test",
			},
		},
		{
			name: "empty values",
			content: `KEY_WITH_VALUE=value
KEY_WITHOUT_VALUE=
ANOTHER_KEY=another`,
			expected: map[string]string{
				"KEY_WITH_VALUE":    "value",
				"KEY_WITHOUT_VALUE": "",
				"ANOTHER_KEY":       "another",
			},
		},
		{
			name: "whitespace handling",
			content: `  KEY_WITH_LEADING_SPACE=value
KEY_WITH_TRAILING_SPACE=value
  KEY_WITH_BOTH  =  value  `,
			expected: map[string]string{
				"KEY_WITH_LEADING_SPACE":  "value",
				"KEY_WITH_TRAILING_SPACE": "value",
				"KEY_WITH_BOTH":           "value",
			},
		},
		{
			name: "special characters in values",
			content: `URL=https://example.com:8080/path?query=value
PATH=/usr/local/bin:/usr/bin
SPECIAL=value!@#$%^&*()`,
			expected: map[string]string{
				"URL":     "https://example.com:8080/path?query=value",
				"PATH":    "/usr/local/bin:/usr/bin",
				"SPECIAL": "value!@#$%^&*()",
			},
		},
		{
			name: "mixed quote styles",
			content: `DOUBLE="double quotes"
SINGLE='single quotes'
NO_QUOTES=no quotes
DOUBLE_WITH_SINGLE="it's a test"
SINGLE_WITH_DOUBLE='he said "hello"'`,
			expected: map[string]string{
				"DOUBLE":             "double quotes",
				"SINGLE":             "single quotes",
				"NO_QUOTES":          "no quotes",
				"DOUBLE_WITH_SINGLE": "it's a test",
				"SINGLE_WITH_DOUBLE": "he said \"hello\"",
			},
		},
		{
			name: "equals sign in value",
			content: `MATH=1+1=2
EQUATION=x=y+z`,
			expected: map[string]string{
				"MATH":     "1+1=2",
				"EQUATION": "x=y+z",
			},
		},
		{
			name: "only comments and empty lines",
			content: `# Comment 1
# Comment 2

# Comment 3`,
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			envPath := filepath.Join(tmpDir, ".env")
			if err := os.WriteFile(envPath, []byte(tt.content), 0600); err != nil {
				t.Fatalf("Failed to write test env file: %v", err)
			}

			got := loadEnvFile(envPath)

			if len(got) != len(tt.expected) {
				t.Errorf("Expected %d keys, got %d. Got: %v, Want: %v", len(tt.expected), len(got), got, tt.expected)
			}

			for key, expectedValue := range tt.expected {
				if gotValue, ok := got[key]; !ok {
					t.Errorf("Missing key %q", key)
				} else if gotValue != expectedValue {
					t.Errorf("Key %q: got %q, want %q", key, gotValue, expectedValue)
				}
			}
		})
	}
}

// TestLoadEnvFile_NonExistent tests that loading non-existent file returns empty map
func TestLoadEnvFile_NonExistent(t *testing.T) {
	envVars := loadEnvFile("/nonexistent/path/.env")
	if len(envVars) != 0 {
		t.Errorf("Expected empty map for non-existent file, got %v", envVars)
	}
}

// TestExpandEnvVarsWithMap_Priority tests environment variable priority
func TestExpandEnvVarsWithMap_Priority(t *testing.T) {
	// Set system env
	os.Setenv("SHARED_KEY", "system_value")
	os.Setenv("SYSTEM_ONLY", "system_only_value")
	defer func() {
		os.Unsetenv("SHARED_KEY")
		os.Unsetenv("SYSTEM_ONLY")
	}()

	envMap := map[string]string{
		"SHARED_KEY": "map_value", // Should override system env
		"MAP_ONLY":   "map_only_value",
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"${SHARED_KEY}", "map_value"},          // Map takes priority
		{"${MAP_ONLY}", "map_only_value"},       // From map
		{"${SYSTEM_ONLY}", "system_only_value"}, // Falls back to system
		{"${UNDEFINED}", ""},                    // Undefined becomes empty
	}

	for _, tt := range tests {
		got := expandEnvVarsWithMap(tt.input, envMap)
		if got != tt.expected {
			t.Errorf("expandEnvVarsWithMap(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

// TestExpandEnvVarsWithMap_ComplexPatterns tests complex variable patterns
func TestExpandEnvVarsWithMap_ComplexPatterns(t *testing.T) {
	envMap := map[string]string{
		"HOST":     "localhost",
		"PORT":     "3000",
		"PROTOCOL": "https",
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"${PROTOCOL}://${HOST}:${PORT}", "https://localhost:3000"},
		{"$PROTOCOL://$HOST:$PORT", "https://localhost:3000"},
		{"${HOST}:${PORT}/api/v1", "localhost:3000/api/v1"},
		{"prefix_${HOST}_suffix", "prefix_localhost_suffix"},
		{"${HOST}${HOST}", "localhostlocalhost"},
	}

	for _, tt := range tests {
		got := expandEnvVarsWithMap(tt.input, envMap)
		if got != tt.expected {
			t.Errorf("expandEnvVarsWithMap(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

// TestParseComposePorts_MultipleEnvFiles tests loading multiple env_file entries
func TestParseComposePorts_MultipleEnvFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .env (default)
	defaultEnv := `BASE_PORT=1000
OVERRIDE_ME=default`
	if err := os.WriteFile(filepath.Join(tmpDir, ".env"), []byte(defaultEnv), 0600); err != nil {
		t.Fatalf("Failed to write .env: %v", err)
	}

	// Create first.env
	firstEnv := `OVERRIDE_ME=first
FIRST_ONLY=100`
	if err := os.WriteFile(filepath.Join(tmpDir, "first.env"), []byte(firstEnv), 0600); err != nil {
		t.Fatalf("Failed to write first.env: %v", err)
	}

	// Create second.env (should override first.env)
	secondEnv := `OVERRIDE_ME=second
SECOND_ONLY=200`
	if err := os.WriteFile(filepath.Join(tmpDir, "second.env"), []byte(secondEnv), 0600); err != nil {
		t.Fatalf("Failed to write second.env: %v", err)
	}

	// Create docker-compose.yml with multiple env_file entries
	compose := `version: '3.8'
services:
  app:
    image: test
    env_file:
      - first.env
      - second.env
    ports:
      - "${BASE_PORT}:80"
      - "${OVERRIDE_ME}:81"
      - "${FIRST_ONLY}:82"
      - "${SECOND_ONLY}:83"`

	composePath := filepath.Join(tmpDir, "docker-compose.yml")
	if err := os.WriteFile(composePath, []byte(compose), 0600); err != nil {
		t.Fatalf("Failed to write docker-compose.yml: %v", err)
	}

	parser := NewPortParser()
	ports := parser.ParsePorts("compose", composePath, tmpDir)

	expected := []string{
		"1000:80",   // BASE_PORT from .env
		"second:81", // OVERRIDE_ME from second.env (overrides first.env)
		"100:82",    // FIRST_ONLY from first.env
		"200:83",    // SECOND_ONLY from second.env
	}

	if len(ports) != len(expected) {
		t.Fatalf("Expected %d ports, got %d: %v", len(expected), len(ports), ports)
	}

	// Check each port
	portMap := make(map[string]bool)
	for _, p := range ports {
		portMap[p] = true
	}

	for _, exp := range expected {
		if !portMap[exp] {
			t.Errorf("Missing expected port: %q. Got: %v", exp, ports)
		}
	}
}

// TestParseComposePorts_EnvFileAsArray tests env_file as array
func TestParseComposePorts_EnvFileAsArray(t *testing.T) {
	tmpDir := t.TempDir()

	// Create env files
	os.WriteFile(filepath.Join(tmpDir, "a.env"), []byte("PORT_A=1000"), 0600)
	os.WriteFile(filepath.Join(tmpDir, "b.env"), []byte("PORT_B=2000"), 0600)

	compose := `version: '3.8'
services:
  app:
    image: test
    env_file:
      - a.env
      - b.env
    ports:
      - "${PORT_A}:80"
      - "${PORT_B}:81"`

	composePath := filepath.Join(tmpDir, "docker-compose.yml")
	os.WriteFile(composePath, []byte(compose), 0600)

	parser := NewPortParser()
	ports := parser.ParsePorts("compose", composePath, tmpDir)

	if len(ports) != 2 {
		t.Fatalf("Expected 2 ports, got %d: %v", len(ports), ports)
	}

	portMap := make(map[string]bool)
	for _, p := range ports {
		portMap[p] = true
	}

	if !portMap["1000:80"] {
		t.Errorf("Missing port 1000:80 in %v", ports)
	}
	if !portMap["2000:81"] {
		t.Errorf("Missing port 2000:81 in %v", ports)
	}
}

// TestParseComposePorts_NoEnvFile tests that missing .env doesn't cause errors
func TestParseComposePorts_NoEnvFile(t *testing.T) {
	tmpDir := t.TempDir()

	compose := `version: '3.8'
services:
  app:
    image: test
    ports:
      - "3000:80"
      - "3001:81"`

	composePath := filepath.Join(tmpDir, "docker-compose.yml")
	if err := os.WriteFile(composePath, []byte(compose), 0600); err != nil {
		t.Fatalf("Failed to write docker-compose.yml: %v", err)
	}

	parser := NewPortParser()
	ports := parser.ParsePorts("compose", composePath, tmpDir)

	expected := []string{"3000:80", "3001:81"}

	if len(ports) != len(expected) {
		t.Fatalf("Expected %d ports, got %d: %v", len(expected), len(ports), ports)
	}
}

// TestParseComposePorts_SystemEnvFallback tests system env as fallback
func TestParseComposePorts_SystemEnvFallback(t *testing.T) {
	tmpDir := t.TempDir()

	// Set system env
	os.Setenv("SYSTEM_PORT", "9000")
	defer os.Unsetenv("SYSTEM_PORT")

	compose := `version: '3.8'
services:
  app:
    image: test
    ports:
      - "${SYSTEM_PORT}:80"`

	composePath := filepath.Join(tmpDir, "docker-compose.yml")
	os.WriteFile(composePath, []byte(compose), 0600)

	parser := NewPortParser()
	ports := parser.ParsePorts("compose", composePath, tmpDir)

	if len(ports) != 1 {
		t.Fatalf("Expected 1 port, got %d: %v", len(ports), ports)
	}

	if ports[0] != "9000:80" {
		t.Errorf("Expected 9000:80, got %s", ports[0])
	}
}

// TestParseComposePorts_AbsoluteEnvFilePath tests absolute path in env_file
func TestParseComposePorts_AbsoluteEnvFilePath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create env file in temp directory with absolute path
	absEnvPath := filepath.Join(tmpDir, "absolute.env")
	os.WriteFile(absEnvPath, []byte("ABS_PORT=7000"), 0600)

	compose := `version: '3.8'
services:
  app:
    image: test
    env_file:
      - ` + absEnvPath + `
    ports:
      - "${ABS_PORT}:80"`

	composePath := filepath.Join(tmpDir, "docker-compose.yml")
	os.WriteFile(composePath, []byte(compose), 0600)

	parser := NewPortParser()
	ports := parser.ParsePorts("compose", composePath, tmpDir)

	if len(ports) != 1 {
		t.Fatalf("Expected 1 port, got %d: %v", len(ports), ports)
	}

	if ports[0] != "7000:80" {
		t.Errorf("Expected 7000:80, got %s", ports[0])
	}
}

// Benchmark tests
func BenchmarkLoadEnvFile(b *testing.B) {
	tmpDir := b.TempDir()
	envContent := `PORT1=3000
PORT2=3001
PORT3=3002
PORT4=3003
PORT5=3004`
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(envPath, []byte(envContent), 0600)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = loadEnvFile(envPath)
	}
}

func BenchmarkExpandEnvVarsWithMap(b *testing.B) {
	envMap := map[string]string{
		"HOST": "localhost",
		"PORT": "3000",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = expandEnvVarsWithMap("${HOST}:${PORT}", envMap)
	}
}

func BenchmarkParseComposePorts(b *testing.B) {
	tmpDir := b.TempDir()

	compose := `version: '3.8'
services:
  web:
    image: nginx
    ports:
      - "8080:80"
      - "8443:443"
  api:
    image: node
    ports:
      - "3000:3000"`

	composePath := filepath.Join(tmpDir, "docker-compose.yml")
	os.WriteFile(composePath, []byte(compose), 0600)

	parser := NewPortParser()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parser.ParsePorts("compose", composePath, tmpDir)
	}
}
