package service

import (
	"context"
	"fmt"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/repository"
	"github.com/dimasma0305/gzcli/internal/gzcli/errors"
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

// ChallengeService handles challenge business logic
type ChallengeService struct {
	challengeRepo  repository.ChallengeRepository
	cache          repository.CacheRepository
	attachmentRepo repository.AttachmentRepository
	flagRepo       repository.FlagRepository
	api            *gzapi.GZAPI
	gameID         int
}

// NewChallengeService creates a new ChallengeService
func NewChallengeService(config ChallengeServiceConfig) *ChallengeService {
	return &ChallengeService{
		challengeRepo:  config.ChallengeRepo,
		cache:          config.Cache,
		attachmentRepo: config.AttachmentRepo,
		flagRepo:       config.FlagRepo,
		api:            config.API,
		gameID:         config.GameID,
	}
}

// Sync synchronizes a challenge configuration with the API
func (s *ChallengeService) Sync(ctx context.Context, challengeConf config.ChallengeYaml) error {
	log.Info("Syncing challenge: %s", challengeConf.Name)

	// Get existing challenges
	challenges, err := s.challengeRepo.GetChallenges(ctx, s.gameID)
	if err != nil {
		return errors.Wrap(err, "failed to get existing challenges")
	}

	// Check if challenge exists
	existingChallenge := s.findChallengeByName(challenges, challengeConf.Name)
	
	if existingChallenge == nil {
		// Create new challenge
		return s.createChallenge(ctx, challengeConf)
	}

	// Update existing challenge
	return s.updateChallenge(ctx, challengeConf, existingChallenge)
}

// createChallenge creates a new challenge
func (s *ChallengeService) createChallenge(ctx context.Context, challengeConf config.ChallengeYaml) error {
	log.Info("Creating new challenge: %s", challengeConf.Name)

	// Convert config to API challenge
	challenge := s.convertConfigToChallenge(challengeConf)

	// Create challenge via API
	createdChallenge, err := s.challengeRepo.CreateChallenge(ctx, s.gameID, challenge)
	if err != nil {
		return errors.Wrapf(err, "failed to create challenge %s", challengeConf.Name)
	}

	// Handle attachments
	if err := s.syncAttachments(ctx, createdChallenge.Id, challengeConf); err != nil {
		log.Error("Failed to sync attachments for %s: %v", challengeConf.Name, err)
		// Don't fail the entire sync for attachment errors
	}

	// Handle flags
	if err := s.syncFlags(ctx, createdChallenge.Id, challengeConf); err != nil {
		log.Error("Failed to sync flags for %s: %v", challengeConf.Name, err)
		// Don't fail the entire sync for flag errors
	}

	// Cache the challenge
	cacheKey := fmt.Sprintf("%s/%s/challenge", challengeConf.Category, challengeConf.Name)
	if err := s.cache.Set(ctx, cacheKey, *createdChallenge); err != nil {
		log.Warn("Failed to cache challenge %s: %v", challengeConf.Name, err)
	}

	log.Info("Successfully created challenge: %s", challengeConf.Name)
	return nil
}

// updateChallenge updates an existing challenge
func (s *ChallengeService) updateChallenge(ctx context.Context, challengeConf config.ChallengeYaml, existingChallenge *gzapi.Challenge) error {
	log.Info("Updating existing challenge: %s", challengeConf.Name)

	// Check if challenge needs updating
	if !s.needsUpdate(challengeConf, existingChallenge) {
		log.Info("Challenge %s is up to date", challengeConf.Name)
		return nil
	}

	// Convert config to API challenge
	updatedChallenge := s.convertConfigToChallenge(challengeConf)
	updatedChallenge.Id = existingChallenge.Id

	// Update challenge via API
	_, err := s.challengeRepo.UpdateChallenge(ctx, s.gameID, updatedChallenge)
	if err != nil {
		return errors.Wrapf(err, "failed to update challenge %s", challengeConf.Name)
	}

	// Handle attachments
	if err := s.syncAttachments(ctx, existingChallenge.Id, challengeConf); err != nil {
		log.Error("Failed to sync attachments for %s: %v", challengeConf.Name, err)
	}

	// Handle flags
	if err := s.syncFlags(ctx, existingChallenge.Id, challengeConf); err != nil {
		log.Error("Failed to sync flags for %s: %v", challengeConf.Name, err)
	}

	// Cache the updated challenge
	cacheKey := fmt.Sprintf("%s/%s/challenge", challengeConf.Category, challengeConf.Name)
	if err := s.cache.Set(ctx, cacheKey, *updatedChallenge); err != nil {
		log.Warn("Failed to cache updated challenge %s: %v", challengeConf.Name, err)
	}

	log.Info("Successfully updated challenge: %s", challengeConf.Name)
	return nil
}

// findChallengeByName finds a challenge by name in the list
func (s *ChallengeService) findChallengeByName(challenges []gzapi.Challenge, name string) *gzapi.Challenge {
	for _, challenge := range challenges {
		if challenge.Title == name {
			return &challenge
		}
	}
	return nil
}

// convertConfigToChallenge converts a ChallengeYaml to gzapi.Challenge
func (s *ChallengeService) convertConfigToChallenge(conf config.ChallengeYaml) *gzapi.Challenge {
	challenge := &gzapi.Challenge{
		Title:       conf.Name,
		Category:    conf.Category,
		Description: conf.Description,
		Value:       conf.Value,
		Type:        conf.Type,
		State:       "visible",
	}

	// Set hints if provided
	if len(conf.Hints) > 0 {
		challenge.Hints = conf.Hints
	}

	// Set container configuration if provided
	if conf.Container.Image != "" {
		challenge.ContainerImage = conf.Container.Image
		challenge.MemoryLimit = conf.Container.MemoryLimit
		challenge.CPUCount = conf.Container.CPUCount
	}

	return challenge
}

// needsUpdate checks if a challenge needs updating
func (s *ChallengeService) needsUpdate(conf config.ChallengeYaml, existing *gzapi.Challenge) bool {
	// This is a simplified check - in practice, you'd want more sophisticated comparison
	return conf.Description != existing.Description ||
		conf.Value != existing.Value ||
		conf.Type != existing.Type
}

// syncAttachments handles attachment synchronization
func (s *ChallengeService) syncAttachments(ctx context.Context, challengeID int, conf config.ChallengeYaml) error {
	if s.attachmentRepo == nil {
		log.Debug("Attachment repository not available, skipping attachment sync")
		return nil
	}

	// Get existing attachments
	existingAttachments, err := s.attachmentRepo.GetAttachments(ctx, challengeID)
	if err != nil {
		log.Warn("Failed to get existing attachments for challenge %d: %v", challengeID, err)
		// Don't fail the entire sync for attachment errors
		return nil
	}

	// Create a map of existing attachments by filename
	existingMap := make(map[string]*gzapi.Attachment)
	for i := range existingAttachments {
		existingMap[existingAttachments[i].FileName] = &existingAttachments[i]
	}

	// Process attachments from configuration
	for _, attachmentPath := range conf.Attachments {
		// Check if attachment already exists
		if _, exists := existingMap[attachmentPath]; exists {
			log.Debug("Attachment %s already exists for challenge %d", attachmentPath, challengeID)
			continue
		}

		// Upload new attachment
		_, err := s.attachmentRepo.UploadAttachment(ctx, challengeID, attachmentPath)
		if err != nil {
			log.Error("Failed to upload attachment %s for challenge %d: %v", attachmentPath, challengeID, err)
			// Continue with other attachments
			continue
		}

		log.Info("Successfully uploaded attachment %s for challenge %d", attachmentPath, challengeID)
	}

	return nil
}

// syncFlags handles flag synchronization
func (s *ChallengeService) syncFlags(ctx context.Context, challengeID int, conf config.ChallengeYaml) error {
	if s.flagRepo == nil {
		log.Debug("Flag repository not available, skipping flag sync")
		return nil
	}

	// Get existing flags
	existingFlags, err := s.flagRepo.GetFlags(ctx, challengeID)
	if err != nil {
		log.Warn("Failed to get existing flags for challenge %d: %v", challengeID, err)
		// Don't fail the entire sync for flag errors
		return nil
	}

	// Create a map of existing flags by value
	existingMap := make(map[string]*gzapi.Flag)
	for i := range existingFlags {
		existingMap[existingFlags[i].Value] = &existingFlags[i]
	}

	// Process flags from configuration
	for _, flagValue := range conf.Flags {
		// Check if flag already exists
		if _, exists := existingMap[flagValue]; exists {
			log.Debug("Flag %s already exists for challenge %d", flagValue, challengeID)
			continue
		}

		// Create new flag
		flag := &gzapi.Flag{
			ChallengeID: challengeID,
			Value:       flagValue,
			Type:        "static", // Default type for static flags
		}

		_, err := s.flagRepo.CreateFlag(ctx, challengeID, flag)
		if err != nil {
			log.Error("Failed to create flag %s for challenge %d: %v", flagValue, challengeID, err)
			// Continue with other flags
			continue
		}

		log.Info("Successfully created flag %s for challenge %d", flagValue, challengeID)
	}

	// Handle dynamic flag template for dynamic container challenges
	if conf.Type == "DynamicContainer" && conf.Container.FlagTemplate != "" {
		// Check if dynamic flag template already exists
		if _, exists := existingMap[conf.Container.FlagTemplate]; !exists {
			flag := &gzapi.Flag{
				ChallengeID: challengeID,
				Value:       conf.Container.FlagTemplate,
				Type:        "dynamic",
			}

			_, err := s.flagRepo.CreateFlag(ctx, challengeID, flag)
			if err != nil {
				log.Error("Failed to create dynamic flag template for challenge %d: %v", challengeID, err)
			} else {
				log.Info("Successfully created dynamic flag template for challenge %d", challengeID)
			}
		}
	}

	return nil
}