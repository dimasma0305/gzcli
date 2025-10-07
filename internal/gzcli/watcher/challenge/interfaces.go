package challenge

import (
	"time"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
)

// ChallengeYaml interface represents a challenge configuration
type ChallengeYaml interface {
	GetName() string
	GetCwd() string
	GetScripts() map[string]ScriptValue
}

// ScriptValue interface for accessing script values
type ScriptValue interface {
	GetCommand() string
	HasInterval() bool
	GetInterval() time.Duration
}

// GZ interface for accessing GZCTF API
type GZ interface {
	GetAPI() *gzapi.GZAPI
}

// ConfigProvider provides configuration access
type ConfigProvider interface {
	GetConfig(api *gzapi.GZAPI) (Config, error)
	GetChallenges() ([]ChallengeYaml, error)
}

// Config interface for configuration operations
type Config interface {
	GetEvent() *gzapi.Game
	SetCS(api *gzapi.GZAPI)
}

// ScriptRunner interface for running scripts
type ScriptRunner interface {
	RunScriptWithIntervalSupport(challenge ChallengeYaml, scriptName string) error
	StopAllScriptsForChallenge(challengeName string)
}

// Logger interface for logging operations
type Logger interface {
	LogToDatabase(level, component, challenge, script, message, errorMsg string, duration int64)
	UpdateChallengeState(challengeName, status, errorMessage string, activeScripts map[string][]string)
}
