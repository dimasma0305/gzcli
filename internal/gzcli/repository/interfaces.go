package repository

import (
	"context"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
)

// ChallengeRepository defines the interface for challenge data operations
type ChallengeRepository interface {
	GetChallenges(ctx context.Context, gameID int) ([]gzapi.Challenge, error)
	GetChallenge(ctx context.Context, gameID int, challengeID int) (*gzapi.Challenge, error)
	CreateChallenge(ctx context.Context, gameID int, challenge *gzapi.Challenge) (*gzapi.Challenge, error)
	UpdateChallenge(ctx context.Context, gameID int, challenge *gzapi.Challenge) (*gzapi.Challenge, error)
	DeleteChallenge(ctx context.Context, gameID int, challengeID int) error
}

// AttachmentRepository defines the interface for attachment operations
type AttachmentRepository interface {
	UploadAttachment(ctx context.Context, challengeID int, filePath string) (*gzapi.Attachment, error)
	DeleteAttachment(ctx context.Context, attachmentID int) error
	GetAttachments(ctx context.Context, challengeID int) ([]gzapi.Attachment, error)
}

// FlagRepository defines the interface for flag operations
type FlagRepository interface {
	CreateFlag(ctx context.Context, challengeID int, flag *gzapi.Flag) (*gzapi.Flag, error)
	UpdateFlag(ctx context.Context, challengeID int, flag *gzapi.Flag) (*gzapi.Flag, error)
	DeleteFlag(ctx context.Context, challengeID int, flagID int) error
	GetFlags(ctx context.Context, challengeID int) ([]gzapi.Flag, error)
}

// CacheRepository defines the interface for cache operations
type CacheRepository interface {
	Get(ctx context.Context, key string, target interface{}) error
	Set(ctx context.Context, key string, value interface{}) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
}

// GameRepository defines the interface for game operations
type GameRepository interface {
	GetGames(ctx context.Context) ([]gzapi.Game, error)
	GetGame(ctx context.Context, gameID int) (*gzapi.Game, error)
	CreateGame(ctx context.Context, game *gzapi.Game) (*gzapi.Game, error)
	UpdateGame(ctx context.Context, game *gzapi.Game) (*gzapi.Game, error)
	DeleteGame(ctx context.Context, gameID int) error
}

// FileRepository defines the interface for file operations
type FileRepository interface {
	ReadFile(ctx context.Context, path string) ([]byte, error)
	WriteFile(ctx context.Context, path string, data []byte) error
	Exists(ctx context.Context, path string) (bool, error)
	DeleteFile(ctx context.Context, path string) error
	ListFiles(ctx context.Context, dir string) ([]string, error)
}