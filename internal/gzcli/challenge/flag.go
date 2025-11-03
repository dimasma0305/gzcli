package challenge

import (
	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
)

// IsFlagExist checks if a flag with a specific value exists in a list of Flag objects.
func IsFlagExist(flag string, flags []gzapi.Flag) bool {
	flagMap := make(map[string]struct{}, len(flags))
	for _, f := range flags {
		flagMap[f.Flag] = struct{}{}
	}
	_, exists := flagMap[flag]
	return exists
}

// UpdateChallengeFlags synchronizes the flags for a challenge between the local configuration
// and the remote server. It deletes flags that are no longer in the configuration and
// creates new flags that have been added.
func UpdateChallengeFlags(conf *config.Config, challengeConf config.ChallengeYaml, challengeData *gzapi.Challenge) error {
	for _, flag := range challengeData.Flags {
		if !IsExistInArray(flag.Flag, challengeConf.Flags) {
			flag.GameId = conf.Event.Id
			flag.ChallengeId = challengeData.Id
			flag.CS = conf.Event.CS
			if err := flag.Delete(); err != nil {
				return err
			}
		}
	}

	isCreatingNewFlag := false

	for _, flag := range challengeConf.Flags {
		if !IsFlagExist(flag, challengeData.Flags) {
			if err := challengeData.CreateFlag(gzapi.CreateFlagForm{
				Flag: flag,
			}); err != nil {
				return err
			}
			isCreatingNewFlag = true
		}
	}

	if isCreatingNewFlag {
		newChallData, err := challengeData.Refresh()
		if err != nil {
			return err
		}
		challengeData.Flags = newChallData.Flags
	}

	return nil
}
