package repository

import (
	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
)

// ChallengeRepository defines operations for challenge data access
type ChallengeRepository interface {
	FindByID(id int) (*gzapi.Challenge, error)
	FindByName(name string) (*gzapi.Challenge, error)
	Create(challenge *gzapi.Challenge) error
	Update(challenge *gzapi.Challenge) error
	Delete(id int) error
	List() ([]gzapi.Challenge, error)
}

// GameRepository defines operations for game data access
type GameRepository interface {
	FindByTitle(title string) (*gzapi.Game, error)
	Create(game *gzapi.Game) error
	Update(game *gzapi.Game) error
	Delete(id int) error
	List() ([]gzapi.Game, error)
}

// AttachmentRepository defines operations for attachment data access
type AttachmentRepository interface {
	Create(challengeID int, attachment gzapi.CreateAttachmentForm) error
	Delete(challengeID int, attachmentID int) error
	List(challengeID int) ([]gzapi.Attachment, error)
}

// FlagRepository defines operations for flag data access
type FlagRepository interface {
	Create(challengeID int, flag gzapi.CreateFlagForm) error
	Delete(challengeID int, flagID int) error
	List(challengeID int) ([]gzapi.Flag, error)
}

// CacheRepository defines operations for cache data access
type CacheRepository interface {
	Get(key string, value interface{}) error
	Set(key string, value interface{}) error
	Delete(key string) error
}

// ConfigRepository defines operations for configuration data access
type ConfigRepository interface {
	GetConfig() (*config.Config, error)
	SaveConfig(conf *config.Config) error
}