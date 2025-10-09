package service

import (
	"fmt"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/repository"
)

// FlagService handles flag business logic
type FlagService struct {
	flagRepo      repository.FlagRepository
	challengeRepo repository.ChallengeRepository
}

// FlagServiceConfig holds configuration for FlagService
type FlagServiceConfig struct {
	FlagRepo      repository.FlagRepository
	ChallengeRepo repository.ChallengeRepository
}

// NewFlagService creates a new FlagService
func NewFlagService(config FlagServiceConfig) *FlagService {
	return &FlagService{
		flagRepo:      config.FlagRepo,
		challengeRepo: config.ChallengeRepo,
	}
}

// Create creates a new flag
func (s *FlagService) Create(challengeID int, flag gzapi.CreateFlagForm) error {
	return s.flagRepo.Create(challengeID, flag)
}

// Delete deletes a flag
func (s *FlagService) Delete(challengeID int, flagID int) error {
	return s.flagRepo.Delete(challengeID, flagID)
}

// List returns all flags for a challenge
func (s *FlagService) List(challengeID int) ([]gzapi.Flag, error) {
	return s.flagRepo.List(challengeID)
}