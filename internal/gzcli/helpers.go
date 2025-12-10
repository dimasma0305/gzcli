package gzcli

import (
	"github.com/dimasma0305/gzcli/internal/gzcli/challenge"
	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/script"
	"github.com/dimasma0305/gzcli/internal/gzcli/structure"
)

// Wrapper functions to bridge between old and new package structures

func deleteCacheWrapper(key string) {
	_ = DeleteCache(key)
}

func getConfigWrapper(api *gzapi.GZAPI) (*config.Config, error) {
	return config.GetConfig(api, GetCache, setCache, deleteCacheWrapper, createNewGameWrapper)
}

func createNewGameWrapper(conf *config.Config, api *gzapi.GZAPI) (*gzapi.Game, error) {
	return challenge.CreateNewGame(conf, api, createPosterIfNotExistOrDifferent, setCache)
}

func genStructureWrapper(challenges []interface{ GetCwd() string }) error {
	// Convert to structure.ChallengeData
	converted := make([]structure.ChallengeData, len(challenges))
	for i, c := range challenges {
		converted[i] = c
	}
	return structure.GenerateStructure(converted)
}

// RunScripts executes scripts for all challenges using a worker pool
// Returns any per-challenge failures for summary reporting.
func RunScripts(scriptName string, eventName string) ([]script.ScriptFailure, error) {
	// Get config for the specific event
	configPkg, err := config.GetConfigWithEvent(&gzapi.GZAPI{}, eventName, GetCache, setCache, deleteCacheWrapper, createNewGameWrapper)
	if err != nil {
		return nil, err
	}

	challengesConf, err := config.GetChallengesYaml(configPkg)
	if err != nil {
		return nil, err
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

	failures, err := script.RunScripts(scriptName, challengeInterfaces, func(conf script.ChallengeConf, script string) error {
		adapter := conf.(challengeConfAdapter)
		// Pass config.ChallengeYaml directly - challenge package now uses this type
		return challenge.RunScript(adapter.c, script)
	})

	return failures, err
}

// Adapter types
type challengeConfAdapter struct {
	c config.ChallengeYaml
}

func (c challengeConfAdapter) GetName() string {
	return c.c.Name
}

func (c challengeConfAdapter) GetScripts() map[string]script.ScriptValue {
	result := make(map[string]script.ScriptValue)
	for k, v := range c.c.Scripts {
		result[k] = scriptValueAdapter{v}
	}
	return result
}

type scriptValueAdapter struct {
	v config.ScriptValue
}

func (s scriptValueAdapter) GetCommand() string {
	return s.v.GetCommand()
}
