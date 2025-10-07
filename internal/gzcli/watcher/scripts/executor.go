package scripts

import (
	"context"
	"time"

	"github.com/dimasma0305/gzcli/internal/gzcli/challenge"
)

// DefaultScriptTimeout is the default timeout for script execution
const DefaultScriptTimeout = challenge.DefaultScriptTimeout

// ValidateInterval validates a script interval duration
func ValidateInterval(interval time.Duration, scriptName string) bool {
	return challenge.ValidateInterval(interval, scriptName)
}

// RunShellForInterval runs a shell script with a given interval context
func RunShellForInterval(ctx context.Context, script string, cwd string, timeout time.Duration) error {
	return challenge.RunShellForInterval(ctx, script, cwd, timeout)
}

// RunShellWithContext runs a shell script with context
func RunShellWithContext(ctx context.Context, script string, cwd string) error {
	return challenge.RunShellWithContext(ctx, script, cwd)
}
