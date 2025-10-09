package container

import (
	"context"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/repository"
	"github.com/dimasma0305/gzcli/internal/gzcli/service"
)

// ContainerConfig holds configuration for the dependency container
type ContainerConfig struct {
	Config      *config.Config
	API         *gzapi.GZAPI
	Game        *gzapi.Game
	GetCache    func(string, interface{}) error
	SetCache    func(string, interface{}) error
	DeleteCache func(string)
}

// Container manages dependencies and provides services
type Container struct {
	config      *config.Config
	api         *gzapi.GZAPI
	game        *gzapi.Game
	getCache    func(string, interface{}) error
	setCache    func(string, interface{}) error
	deleteCache func(string)

	// Cached repositories
	challengeRepo  repository.ChallengeRepository
	cacheRepo      repository.CacheRepository
	attachmentRepo repository.AttachmentRepository
	flagRepo       repository.FlagRepository
	gameRepo       repository.GameRepository

	// Cached services
	challengeService *service.ChallengeService
	gameService      *service.GameService
}

// NewContainer creates a new dependency container
func NewContainer(config ContainerConfig) *Container {
	return &Container{
		config:      config.Config,
		api:         config.API,
		game:        config.Game,
		getCache:    config.GetCache,
		setCache:    config.SetCache,
		deleteCache: config.DeleteCache,
	}
}

// ChallengeRepository returns the challenge repository
func (c *Container) ChallengeRepository() repository.ChallengeRepository {
	if c.challengeRepo == nil {
		c.challengeRepo = repository.NewGZAPIChallengeRepository(c.api)
	}
	return c.challengeRepo
}

// CacheRepository returns the cache repository
func (c *Container) CacheRepository() repository.CacheRepository {
	if c.cacheRepo == nil {
		c.cacheRepo = repository.NewYAMLCacheRepository(c.getCache, c.setCache, c.deleteCache)
	}
	return c.cacheRepo
}

// AttachmentRepository returns the attachment repository
func (c *Container) AttachmentRepository() repository.AttachmentRepository {
	if c.attachmentRepo == nil {
		// This would be implemented when attachment repository is created
		c.attachmentRepo = nil
	}
	return c.attachmentRepo
}

// FlagRepository returns the flag repository
func (c *Container) FlagRepository() repository.FlagRepository {
	if c.flagRepo == nil {
		// This would be implemented when flag repository is created
		c.flagRepo = nil
	}
	return c.flagRepo
}

// GameRepository returns the game repository
func (c *Container) GameRepository() repository.GameRepository {
	if c.gameRepo == nil {
		// This would be implemented when game repository is created
		c.gameRepo = nil
	}
	return c.gameRepo
}

// ChallengeService returns the challenge service
func (c *Container) ChallengeService() *service.ChallengeService {
	if c.challengeService == nil {
		c.challengeService = service.NewChallengeService(service.ChallengeServiceConfig{
			ChallengeRepo:  c.ChallengeRepository(),
			Cache:          c.CacheRepository(),
			AttachmentRepo: c.AttachmentRepository(),
			FlagRepo:       c.FlagRepository(),
			API:            c.api,
			GameID:         c.game.Id,
		})
	}
	return c.challengeService
}

// GameService returns the game service
func (c *Container) GameService() *service.GameService {
	if c.gameService == nil {
		c.gameService = service.NewGameService(
			c.CacheRepository(),
			c.api,
		)
	}
	return c.gameService
}

// WithContext creates a new container with context for operations
func (c *Container) WithContext(ctx context.Context) *Container {
	// For now, return the same container
	// In the future, this could be enhanced to handle context-specific operations
	return c
}