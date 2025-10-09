package gzcli

import (
	"context"
	"fmt"
	"sync"

	"github.com/dimasma0305/gzcli/internal/gzcli/common"
	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/container"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/service"
	"github.com/dimasma0305/gzcli/internal/log"
)

// SyncManager handles synchronization operations
type SyncManager struct {
	api *gzapi.GZAPI
}

// NewSyncManager creates a new sync manager
func NewSyncManager(api *gzapi.GZAPI) *SyncManager {
	return &SyncManager{
		api: api,
	}
}

// SyncConfig holds configuration for sync operations
type SyncConfig struct {
	UpdateGame bool
	EventName  string
}

// Sync synchronizes challenges with the API
func (sm *SyncManager) Sync(conf *config.Config, syncConfig SyncConfig) error {
	log.Info("Starting synchronization process...")

	// Get challenges configuration
	challengesConf, err := config.GetChallengesYaml(conf)
	if err != nil {
		log.Error("Failed to get challenges YAML: %v", err)
		return fmt.Errorf("challenges config error: %w", err)
	}
	log.Info("Loaded %d challenges from configuration", len(challengesConf))

	// Create container with dependencies
	cnt := container.NewContainer(container.ContainerConfig{
		Config:      conf,
		API:         sm.api,
		Game:        &conf.Event,
		GetCache:    GetCache,
		SetCache:    setCache,
		DeleteCache: deleteCacheWrapperWithError,
	})

	// Get fresh games list using service
	log.Info("Fetching games from API...")
	gameSvc := cnt.GameService()
	ctx := context.Background()
	games, err := gameSvc.GetGames(ctx)
	if err != nil {
		log.Error("Failed to get games: %v", err)
		return fmt.Errorf("games fetch error: %w", err)
	}
	log.Info("Found %d games", len(games))

	// Use service to find current game
	currentGame := gameSvc.FindGame(ctx, games, conf.Event.Title)
	if currentGame == nil {
		log.Info("Current game not found, clearing cache and retrying...")
		_ = DeleteCache("config")
		return sm.Sync(conf, syncConfig) // Recursive call after cache clear
	}
	log.Info("Found current game: %s (ID: %d)", currentGame.Title, currentGame.Id)

	// Update game if needed
	if syncConfig.UpdateGame {
		if err := sm.updateGame(ctx, gameSvc, conf, currentGame); err != nil {
			return err
		}
	}

	// Validate challenges
	if err := sm.validateChallenges(challengesConf); err != nil {
		return err
	}

	// Get fresh challenges list
	log.Info("Fetching existing challenges from API...")
	conf.Event.CS = sm.api
	challenges, err := conf.Event.GetChallenges()
	if err != nil {
		log.Error("Failed to get challenges from API: %v", err)
		return fmt.Errorf("API challenges fetch error: %w", err)
	}
	log.Info("Found %d existing challenges in API", len(challenges))

	// Sync challenges
	return sm.syncChallenges(ctx, cnt, challengesConf)
}

// updateGame updates the game if needed
func (sm *SyncManager) updateGame(ctx context.Context, gameSvc *service.GameService, conf *config.Config, currentGame *gzapi.Game) error {
	log.Info("Updating game configuration...")
	if err := gameSvc.UpdateGameIfNeeded(ctx, conf, currentGame, createPosterIfNotExistOrDifferent, setCache); err != nil {
		log.Error("Failed to update game: %v", err)
		return fmt.Errorf("game update error: %w", err)
	}
	log.Info("Game updated successfully")
	return nil
}

// validateChallenges validates all challenges
func (sm *SyncManager) validateChallenges(challengesConf []config.ChallengeYaml) error {
	log.Info("Validating challenges...")
	validator := common.NewValidator()
	for _, challengeConf := range challengesConf {
		if err := validator.ValidateChallenge(challengeConf); err != nil {
			log.Error("Challenge validation failed for %s: %v", challengeConf.Name, err)
			return fmt.Errorf("validation error: %w", err)
		}
	}
	log.Info("All challenges validated successfully")
	return nil
}

// syncChallenges synchronizes all challenges
func (sm *SyncManager) syncChallenges(ctx context.Context, cnt *container.Container, challengesConf []config.ChallengeYaml) error {
	// Get challenge service from container
	challengeSvc := cnt.ChallengeService()

	// Process challenges
	log.Info("Starting challenge synchronization...")
	var wg sync.WaitGroup
	errChan := make(chan error, len(challengesConf))
	successCount := 0
	failureCount := 0

	// Create per-challenge mutexes to prevent race conditions
	challengeMutexes := make(map[string]*sync.Mutex)
	var mutexesMu sync.RWMutex

	for _, challengeConf := range challengesConf {
		wg.Add(1)
		go func(c config.ChallengeYaml) {
			defer wg.Done()

			// Get or create mutex for this challenge to prevent duplicates
			mutexesMu.Lock()
			if challengeMutexes[c.Name] == nil {
				challengeMutexes[c.Name] = &sync.Mutex{}
			}
			mutex := challengeMutexes[c.Name]
			mutexesMu.Unlock()

			// Synchronize access per challenge to prevent race conditions
			mutex.Lock()
			defer mutex.Unlock()

			log.Info("Processing challenge: %s", c.Name)
			if err := challengeSvc.Sync(ctx, c); err != nil {
				log.Error("Failed to sync challenge %s: %v", c.Name, err)
				errChan <- fmt.Errorf("challenge sync failed for %s: %w", c.Name, err)
				failureCount++
			} else {
				log.Info("Successfully synced challenge: %s", c.Name)
				successCount++
			}
		}(challengeConf)
	}

	wg.Wait()
	close(errChan)

	log.Info("Sync completed. Success: %d, Failures: %d", successCount, failureCount)

	// Check for errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("sync completed with %d errors", len(errors))
	}

	return nil
}