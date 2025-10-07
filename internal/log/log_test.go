//nolint:errcheck,gosec // Test file with acceptable error handling patterns
package log

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestSetDebugMode(t *testing.T) {
	// Save original state
	originalDebugMode := debugMode
	defer func() { debugMode = originalDebugMode }()

	tests := []struct {
		name    string
		enabled bool
	}{
		{
			name:    "enable debug",
			enabled: true,
		},
		{
			name:    "disable debug",
			enabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetDebugMode(tt.enabled)
			if debugMode != tt.enabled {
				t.Errorf("SetDebugMode(%v) did not set debugMode correctly", tt.enabled)
			}
		})
	}
}

func TestDebugOutput(t *testing.T) {
	// Save original state
	originalDebugMode := debugMode
	defer func() { debugMode = originalDebugMode }()

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Enable debug mode and test
	SetDebugMode(true)
	Debug("test %s", "message")

	// Restore stdout
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "test message") {
		t.Errorf("Debug() did not output expected message, got: %s", output)
	}
	if !strings.Contains(output, "[DEBUG]") {
		t.Errorf("Debug() did not include [DEBUG] prefix, got: %s", output)
	}
}

func TestDebugDisabled(t *testing.T) {
	// Save original state
	originalDebugMode := debugMode
	defer func() { debugMode = originalDebugMode }()

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Disable debug mode and test
	SetDebugMode(false)
	Debug("test message")

	// Restore stdout
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if output != "" {
		t.Errorf("Debug() should not output when disabled, got: %s", output)
	}
}
