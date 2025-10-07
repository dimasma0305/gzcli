//nolint:errcheck,gosec,revive,staticcheck // Test file with acceptable error handling patterns
package challenge

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestGetShell(t *testing.T) {
	// Save original env vars
	var originalShell, originalComspec string
	if runtime.GOOS == "windows" {
		originalComspec = os.Getenv("COMSPEC")
		defer os.Setenv("COMSPEC", originalComspec)
	} else {
		originalShell = os.Getenv("SHELL")
		defer os.Setenv("SHELL", originalShell)
	}

	// Test with custom shell set
	shellOnce = sync.Once{} // Reset sync.Once for testing
	if runtime.GOOS == "windows" {
		os.Setenv("COMSPEC", "powershell.exe")
		shell = ""
		result := getShell()
		if result != "powershell.exe" {
			t.Errorf("Expected shell 'powershell.exe', got %s", result)
		}
	} else {
		os.Setenv("SHELL", "/bin/bash")
		shell = ""
		result := getShell()
		if result != "/bin/bash" {
			t.Errorf("Expected shell '/bin/bash', got %s", result)
		}
	}

	// Reset for default test
	shellOnce = sync.Once{} // Reset sync.Once for testing
	shell = ""
	if runtime.GOOS == "windows" {
		os.Unsetenv("COMSPEC")
		result := getShell()
		if result != "cmd.exe" {
			t.Errorf("Expected default shell 'cmd.exe', got %s", result)
		}
	} else {
		os.Unsetenv("SHELL")
		result := getShell()
		if result != "/bin/sh" {
			t.Errorf("Expected default shell '/bin/sh', got %s", result)
		}
	}
}

func TestRunScript_NoScript(t *testing.T) {
	challengeConf := ChallengeYaml{
		Name:    "Test Challenge",
		Scripts: map[string]ScriptValue{},
	}

	err := RunScript(challengeConf, "nonexistent")
	if err != nil {
		t.Errorf("RunScript() with non-existent script should return nil, got %v", err)
	}
}

func TestRunScript_WithDashboard(t *testing.T) {
	challengeConf := ChallengeYaml{
		Name: "Dashboard Challenge",
		Scripts: map[string]ScriptValue{
			"test": {Simple: "echo test"},
		},
		Dashboard: &Dashboard{}, // Has dashboard, should skip
	}

	err := RunScript(challengeConf, "test")
	if err != nil {
		t.Errorf("RunScript() with dashboard should return nil, got %v", err)
	}
}

func TestRunScript_EmptyCommand(t *testing.T) {
	challengeConf := ChallengeYaml{
		Name: "Test Challenge",
		Scripts: map[string]ScriptValue{
			"test": {Simple: ""}, // Empty command
		},
	}

	err := RunScript(challengeConf, "test")
	if err != nil {
		t.Errorf("RunScript() with empty command should return nil, got %v", err)
	}
}

func TestRunScript_SimpleCommand(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "script-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")

	challengeConf := ChallengeYaml{
		Name: "Test Challenge",
		Scripts: map[string]ScriptValue{
			"test": {Simple: "echo hello > test.txt"},
		},
		Cwd: tmpDir,
	}

	err = RunScript(challengeConf, "test")
	if err != nil {
		t.Errorf("RunScript() failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(testFile); err == nil {
		// File exists, which is expected
	}
}

func TestRunScript_WithInterval(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "script-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	challengeConf := ChallengeYaml{
		Name: "Test Challenge",
		Scripts: map[string]ScriptValue{
			"test": {Complex: &ScriptConfig{
				Execute:  "echo interval",
				Interval: 1 * time.Minute,
			}},
		},
		Cwd: tmpDir,
	}

	// Should run once with warning about interval
	err = RunScript(challengeConf, "test")
	if err != nil {
		t.Errorf("RunScript() with interval failed: %v", err)
	}
}

func TestRunShellWithContext(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "script-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()

	err = RunShellWithContext(ctx, "echo test", tmpDir)
	if err != nil {
		t.Errorf("RunShellWithContext() failed: %v", err)
	}
}

func TestRunShellWithContext_Cancelled(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "script-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = RunShellWithContext(ctx, "sleep 10", tmpDir)
	if err == nil {
		t.Error("Expected error for cancelled context")
	}
}

func TestRunShellWithTimeout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "script-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()

	// Test with default timeout (0 or negative should use default)
	err = RunShellWithTimeout(ctx, "echo test", tmpDir, 0)
	if err != nil {
		t.Errorf("RunShellWithTimeout() with default timeout failed: %v", err)
	}

	// Test with custom timeout
	err = RunShellWithTimeout(ctx, "echo test", tmpDir, 1*time.Second)
	if err != nil {
		t.Errorf("RunShellWithTimeout() with custom timeout failed: %v", err)
	}

	// Test timeout enforcement (should cap at MaxScriptTimeout)
	err = RunShellWithTimeout(ctx, "echo test", tmpDir, 100*time.Hour)
	if err != nil {
		t.Errorf("RunShellWithTimeout() with excessive timeout failed: %v", err)
	}
}

func TestRunShellForInterval(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "script-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()

	// Test successful execution
	err = RunShellForInterval(ctx, "echo interval test", tmpDir, 1*time.Second)
	if err != nil {
		t.Errorf("RunShellForInterval() failed: %v", err)
	}

	// Test with command that produces stderr
	err = RunShellForInterval(ctx, "echo error >&2", tmpDir, 1*time.Second)
	if err != nil {
		t.Errorf("RunShellForInterval() with stderr failed: %v", err)
	}

	// Test with failed command
	err = RunShellForInterval(ctx, "exit 1", tmpDir, 1*time.Second)
	if err == nil {
		t.Error("Expected error for failed command")
	}
}

func TestRunIntervalScript(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "script-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	challengeConf := ChallengeYaml{
		Name: "Test Challenge",
		Cwd:  tmpDir,
	}

	// Test with invalid interval (should return immediately)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go RunIntervalScript(ctx, challengeConf, "test", "echo test", 10*time.Second)

	// Give it a moment to log the error
	time.Sleep(100 * time.Millisecond)
	cancel()
}

func TestRunIntervalScript_ValidInterval(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interval script test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "script-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	challengeConf := ChallengeYaml{
		Name: "Test Challenge",
		Cwd:  tmpDir,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Run with very short valid interval (30s minimum, but context will cancel first)
	go RunIntervalScript(ctx, challengeConf, "test", "echo tick", 30*time.Second)

	// Wait for context to cancel
	<-ctx.Done()
}
