package repository

import (
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
)

// ChallengeRepository defines operations for challenge data access
type ChallengeRepository interface {
	// FindByName finds a challenge by its name
	FindByName(name string) (*gzapi.Challenge, error)
	
	// FindByID finds a challenge by its ID
	FindByID(id int) (*gzapi.Challenge, error)
	
	// List returns all challenges
	List() ([]gzapi.Challenge, error)
	
	// Create creates a new challenge
	Create(challenge *gzapi.Challenge) (*gzapi.Challenge, error)
	
	// Update updates an existing challenge
	Update(challenge *gzapi.Challenge) (*gzapi.Challenge, error)
	
	// Delete deletes a challenge
	Delete(id int) error
	
	// Exists checks if a challenge exists by name
	Exists(name string) (bool, error)
}

// AttachmentRepository defines operations for attachment data access
type AttachmentRepository interface {
	// Create creates a new attachment
	Create(challengeID int, attachment gzapi.CreateAttachmentForm) error
	
	// Delete deletes an attachment
	Delete(challengeID int, attachmentID int) error
	
	// List returns all attachments for a challenge
	List(challengeID int) ([]gzapi.Attachment, error)
}

// FlagRepository defines operations for flag data access
type FlagRepository interface {
	// Create creates a new flag
	Create(challengeID int, flag gzapi.CreateFlagForm) error
	
	// Delete deletes a flag
	Delete(flagID int) error
	
	// List returns all flags for a challenge
	List(challengeID int) ([]gzapi.Flag, error)
	
	// Exists checks if a flag exists
	Exists(challengeID int, flag string) (bool, error)
}

// GameRepository defines operations for game data access
type GameRepository interface {
	// FindByTitle finds a game by its title
	FindByTitle(title string) (*gzapi.Game, error)
	
	// FindByID finds a game by its ID
	FindByID(id int) (*gzapi.Game, error)
	
	// List returns all games
	List() ([]*gzapi.Game, error)
	
	// Create creates a new game
	Create(game gzapi.CreateGameForm) (*gzapi.Game, error)
	
	// Update updates an existing game
	Update(game *gzapi.Game) (*gzapi.Game, error)
	
	// Delete deletes a game
	Delete(id int) error
}

// CacheRepository defines operations for cache data access
type CacheRepository interface {
	// Get retrieves a value from cache
	Get(key string, dest interface{}) error
	
	// Set stores a value in cache
	Set(key string, value interface{}) error
	
	// Delete removes a value from cache
	Delete(key string) error
	
	// Exists checks if a key exists in cache
	Exists(key string) (bool, error)
}