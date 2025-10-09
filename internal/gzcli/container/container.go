package container

import (
	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/repository"
	"github.com/dimasma0305/gzcli/internal/gzcli/service"
)

// Container holds all dependencies and provides services
type Container struct {
	config *config.Config
	api    *gzapi.GZAPI
	game   *config.Event

	// Repositories
	challengeRepo  repository.ChallengeRepository
	gameRepo       repository.GameRepository
	attachmentRepo repository.AttachmentRepository
	flagRepo       repository.FlagRepository
	cacheRepo      repository.CacheRepository

	// Services
	challengeService  *service.ChallengeService
	gameService       *service.GameService
	attachmentService *service.AttachmentService
	flagService       *service.FlagService
}

// ContainerConfig holds configuration for the container
type ContainerConfig struct {
	Config      *config.Config
	API         *gzapi.GZAPI
	Game        *config.Event
	GetCache    func(string, interface{}) error
	SetCache    func(string, interface{}) error
	DeleteCache func(string) error
}

// NewContainer creates a new container with all dependencies
func NewContainer(config ContainerConfig) *Container {
	c := &Container{
		config: config.Config,
		api:    config.API,
		game:   config.Game,
	}

	// Initialize repositories
	c.challengeRepo = repository.NewGZAPIChallengeRepository(config.API)
	c.gameRepo = repository.NewGZAPIGameRepository(config.API)
	c.attachmentRepo = repository.NewGZAPIAttachmentRepository(c.challengeRepo)
	c.flagRepo = repository.NewGZAPIFlagRepository(config.API)
	c.cacheRepo = repository.NewYAMLCacheRepository(config.GetCache, config.SetCache, config.DeleteCache)

	// Initialize services
	c.challengeService = service.NewChallengeService(service.ChallengeServiceConfig{
		ChallengeRepo:  c.challengeRepo,
		Cache:          c.cacheRepo,
		AttachmentRepo: c.attachmentRepo,
		FlagRepo:       c.flagRepo,
		API:            config.API,
		GameID:         config.Game.Id,
	})

	c.gameService = service.NewGameService(service.GameServiceConfig{
		GameRepo: c.gameRepo,
		Cache:    c.cacheRepo,
		API:      config.API,
	})

	c.attachmentService = service.NewAttachmentService(service.AttachmentServiceConfig{
		AttachmentRepo: c.attachmentRepo,
		ChallengeRepo:  c.challengeRepo,
	})

	c.flagService = service.NewFlagService(service.FlagServiceConfig{
		FlagRepo:      c.flagRepo,
		ChallengeRepo: c.challengeRepo,
	})

	return c
}

// ChallengeService returns the challenge service
func (c *Container) ChallengeService() *service.ChallengeService {
	return c.challengeService
}

// GameService returns the game service
func (c *Container) GameService() *service.GameService {
	return c.gameService
}

// AttachmentService returns the attachment service
func (c *Container) AttachmentService() *service.AttachmentService {
	return c.attachmentService
}

// FlagService returns the flag service
func (c *Container) FlagService() *service.FlagService {
	return c.flagService
}

// Config returns the configuration
func (c *Container) Config() *config.Config {
	return c.config
}

// API returns the API client
func (c *Container) API() *gzapi.GZAPI {
	return c.api
}

// Game returns the game configuration
func (c *Container) Game() *config.Event {
	return c.game
}