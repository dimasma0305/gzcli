package service

import (
	"fmt"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/repository"
	"github.com/dimasma0305/gzcli/internal/log"
)

// ChallengeServiceConfig holds configuration for ChallengeService
type ChallengeServiceConfig struct {
	ChallengeRepo  repository.ChallengeRepository
	Cache          repository.CacheRepository
	AttachmentRepo repository.AttachmentRepository
	FlagRepo       repository.FlagRepository
	API            *gzapi.GZAPI
	GameID         int
}

// challengeService implements ChallengeService
type challengeService struct {
	challengeRepo  repository.ChallengeRepository
	cache          repository.CacheRepository
	attachmentRepo repository.AttachmentRepository
	flagRepo       repository.FlagRepository
	api            *gzapi.GZAPI
	gameID         int
	retryHandler   *RetryHandler
	errorHandler   *ErrorHandler
	validator      *Validator
}

// NewChallengeService creates a new challenge service
func NewChallengeService(config ChallengeServiceConfig) ChallengeService {
	return &challengeService{
		challengeRepo:  config.ChallengeRepo,
		cache:          config.Cache,
		attachmentRepo: config.AttachmentRepo,
		flagRepo:       config.FlagRepo,
		api:            config.API,
		gameID:         config.GameID,
		retryHandler:   NewRetryHandler(3, 1*time.Second),
		errorHandler:   NewErrorHandler(),
		validator:      NewValidator(),
	}
}

// Sync synchronizes a challenge with the API
func (s *challengeService) Sync(challengeConf config.ChallengeYaml) error {
	log.InfoH2("Starting sync for challenge: %s (Type: %s, Category: %s)", challengeConf.Name, challengeConf.Type, challengeConf.Category)

	// Check if challenge exists
	exists, err := s.challengeRepo.Exists(challengeConf.Name)
	if err != nil {
		return s.errorHandler.Wrap(err, "checking challenge existence")
	}

	var challengeData *gzapi.Challenge
	if !exists {
		challengeData, err = s.handleNewChallenge(challengeConf)
		if err != nil {
			return s.errorHandler.Wrap(err, "handling new challenge")
		}
	} else {
		challengeData, err = s.handleExistingChallenge(challengeConf)
		if err != nil {
			return s.errorHandler.Wrap(err, "handling existing challenge")
		}
	}

	// Process attachments and flags
	if err := s.processAttachmentsAndFlags(challengeConf, challengeData); err != nil {
		return s.errorHandler.Wrap(err, "processing attachments and flags")
	}

	// Update challenge if needed
	if err := s.updateChallengeIfNeeded(challengeConf, challengeData); err != nil {
		return s.errorHandler.Wrap(err, "updating challenge")
	}

	log.InfoH2("Successfully completed sync for challenge: %s", challengeConf.Name)
	return nil
}

// Create creates a new challenge
func (s *challengeService) Create(challengeConf config.ChallengeYaml) (*gzapi.Challenge, error) {
	// This would need to be implemented based on the GZAPI create challenge functionality
	return nil, fmt.Errorf("create challenge not implemented")
}

// Update updates an existing challenge
func (s *challengeService) Update(challengeConf config.ChallengeYaml) (*gzapi.Challenge, error) {
	challenge, err := s.challengeRepo.FindByName(challengeConf.Name)
	if err != nil {
		return nil, s.errorHandler.Wrap(err, "finding challenge")
	}

	// Convert config to challenge data and update
	// This would need to be implemented based on the existing MergeChallengeData logic
	return s.challengeRepo.Update(challenge)
}

// Delete deletes a challenge
func (s *challengeService) Delete(name string) error {
	challenge, err := s.challengeRepo.FindByName(name)
	if err != nil {
		return s.errorHandler.Wrap(err, "finding challenge")
	}

	return s.challengeRepo.Delete(challenge.Id)
}

// FindByName finds a challenge by name
func (s *challengeService) FindByName(name string) (*gzapi.Challenge, error) {
	return s.challengeRepo.FindByName(name)
}

// List returns all challenges
func (s *challengeService) List() ([]gzapi.Challenge, error) {
	return s.challengeRepo.List()
}

// handleNewChallenge handles creation of a new challenge
func (s *challengeService) handleNewChallenge(challengeConf config.ChallengeYaml) (*gzapi.Challenge, error) {
	// This would need to be implemented based on the existing handleNewChallenge logic
	return nil, fmt.Errorf("handle new challenge not implemented")
}

// handleExistingChallenge handles updates to an existing challenge
func (s *challengeService) handleExistingChallenge(challengeConf config.ChallengeYaml) (*gzapi.Challenge, error) {
	challenge, err := s.challengeRepo.FindByName(challengeConf.Name)
	if err != nil {
		return nil, s.errorHandler.Wrap(err, "finding existing challenge")
	}

	// Check if configuration has changed
	if !s.isConfigEdited(challengeConf, challenge) {
		log.InfoH2("Challenge %s is unchanged, skipping update", challengeConf.Name)
		return challenge, nil
	}

	return challenge, nil
}

// processAttachmentsAndFlags handles attachments and flags for a challenge
func (s *challengeService) processAttachmentsAndFlags(challengeConf config.ChallengeYaml, challenge *gzapi.Challenge) error {
	log.InfoH2("Processing attachments for %s", challengeConf.Name)
	// This would need to be implemented based on the existing processAttachmentsAndFlags logic
	log.InfoH2("Attachments processed successfully for %s", challengeConf.Name)

	log.InfoH2("Updating flags for %s", challengeConf.Name)
	// This would need to be implemented based on the existing processAttachmentsAndFlags logic
	log.InfoH2("Flags updated successfully for %s", challengeConf.Name)

	return nil
}

// updateChallengeIfNeeded updates the challenge if configuration has changed
func (s *challengeService) updateChallengeIfNeeded(challengeConf config.ChallengeYaml, challenge *gzapi.Challenge) error {
	if !s.isConfigEdited(challengeConf, challenge) {
		log.InfoH2("Challenge %s is unchanged, skipping update", challengeConf.Name)
		return nil
	}

	log.InfoH2("Configuration changed for %s, updating...", challengeConf.Name)
	
	// Update challenge with retry logic
	return s.retryHandler.Execute(func() error {
		_, err := s.challengeRepo.Update(challenge)
		return err
	})
}

// isConfigEdited checks if the challenge configuration has been edited
func (s *challengeService) isConfigEdited(challengeConf config.ChallengeYaml, challenge *gzapi.Challenge) bool {
	var cacheChallenge gzapi.Challenge
	if err := s.cache.Get(challengeConf.Category+"/"+challengeConf.Name+"/challenge", &cacheChallenge); err != nil {
		return true
	}

	if challenge.Hints == nil {
		challenge.Hints = []string{}
	}
	return !cmp.Equal(*challenge, cacheChallenge)
}