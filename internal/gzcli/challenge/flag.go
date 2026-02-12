package challenge

import (
	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
)

// IsFlagExist checks if a flag exists in the provided flags list
func IsFlagExist(flag string, flags []gzapi.Flag) bool {
	flagMap := make(map[string]struct{}, len(flags))
	for _, f := range flags {
		flagMap[f.Flag] = struct{}{}
	}
	_, exists := flagMap[flag]
	return exists
}

// UpdateChallengeFlags synchronizes challenge flags between configuration and API
func UpdateChallengeFlags(conf *config.Config, challengeConf config.ChallengeYaml, challengeData *gzapi.Challenge) error {
	mutated := false
	desiredFlags := make(map[string]struct{}, len(challengeConf.Flags))
	for _, flag := range challengeConf.Flags {
		desiredFlags[flag] = struct{}{}
	}

	existingFlags := make(map[string]gzapi.Flag, len(challengeData.Flags))
	for _, flag := range challengeData.Flags {
		existingFlags[flag.Flag] = flag
	}

	for _, flag := range challengeData.Flags {
		if _, keep := desiredFlags[flag.Flag]; !keep {
			flag.GameId = conf.Event.Id
			flag.ChallengeId = challengeData.Id
			flag.CS = conf.Event.CS
			if err := flag.Delete(); err != nil {
				return err
			}
			mutated = true
		}
	}

	toCreate := make([]gzapi.CreateFlagForm, 0, len(desiredFlags))
	for flag := range desiredFlags {
		if _, exists := existingFlags[flag]; !exists {
			toCreate = append(toCreate, gzapi.CreateFlagForm{Flag: flag})
		}
	}

	if len(toCreate) > 0 {
		if err := challengeData.CreateFlags(toCreate); err != nil {
			return err
		}
		mutated = true
	}

	if mutated {
		// Keep local state consistent without an extra GET /challenge refresh.
		newFlags := make([]gzapi.Flag, 0, len(desiredFlags))
		for _, desired := range challengeConf.Flags {
			if existing, ok := existingFlags[desired]; ok {
				newFlags = append(newFlags, existing)
				continue
			}
			newFlags = append(newFlags, gzapi.Flag{
				Flag:        desired,
				GameId:      conf.Event.Id,
				ChallengeId: challengeData.Id,
				CS:          conf.Event.CS,
			})
		}
		challengeData.Flags = newFlags
	}

	return nil
}
