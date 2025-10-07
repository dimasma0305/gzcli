package challenge

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/dimasma0305/gzcli/internal/log"
)

// Script execution timeout constants
const (
	DefaultScriptTimeout = 5 * time.Minute
	MaxScriptTimeout     = 30 * time.Minute
)

var (
	shell     string
	shellOnce sync.Once
)

// getShell returns the shell to use for script execution in a thread-safe way
func getShell() string {
	shellOnce.Do(func() {
		shell = os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}
	})
	return shell
}

// RunScript executes a specified script for a challenge
func RunScript(challengeConf ChallengeYaml, script string) error {
	scriptValue, exists := challengeConf.Scripts[script]
	if !exists {
		return nil
	}

	if challengeConf.Dashboard != nil {
		return nil
	}

	command := scriptValue.GetCommand()
	if command == "" {
		return nil
	}

	// Check if script has an interval configured
	if scriptValue.HasInterval() {
		interval := scriptValue.GetInterval()
		log.InfoH2("Warning: Interval script '%s' with interval %v detected", script, interval)
		log.InfoH3("Interval scripts are only supported when using the watcher. Running once instead.")
		log.InfoH3("Script command: %s", command)
	}

	// Run simple one-time script
	log.InfoH2("Running:\n%s", command)
	return runShell(command, challengeConf.Cwd)
}

//nolint:gosec // G204: Script execution is the intended purpose of this function
func runShell(script string, cwd string) error {
	cmd := exec.Command(getShell(), "-c", script)
	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunShellWithContext executes a shell command with context cancellation support
//
//nolint:gosec // G204: Script execution is the intended purpose of this function
func RunShellWithContext(ctx context.Context, script string, cwd string) error {
	cmd := exec.CommandContext(ctx, getShell(), "-c", script)
	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunShellWithTimeout executes a shell command with timeout protection
func RunShellWithTimeout(ctx context.Context, script string, cwd string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = DefaultScriptTimeout
	}
	if timeout > MaxScriptTimeout {
		timeout = MaxScriptTimeout
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return RunShellWithContext(timeoutCtx, script, cwd)
}

// RunShellForInterval executes a shell command for interval scripts with proper output management
func RunShellForInterval(ctx context.Context, script string, cwd string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = DefaultScriptTimeout
	}
	if timeout > MaxScriptTimeout {
		timeout = MaxScriptTimeout
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	//nolint:gosec // G204: Script execution is the intended purpose of this function
	cmd := exec.CommandContext(timeoutCtx, getShell(), "-c", script)
	cmd.Dir = cwd

	// For interval scripts, capture output for logging instead of stdout
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Log output if present
	if stdout.Len() > 0 {
		output := strings.TrimSpace(stdout.String())
		if output != "" {
			log.InfoH3("Script output: %s", output)
		}
	}
	if stderr.Len() > 0 {
		errOutput := strings.TrimSpace(stderr.String())
		if errOutput != "" {
			log.Error("Script error output: %s", errOutput)
		}
	}

	return err
}

// RunIntervalScript executes a script at regular intervals with context cancellation
func RunIntervalScript(ctx context.Context, challengeConf ChallengeYaml, scriptName, command string, interval time.Duration) {
	// Validate interval
	if !ValidateInterval(interval, scriptName) {
		log.Error("Invalid interval for script '%s' in challenge '%s', skipping", scriptName, challengeConf.Name)
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.InfoH3("Started interval script '%s' for challenge '%s' with interval %v", scriptName, challengeConf.Name, interval)

	for {
		select {
		case <-ctx.Done():
			log.InfoH3("Stopped interval script '%s' for challenge '%s' (context cancelled)", scriptName, challengeConf.Name)
			return
		case <-ticker.C:
			log.InfoH3("Executing interval script '%s' for challenge '%s'", scriptName, challengeConf.Name)

			// Use context-aware execution with proper timeout and output handling
			start := time.Now()
			if err := RunShellForInterval(ctx, command, challengeConf.Cwd, DefaultScriptTimeout); err != nil {
				duration := time.Since(start)
				log.Error("Interval script '%s' failed for challenge '%s' after %v: %v", scriptName, challengeConf.Name, duration, err)
			} else {
				duration := time.Since(start)
				log.InfoH3("Interval script '%s' completed successfully for challenge '%s' in %v", scriptName, challengeConf.Name, duration)
			}
		}
	}
}
