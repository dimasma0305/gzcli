package gzcli

import (
	"github.com/dimasma0305/gzcli/internal/gzcli/challenge"
	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/script"
	"github.com/dimasma0305/gzcli/internal/gzcli/structure"
)

// deleteCacheWrapper provides a simplified interface for deleting cache entries,
// intended for use as a callback in other packages.
func deleteCacheWrapper(key string) {
	_ = DeleteCache(key)
}

// getConfigWrapper retrieves the application configuration, injecting the necessary cache
// and game creation functions as dependencies.
func getConfigWrapper(api *gzapi.GZAPI) (*config.Config, error) {
	return config.GetConfig(api, GetCache, setCache, deleteCacheWrapper, createNewGameWrapper)
}

// createNewGameWrapper is a wrapper around the challenge.CreateNewGame function,
// providing the necessary dependencies for poster creation and caching.
func createNewGameWrapper(conf *config.Config, api *gzapi.GZAPI) (*gzapi.Game, error) {
	return challenge.CreateNewGame(conf, api, createPosterIfNotExistOrDifferent, setCache)
}

// genStructureWrapper adapts the challenge data and calls the structure generation logic.
func genStructureWrapper(challenges []interface{ GetCwd() string }) error {
	// Convert to structure.ChallengeData
	converted := make([]structure.ChallengeData, len(challenges))
	for i, c := range challenges {
		converted[i] = c
	}
	return structure.GenerateStructure(converted)
}

// RunScripts executes a named script for all challenges in a given event.
// It uses a worker pool for parallel execution of the scripts.
func RunScripts(scriptName string, eventName string) error {
	// Get config for the specific event
	configPkg, err := config.GetConfigWithEvent(&gzapi.GZAPI{}, eventName, GetCache, setCache, deleteCacheWrapper, createNewGameWrapper)
	if err != nil {
		return err
	}

	challengesConf, err := config.GetChallengesYaml(configPkg)
	if err != nil {
		return err
	}

	// Convert to interface for script package
	challenges := make([]challengeConfAdapter, len(challengesConf))
	for i, c := range challengesConf {
		challenges[i] = challengeConfAdapter{c}
	}

	challengeInterfaces := make([]script.ChallengeConf, len(challenges))
	for i := range challenges {
		challengeInterfaces[i] = challenges[i]
	}

	return script.RunScripts(scriptName, challengeInterfaces, func(conf script.ChallengeConf, script string) error {
		adapter := conf.(challengeConfAdapter)
		// Pass config.ChallengeYaml directly - challenge package now uses this type
		return challenge.RunScript(adapter.c, script)
	})
}

// challengeConfAdapter adapts the config.ChallengeYaml type to the script.ChallengeConf interface.
type challengeConfAdapter struct {
	c config.ChallengeYaml
}

// GetName returns the name of the challenge.
func (c challengeConfAdapter) GetName() string {
	return c.c.Name
}

// GetScripts returns the scripts defined for the challenge, adapted to the script.ScriptValue interface.
func (c challengeConfAdapter) GetScripts() map[string]script.ScriptValue {
	result := make(map[string]script.ScriptValue)
	for k, v := range c.c.Scripts {
		result[k] = scriptValueAdapter{v}
	}
	return result
}

// scriptValueAdapter adapts the config.ScriptValue type to the script.ScriptValue interface.
type scriptValueAdapter struct {
	v config.ScriptValue
}

// GetCommand returns the command string for the script.
func (s scriptValueAdapter) GetCommand() string {
	return s.v.GetCommand()
}
