package service

import (
	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/repository"
)

// ContainerConfig holds configuration for the dependency container
type ContainerConfig struct {
	Config      *config.Config
	API         *gzapi.GZAPI
	Game        *gzapi.Game
	GetCache    func(string, interface{}) error
	SetCache    func(string, interface{}) error
	DeleteCache func(string) error
}

// Container manages service dependencies
type Container struct {
	config      *config.Config
	api         *gzapi.GZAPI
	game        *gzapi.Game
	getCache    func(string, interface{}) error
	setCache    func(string, interface{}) error
	deleteCache func(string) error
	
	// Repositories
	challengeRepo  repository.ChallengeRepository
	attachmentRepo repository.AttachmentRepository
	flagRepo       repository.FlagRepository
	gameRepo       repository.GameRepository
	cacheRepo      repository.CacheRepository
	
	// Services
	challengeService  ChallengeService
	attachmentService AttachmentService
	flagService       FlagService
	gameService       GameService
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

// ChallengeService returns the challenge service
func (c *Container) ChallengeService() ChallengeService {
	if c.challengeService == nil {
		c.challengeService = NewChallengeService(ChallengeServiceConfig{
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

// AttachmentService returns the attachment service
func (c *Container) AttachmentService() AttachmentService {
	if c.attachmentService == nil {
		c.attachmentService = NewAttachmentService(c.AttachmentRepository())
	}
	return c.attachmentService
}

// FlagService returns the flag service
func (c *Container) FlagService() FlagService {
	if c.flagService == nil {
		c.flagService = NewFlagService(c.FlagRepository())
	}
	return c.flagService
}

// GameService returns the game service
func (c *Container) GameService() GameService {
	if c.gameService == nil {
		c.gameService = NewGameService(c.GameRepository(), c.CacheRepository())
	}
	return c.gameService
}

// ChallengeRepository returns the challenge repository
func (c *Container) ChallengeRepository() repository.ChallengeRepository {
	if c.challengeRepo == nil {
		c.challengeRepo = repository.NewGZAPIChallengeRepository(c.game)
	}
	return c.challengeRepo
}

// AttachmentRepository returns the attachment repository
func (c *Container) AttachmentRepository() repository.AttachmentRepository {
	if c.attachmentRepo == nil {
		c.attachmentRepo = repository.NewGZAPIAttachmentRepository()
	}
	return c.attachmentRepo
}

// FlagRepository returns the flag repository
func (c *Container) FlagRepository() repository.FlagRepository {
	if c.flagRepo == nil {
		c.flagRepo = repository.NewGZAPIFlagRepository()
	}
	return c.flagRepo
}

// GameRepository returns the game repository
func (c *Container) GameRepository() repository.GameRepository {
	if c.gameRepo == nil {
		c.gameRepo = repository.NewGZAPIGameRepository(c.api)
	}
	return c.gameRepo
}

// CacheRepository returns the cache repository
func (c *Container) CacheRepository() repository.CacheRepository {
	if c.cacheRepo == nil {
		c.cacheRepo = repository.NewYAMLCacheRepository(c.getCache, c.setCache, c.deleteCache)
	}
	return c.cacheRepo
}