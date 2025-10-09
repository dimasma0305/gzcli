package service

import (
	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
)

// ChallengeService defines operations for challenge management
type ChallengeService interface {
	// Sync synchronizes a challenge with the API
	Sync(challengeConf config.ChallengeYaml) error
	
	// Create creates a new challenge
	Create(challengeConf config.ChallengeYaml) (*gzapi.Challenge, error)
	
	// Update updates an existing challenge
	Update(challengeConf config.ChallengeYaml) (*gzapi.Challenge, error)
	
	// Delete deletes a challenge
	Delete(name string) error
	
	// FindByName finds a challenge by name
	FindByName(name string) (*gzapi.Challenge, error)
	
	// List returns all challenges
	List() ([]gzapi.Challenge, error)
}

// AttachmentService defines operations for attachment management
type AttachmentService interface {
	// HandleAttachments handles attachments for a challenge
	HandleAttachments(challengeConf config.ChallengeYaml, challenge *gzapi.Challenge) error
	
	// Create creates a new attachment
	Create(challengeID int, attachment gzapi.CreateAttachmentForm) error
	
	// Delete deletes an attachment
	Delete(challengeID int, attachmentID int) error
	
	// List returns all attachments for a challenge
	List(challengeID int) ([]gzapi.Attachment, error)
}

// FlagService defines operations for flag management
type FlagService interface {
	// UpdateFlags updates flags for a challenge
	UpdateFlags(challengeConf config.ChallengeYaml, challenge *gzapi.Challenge) error
	
	// Create creates a new flag
	Create(challengeID int, flag gzapi.CreateFlagForm) error
	
	// Delete deletes a flag
	Delete(flagID int) error
	
	// List returns all flags for a challenge
	List(challengeID int) ([]gzapi.Flag, error)
}

// GameService defines operations for game management
type GameService interface {
	// FindGame finds a game by title
	FindGame(games []*gzapi.Game, title string) *gzapi.Game
	
	// CreateGame creates a new game
	CreateGame(conf *config.Config) (*gzapi.Game, error)
	
	// UpdateGameIfNeeded updates a game if needed
	UpdateGameIfNeeded(conf *config.Config, currentGame *gzapi.Game, api *gzapi.GZAPI, createPosterFunc func(string, *gzapi.Game, *gzapi.GZAPI) (string, error), setCache func(string, interface{}) error) error
	
	// List returns all games
	List() ([]*gzapi.Game, error)
}