package service

import (
	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/repository"
	"github.com/dimasma0305/gzcli/internal/log"
)

// flagService implements FlagService
type flagService struct {
	flagRepo     repository.FlagRepository
	errorHandler *ErrorHandler
}

// NewFlagService creates a new flag service
func NewFlagService(flagRepo repository.FlagRepository) FlagService {
	return &flagService{
		flagRepo:     flagRepo,
		errorHandler: NewErrorHandler(),
	}
}

// UpdateFlags updates flags for a challenge
func (s *flagService) UpdateFlags(challengeConf config.ChallengeYaml, challenge *gzapi.Challenge) error {
	log.InfoH2("Updating flags for %s", challengeConf.Name)

	// Get existing flags
	existingFlags, err := s.flagRepo.List(challenge.Id)
	if err != nil {
		return s.errorHandler.Wrap(err, "getting existing flags")
	}

	// Delete flags that are no longer in the configuration
	for _, flag := range existingFlags {
		if !s.isFlagInConfig(flag.Flag, challengeConf.Flags) {
			if err := s.flagRepo.Delete(flag.Id); err != nil {
				return s.errorHandler.Wrap(err, "deleting flag")
			}
		}
	}

	// Create new flags that are in the configuration but not in the API
	isCreatingNewFlag := false
	for _, flag := range challengeConf.Flags {
		if !s.isFlagInAPI(flag, existingFlags) {
			if err := s.flagRepo.Create(challenge.Id, gzapi.CreateFlagForm{
				Flag: flag,
			}); err != nil {
				return s.errorHandler.Wrap(err, "creating flag")
			}
			isCreatingNewFlag = true
		}
	}

	if isCreatingNewFlag {
		log.InfoH2("New flags created for %s, refreshing challenge data", challengeConf.Name)
		// The challenge data would need to be refreshed here
		// This would need to be implemented based on the existing logic
	}

	log.InfoH2("Flags updated successfully for %s", challengeConf.Name)
	return nil
}

// Create creates a new flag
func (s *flagService) Create(challengeID int, flag gzapi.CreateFlagForm) error {
	return s.flagRepo.Create(challengeID, flag)
}

// Delete deletes a flag
func (s *flagService) Delete(flagID int) error {
	return s.flagRepo.Delete(flagID)
}

// List returns all flags for a challenge
func (s *flagService) List(challengeID int) ([]gzapi.Flag, error) {
	return s.flagRepo.List(challengeID)
}

// isFlagInConfig checks if a flag exists in the configuration
func (s *flagService) isFlagInConfig(flag string, configFlags []string) bool {
	for _, f := range configFlags {
		if f == flag {
			return true
		}
	}
	return false
}

// isFlagInAPI checks if a flag exists in the API flags
func (s *flagService) isFlagInAPI(flag string, apiFlags []gzapi.Flag) bool {
	for _, f := range apiFlags {
		if f.Flag == flag {
			return true
		}
	}
	return false
}