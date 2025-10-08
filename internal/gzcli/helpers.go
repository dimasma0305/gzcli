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
	// Convert config.Config to challenge.Config for the challenge package
	challengeConfig := &challenge.Config{
		Url:   conf.Url,
		Creds: conf.Creds,
		Event: conf.Event,
	}
	return challenge.CreateNewGame(challengeConfig, api, createPosterIfNotExistOrDifferent, setCache)
}

func findCurrentGameWrapper(games []*gzapi.Game, title string, api *gzapi.GZAPI) *gzapi.Game {
	return challenge.FindCurrentGame(games, title, api)
}

func updateGameIfNeededWrapper(conf *config.Config, currentGame *gzapi.Game, api *gzapi.GZAPI) error {
	challengeConfig := &challenge.Config{
		Url:   conf.Url,
		Creds: conf.Creds,
		Event: conf.Event,
	}
	return challenge.UpdateGameIfNeeded(challengeConfig, currentGame, api, createPosterIfNotExistOrDifferent, setCache)
}

func validateChallengesWrapper(challengesConf []config.ChallengeYaml) error {
	// Convert config.ChallengeYaml to challenge.ChallengeYaml
	challenges := make([]challenge.ChallengeYaml, len(challengesConf))
	for i, c := range challengesConf {
		challenges[i] = challenge.ChallengeYaml{
			Name:        c.Name,
			Author:      c.Author,
			Description: c.Description,
			Flags:       c.Flags,
			Value:       c.Value,
			Provide:     c.Provide,
			Visible:     c.Visible,
			Type:        c.Type,
			Hints:       c.Hints,
			Container: challenge.Container{
				FlagTemplate:         c.Container.FlagTemplate,
				ContainerImage:       c.Container.ContainerImage,
				MemoryLimit:          c.Container.MemoryLimit,
				CpuCount:             c.Container.CpuCount,
				StorageLimit:         c.Container.StorageLimit,
				ContainerExposePort:  c.Container.ContainerExposePort,
				EnableTrafficCapture: c.Container.EnableTrafficCapture,
			},
			Scripts:   convertScripts(c.Scripts),
			Dashboard: convertDashboard(c.Dashboard),
			Category:  c.Category,
			Cwd:       c.Cwd,
		}
	}
	return challenge.ValidateChallenges(challenges)
}

func syncChallengeWrapper(conf *config.Config, challengeConf config.ChallengeYaml, challenges []gzapi.Challenge, api *gzapi.GZAPI) error {
	// Convert to challenge package types
	challengeConfig := &challenge.Config{
		Url:   conf.Url,
		Creds: conf.Creds,
		Event: conf.Event,
	}

	challYaml := challenge.ChallengeYaml{
		Name:        challengeConf.Name,
		Author:      challengeConf.Author,
		Description: challengeConf.Description,
		Flags:       challengeConf.Flags,
		Value:       challengeConf.Value,
		Provide:     challengeConf.Provide,
		Visible:     challengeConf.Visible,
		Type:        challengeConf.Type,
		Hints:       challengeConf.Hints,
		Container: challenge.Container{
			FlagTemplate:         challengeConf.Container.FlagTemplate,
			ContainerImage:       challengeConf.Container.ContainerImage,
			MemoryLimit:          challengeConf.Container.MemoryLimit,
			CpuCount:             challengeConf.Container.CpuCount,
			StorageLimit:         challengeConf.Container.StorageLimit,
			ContainerExposePort:  challengeConf.Container.ContainerExposePort,
			EnableTrafficCapture: challengeConf.Container.EnableTrafficCapture,
		},
		Scripts:   convertScripts(challengeConf.Scripts),
		Dashboard: convertDashboard(challengeConf.Dashboard),
		Category:  challengeConf.Category,
		Cwd:       challengeConf.Cwd,
	}

	return challenge.SyncChallenge(challengeConfig, challYaml, challenges, api, GetCache, setCache)
}

func convertScripts(scripts map[string]config.ScriptValue) map[string]challenge.ScriptValue {
	result := make(map[string]challenge.ScriptValue)
	for k, v := range scripts {
		result[k] = challenge.ScriptValue{
			Simple: v.Simple,
			Complex: func() *challenge.ScriptConfig {
				if v.Complex != nil {
					return &challenge.ScriptConfig{
						Execute:  v.Complex.Execute,
						Interval: v.Complex.Interval,
					}
				}
				return nil
			}(),
		}
	}
	return result
}

func convertDashboard(dashboard *config.Dashboard) *challenge.Dashboard {
	if dashboard == nil {
		return nil
	}
	return &challenge.Dashboard{
		Compose:                  dashboard.Compose,
		ChallengeDurationMinutes: dashboard.ChallengeDurationMinutes,
		ResetTimerMinutes:        dashboard.ResetTimerMinutes,
		RestartCooldownMinutes:   dashboard.RestartCooldownMinutes,
	}
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
		challYaml := challenge.ChallengeYaml{
			Name:        adapter.c.Name,
			Author:      adapter.c.Author,
			Description: adapter.c.Description,
			Flags:       adapter.c.Flags,
			Value:       adapter.c.Value,
			Provide:     adapter.c.Provide,
			Visible:     adapter.c.Visible,
			Type:        adapter.c.Type,
			Hints:       adapter.c.Hints,
			Container: challenge.Container{
				FlagTemplate:         adapter.c.Container.FlagTemplate,
				ContainerImage:       adapter.c.Container.ContainerImage,
				MemoryLimit:          adapter.c.Container.MemoryLimit,
				CpuCount:             adapter.c.Container.CpuCount,
				StorageLimit:         adapter.c.Container.StorageLimit,
				ContainerExposePort:  adapter.c.Container.ContainerExposePort,
				EnableTrafficCapture: adapter.c.Container.EnableTrafficCapture,
			},
			Scripts:   convertScripts(adapter.c.Scripts),
			Dashboard: convertDashboard(adapter.c.Dashboard),
			Category:  adapter.c.Category,
			Cwd:       adapter.c.Cwd,
		}
		return challenge.RunScript(challYaml, script)
	})
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
