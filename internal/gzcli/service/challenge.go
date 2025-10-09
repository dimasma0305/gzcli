package service

import (
	"fmt"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/repository"
)

// ChallengeService handles challenge business logic
type ChallengeService struct {
	challengeRepo  repository.ChallengeRepository
	cacheRepo      repository.CacheRepository
	attachmentRepo repository.AttachmentRepository
	flagRepo       repository.FlagRepository
	api            *gzapi.GZAPI
	gameID         int
}

// ChallengeServiceConfig holds configuration for ChallengeService
type ChallengeServiceConfig struct {
	ChallengeRepo  repository.ChallengeRepository
	Cache          repository.CacheRepository
	AttachmentRepo repository.AttachmentRepository
	FlagRepo       repository.FlagRepository
	API            *gzapi.GZAPI
	GameID         int
}

// NewChallengeService creates a new ChallengeService
func NewChallengeService(config ChallengeServiceConfig) *ChallengeService {
	return &ChallengeService{
		challengeRepo:  config.ChallengeRepo,
		cacheRepo:      config.Cache,
		attachmentRepo: config.AttachmentRepo,
		flagRepo:       config.FlagRepo,
		api:            config.API,
		gameID:         config.GameID,
	}
}

// Sync synchronizes a challenge with the API
func (s *ChallengeService) Sync(challengeConf config.ChallengeYaml) error {
	// This is a placeholder implementation
	// The actual implementation would need to be extracted from the existing challenge/sync.go
	return fmt.Errorf("Sync not implemented yet - needs to be extracted from challenge/sync.go")
}

// Create creates a new challenge
func (s *ChallengeService) Create(challengeConf config.ChallengeYaml) error {
	// This is a placeholder implementation
	// The actual implementation would need to be extracted from the existing challenge/sync.go
	return fmt.Errorf("Create not implemented yet - needs to be extracted from challenge/sync.go")
}

// Update updates an existing challenge
func (s *ChallengeService) Update(challengeConf config.ChallengeYaml) error {
	// This is a placeholder implementation
	// The actual implementation would need to be extracted from the existing challenge/sync.go
	return fmt.Errorf("Update not implemented yet - needs to be extracted from challenge/sync.go")
}

// Delete deletes a challenge
func (s *ChallengeService) Delete(challengeName string) error {
	// This is a placeholder implementation
	// The actual implementation would need to be extracted from the existing challenge/sync.go
	return fmt.Errorf("Delete not implemented yet - needs to be extracted from challenge/sync.go")
}

// List returns all challenges
func (s *ChallengeService) List() ([]gzapi.Challenge, error) {
	return s.challengeRepo.List()
}

// FindByName finds a challenge by name
func (s *ChallengeService) FindByName(name string) (*gzapi.Challenge, error) {
	return s.challengeRepo.FindByName(name)
}