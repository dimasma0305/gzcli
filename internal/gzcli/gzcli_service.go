package gzcli

import (
	"fmt"
	"sync"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/service"
	"github.com/dimasma0305/gzcli/internal/log"
)

// SyncWithServices synchronizes challenges using the new service layer
func (gz *GZ) SyncWithServices() error {
	log.Info("Starting sync process with service layer...")

	// Use the event name stored in the GZ instance
	conf, err := config.GetConfigWithEvent(gz.api, gz.eventName, GetCache, setCache, deleteCacheWrapper, createNewGameWrapper)
	if err != nil {
		log.Error("Failed to get config: %v", err)
		return fmt.Errorf("config error: %w", err)
	}
	log.Info("Config loaded successfully")

	// Get fresh challenges config
	log.Info("Loading challenges configuration...")
	challengesConf, err := config.GetChallengesYaml(conf)
	if err != nil {
		log.Error("Failed to get challenges YAML: %v", err)
		return fmt.Errorf("challenges config error: %w", err)
	}
	log.Info("Loaded %d challenges from configuration", len(challengesConf))

	// Get fresh games list
	log.Info("Fetching games from API...")
	games, err := gz.api.GetGames()
	if err != nil {
		log.Error("Failed to get games: %v", err)
		return fmt.Errorf("games fetch error: %w", err)
	}
	log.Info("Found %d games", len(games))

	// Create dependency container
	container := service.NewContainer(service.ContainerConfig{
		Config:      conf,
		API:         gz.api,
		Game:        &conf.Event,
		GetCache:    GetCache,
		SetCache:    setCache,
		DeleteCache: func(key string) error { deleteCacheWrapper(key); return nil },
	})

	// Get services from container
	gameService := container.GameService()
	challengeService := container.ChallengeService()

	// Find current game
	currentGame := gameService.FindGame(games, conf.Event.Title)
	if currentGame == nil {
		log.Info("Current game not found, clearing cache and retrying...")
		_ = DeleteCache("config")
		return gz.SyncWithServices()
	}
	log.Info("Found current game: %s (ID: %d)", currentGame.Title, currentGame.Id)

	// Update game if needed
	if gz.UpdateGame {
		log.Info("Updating game configuration...")
		if err := gameService.UpdateGameIfNeeded(conf, currentGame, gz.api, createPosterIfNotExistOrDifferent, setCache); err != nil {
			log.Error("Failed to update game: %v", err)
			return fmt.Errorf("game update error: %w", err)
		}
		log.Info("Game updated successfully")
	}

	// Validate challenges
	log.Info("Validating challenges...")
	if err := validateChallengesWrapper(challengesConf); err != nil {
		log.Error("Challenge validation failed: %v", err)
		return fmt.Errorf("validation error: %w", err)
	}
	log.Info("All challenges validated successfully")

	// Process challenges using service layer
	log.Info("Starting challenge synchronization with service layer...")
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
			if err := challengeService.Sync(c); err != nil {
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

	// Return first error if any
	select {
	case err := <-errChan:
		return err
	default:
		log.Info("All challenges synced successfully!")
		return nil
	}
}