package challenge

import (
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/log"
)

// IsChallengeExist checks if a challenge with a given name exists in a slice of challenges.
func IsChallengeExist(challengeName string, challenges []gzapi.Challenge) bool {
	challengeMap := make(map[string]struct{}, len(challenges))
	for _, c := range challenges {
		challengeMap[c.Title] = struct{}{}
	}
	_, exists := challengeMap[challengeName]
	return exists
}

// IsExistInArray checks if a string value exists in a slice of strings.
func IsExistInArray(value string, array []string) bool {
	for _, v := range array {
		if v == value {
			return true
		}
	}
	return false
}

// buildChallengeCacheKey constructs a unique cache key for a challenge based on its event,
// category, and name.
// The format is: <eventName>/<category>/<challengeName>/challenge
func buildChallengeCacheKey(eventName, category, challengeName string) string {
	return fmt.Sprintf("%s/%s/%s/challenge", eventName, category, challengeName)
}

// IsConfigEdited compares the current challenge data with its cached version to determine
// if there have been any changes.
func IsConfigEdited(conf *config.Config, challengeConf *config.ChallengeYaml, challengeData *gzapi.Challenge, getCache func(string, interface{}) error) bool {
	var cacheChallenge gzapi.Challenge
	cacheKey := buildChallengeCacheKey(conf.EventName, challengeConf.Category, challengeConf.Name)
	if err := getCache(cacheKey, &cacheChallenge); err != nil {
		return true
	}

	if challengeData.Hints == nil {
		challengeData.Hints = []string{}
	}
	return !cmp.Equal(*challengeData, cacheChallenge)
}

// MergeChallengeData merges the local challenge configuration (from YAML) into the challenge
// data object fetched from the API.
func MergeChallengeData(challengeConf *config.ChallengeYaml, challengeData *gzapi.Challenge) *gzapi.Challenge {
	// Set resource limits from container configuration, with defaults if not specified
	if challengeConf.Container.MemoryLimit > 0 {
		challengeData.MemoryLimit = challengeConf.Container.MemoryLimit
	} else {
		challengeData.MemoryLimit = 128 // Default fallback
	}

	if challengeConf.Container.CpuCount > 0 {
		challengeData.CpuCount = challengeConf.Container.CpuCount
	} else {
		challengeData.CpuCount = 1 // Default fallback
	}

	if challengeConf.Container.StorageLimit > 0 {
		challengeData.StorageLimit = challengeConf.Container.StorageLimit
	} else {
		challengeData.StorageLimit = 128 // Default fallback
	}

	// Normalize category and name before setting (ensures consistency across sync and watcher)
	normalizedCategory, normalizedName := config.NormalizeChallengeCategory(challengeConf.Category, challengeConf.Name)

	challengeData.Title = normalizedName
	challengeData.Category = normalizedCategory
	challengeData.Content = fmt.Sprintf("Author: **%s**\n\n%s", challengeConf.Author, challengeConf.Description)
	challengeData.Type = challengeConf.Type
	challengeData.Hints = challengeConf.Hints
	challengeData.FlagTemplate = challengeConf.Container.FlagTemplate
	challengeData.ContainerImage = challengeConf.Container.ContainerImage
	challengeData.ContainerExposePort = challengeConf.Container.ContainerExposePort
	challengeData.EnableTrafficCapture = challengeConf.Container.EnableTrafficCapture
	challengeData.OriginalScore = challengeConf.Value

	if challengeData.OriginalScore >= 100 {
		challengeData.MinScoreRate = 0.10
	} else {
		challengeData.MinScoreRate = 1
	}

	return challengeData
}

// isDuplicateError checks if an error indicates a duplicate entry, which can occur during
// challenge creation if the challenge already exists.
func isDuplicateError(err error) bool {
	errLower := strings.ToLower(err.Error())
	return strings.Contains(errLower, "already exists") ||
		strings.Contains(errLower, "duplicate") ||
		strings.Contains(errLower, "conflict")
}

// createChallengeWithRetry attempts to create a new challenge on the server. If a duplicate
// error is encountered, it fetches the existing challenge data instead.
func createChallengeWithRetry(conf *config.Config, challengeConf config.ChallengeYaml, api *gzapi.GZAPI) (*gzapi.Challenge, error) {
	challengeData, err := conf.Event.CreateChallenge(gzapi.CreateChallengeForm{
		Title:    challengeConf.Name,
		Category: challengeConf.Category,
		Tag:      challengeConf.Category,
		Type:     challengeConf.Type,
	})

	if err != nil {
		if isDuplicateError(err) {
			challengeData, err = conf.Event.GetChallenge(challengeConf.Name)
			if err != nil {
				log.Error("Failed to get existing challenge %s after creation conflict: %v", challengeConf.Name, err)
				return nil, fmt.Errorf("get existing challenge %s: %w", challengeConf.Name, err)
			}
		} else {
			log.Error("Failed to create challenge %s: %v", challengeConf.Name, err)
			return nil, fmt.Errorf("create challenge %s: %w", challengeConf.Name, err)
		}
	}

	challengeData.CS = api
	return challengeData, nil
}

// handleNewChallenge orchestrates the process of creating a new challenge, including a final
// check to prevent race conditions from concurrent operations.
func handleNewChallenge(conf *config.Config, challengeConf config.ChallengeYaml, challenges []gzapi.Challenge, api *gzapi.GZAPI) (*gzapi.Challenge, error) {
	freshChallenges, err := conf.Event.GetChallenges()
	if err != nil {
		log.Error("Failed to get fresh challenges list for %s: %v", challengeConf.Name, err)
		freshChallenges = challenges
	}

	// Final check to prevent duplicates
	if !IsChallengeExist(challengeConf.Name, freshChallenges) {
		return createChallengeWithRetry(conf, challengeConf, api)
	}

	// Challenge was created by another goroutine, fetch it
	challengeData, err := conf.Event.GetChallenge(challengeConf.Name)
	if err != nil {
		log.Error("Failed to get newly created challenge %s: %v", challengeConf.Name, err)
		return nil, fmt.Errorf("get challenge %s: %w", challengeConf.Name, err)
	}

	challengeData.CS = api
	return challengeData, nil
}

// handleExistingChallenge retrieves the data for an existing challenge, first checking the
// cache and then falling back to an API call if necessary.
func handleExistingChallenge(conf *config.Config, challengeConf config.ChallengeYaml, api *gzapi.GZAPI, getCache func(string, interface{}) error) (*gzapi.Challenge, error) {
	var challengeData *gzapi.Challenge

	cacheKey := buildChallengeCacheKey(conf.EventName, challengeConf.Category, challengeConf.Name)
	err := getCache(cacheKey, &challengeData)
	if err != nil {
		challengeData, err = conf.Event.GetChallenge(challengeConf.Name)
		if err != nil {
			log.Error("Failed to get challenge %s from API: %v", challengeConf.Name, err)
			return nil, fmt.Errorf("get challenge %s: %w", challengeConf.Name, err)
		}
	}

	// fix bug nill pointer because cache didn't return gzapi
	challengeData.CS = api
	// fix bug isEnable always be false after sync
	challengeData.IsEnabled = nil

	return challengeData, nil
}

// SyncChallenge is the main entry point for synchronizing a single challenge. It determines
// whether the challenge is new or existing and calls the appropriate handlers.
func SyncChallenge(conf *config.Config, challengeConf config.ChallengeYaml, challenges []gzapi.Challenge, api *gzapi.GZAPI, getCache func(string, interface{}) error, setCache func(string, interface{}) error) error {
	return SyncChallengeWithExisting(conf, challengeConf, challenges, api, getCache, setCache, nil)
}

// SyncChallengeWithExisting provides a way to sync a challenge, with an option to force
// an update by providing the existing challenge data.
func SyncChallengeWithExisting(conf *config.Config, challengeConf config.ChallengeYaml, challenges []gzapi.Challenge, api *gzapi.GZAPI, getCache func(string, interface{}) error, setCache func(string, interface{}) error, existingChallenge *gzapi.Challenge) error {
	var challengeData *gzapi.Challenge
	var err error

	// Determine the sync path based on challenge state
	switch {
	case existingChallenge != nil:
		// If an existing challenge is provided, use it directly (force update mode)
		challengeData = existingChallenge
		challengeData.CS = api
		challengeData.IsEnabled = nil // fix bug isEnable always be false after sync
	case !IsChallengeExist(challengeConf.Name, challenges):
		// Check existence using the original challenges list first to avoid unnecessary API calls
		challengeData, err = handleNewChallenge(conf, challengeConf, challenges, api)
		if err != nil {
			return err
		}
	default:
		challengeData, err = handleExistingChallenge(conf, challengeConf, api, getCache)
		if err != nil {
			return err
		}
	}

	if err := processAttachmentsAndFlags(conf, challengeConf, challengeData, api); err != nil {
		return err
	}

	challengeData = MergeChallengeData(&challengeConf, challengeData)

	if err := updateChallengeIfNeeded(conf, &challengeConf, challengeData, getCache, setCache); err != nil {
		return err
	}

	log.Info("✓ %s", challengeConf.Name)
	return nil
}

// processAttachmentsAndFlags handles the synchronization of challenge attachments and flags.
func processAttachmentsAndFlags(conf *config.Config, challengeConf config.ChallengeYaml, challengeData *gzapi.Challenge, api *gzapi.GZAPI) error {
	err := HandleChallengeAttachments(challengeConf, challengeData, api)
	if err != nil {
		log.Error("Failed to handle attachments for %s: %v", challengeConf.Name, err)
		return fmt.Errorf("attachment handling failed for %s: %w", challengeConf.Name, err)
	}

	err = UpdateChallengeFlags(conf, challengeConf, challengeData)
	if err != nil {
		log.Error("Failed to update flags for %s: %v", challengeConf.Name, err)
		return fmt.Errorf("update flags for %s: %w", challengeConf.Name, err)
	}

	return nil
}

// updateChallengeWithRetry attempts to update a challenge on the server. It includes retry logic
// to handle cases where a 404 error might occur due to eventual consistency.
func updateChallengeWithRetry(conf *config.Config, challengeConf *config.ChallengeYaml, challengeData *gzapi.Challenge) (*gzapi.Challenge, error) {
	fmt.Printf("%+v\n", challengeData)
	updatedData, err := challengeData.Update(*challengeData)
	// print all object and key
	if err == nil {
		return updatedData, nil
	}

	log.Error("Update failed for %s: %v", challengeConf.Name, err.Error())

	if !strings.Contains(err.Error(), "404") {
		return nil, fmt.Errorf("update challenge %s: %w", challengeConf.Name, err)
	}

	log.InfoH3("Got 404 error, refreshing challenge data for %s", challengeConf.Name)
	challengeData, err = conf.Event.GetChallenge(challengeConf.Name)
	if err != nil {
		log.Error("Failed to get challenge %s after 404: %v", challengeConf.Name, err)
		return nil, fmt.Errorf("get challenge %s: %w", challengeConf.Name, err)
	}

	log.InfoH3("Retrying update for %s", challengeConf.Name)
	updatedData, err = challengeData.Update(*challengeData)
	if err != nil {
		log.Error("Update retry failed for %s: %v", challengeConf.Name, err)
		return nil, fmt.Errorf("update challenge %s: %w", challengeConf.Name, err)
	}

	return updatedData, nil
}

// updateChallengeIfNeeded checks if a challenge's configuration has changed and, if so,
// triggers an update on the server and updates the cache.
func updateChallengeIfNeeded(conf *config.Config, challengeConf *config.ChallengeYaml, challengeData *gzapi.Challenge, getCache func(string, interface{}) error, setCache func(string, interface{}) error) error {
	if !IsConfigEdited(conf, challengeConf, challengeData, getCache) {
		return nil
	}

	updatedData, err := updateChallengeWithRetry(conf, challengeConf, challengeData)
	if err != nil {
		return err
	}

	if updatedData == nil {
		log.Error("Update returned nil challenge data for %s", challengeConf.Name)
		return fmt.Errorf("update challenge failed for %s", challengeConf.Name)
	}

	cacheKey := buildChallengeCacheKey(conf.EventName, updatedData.Category, challengeConf.Name)
	if err := setCache(cacheKey, updatedData); err != nil {
		log.Error("Failed to cache challenge data for %s: %v", challengeConf.Name, err)
		return fmt.Errorf("cache error for %s: %w", challengeConf.Name, err)
	}

	return nil
}
