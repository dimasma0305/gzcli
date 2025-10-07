//nolint:revive // Exported functions follow project conventions
package challenge

import (
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"

	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/log"
)

func IsChallengeExist(challengeName string, challenges []gzapi.Challenge) bool {
	challengeMap := make(map[string]struct{}, len(challenges))
	for _, c := range challenges {
		challengeMap[c.Title] = struct{}{}
	}
	_, exists := challengeMap[challengeName]
	return exists
}

func IsExistInArray(value string, array []string) bool {
	for _, v := range array {
		if v == value {
			return true
		}
	}
	return false
}

func IsConfigEdited(challengeConf *ChallengeYaml, challengeData *gzapi.Challenge, getCache func(string, interface{}) error) bool {
	var cacheChallenge gzapi.Challenge
	if err := getCache(challengeConf.Category+"/"+challengeConf.Name+"/challenge", &cacheChallenge); err != nil {
		return true
	}

	if challengeData.Hints == nil {
		challengeData.Hints = []string{}
	}
	return !cmp.Equal(*challengeData, cacheChallenge)
}

func MergeChallengeData(challengeConf *ChallengeYaml, challengeData *gzapi.Challenge) *gzapi.Challenge {
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

	challengeData.Title = challengeConf.Name
	challengeData.Category = challengeConf.Category
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

func SyncChallenge(config *Config, challengeConf ChallengeYaml, challenges []gzapi.Challenge, api *gzapi.GZAPI, getCache func(string, interface{}) error, setCache func(string, interface{}) error) error {
	var challengeData *gzapi.Challenge
	var err error

	log.InfoH2("Starting sync for challenge: %s (Type: %s, Category: %s)", challengeConf.Name, challengeConf.Type, challengeConf.Category)

	// Check existence using the original challenges list first to avoid unnecessary API calls
	if !IsChallengeExist(challengeConf.Name, challenges) {
		// Double-check with fresh challenges list to prevent race conditions
		// This check happens inside the mutex-protected section in the calling function
		log.InfoH3("Challenge %s not found in initial list, fetching fresh challenges list", challengeConf.Name)
		freshChallenges, err := config.Event.GetChallenges()
		if err != nil {
			log.Error("Failed to get fresh challenges list for %s: %v", challengeConf.Name, err)
			// Fallback to original challenges list if fresh fetch fails
			freshChallenges = challenges
		} else {
			log.InfoH3("Fetched fresh challenges list for %s (%d challenges)", challengeConf.Name, len(freshChallenges))
		}

		// Final check to prevent duplicates
		if !IsChallengeExist(challengeConf.Name, freshChallenges) {
			log.InfoH2("Creating new challenge: %s", challengeConf.Name)
			challengeData, err = config.Event.CreateChallenge(gzapi.CreateChallengeForm{
				Title:    challengeConf.Name,
				Category: challengeConf.Category,
				Tag:      challengeConf.Category,
				Type:     challengeConf.Type,
			})
			if err != nil {
				// Check if this is a duplicate creation error (common with race conditions)
				if strings.Contains(strings.ToLower(err.Error()), "already exists") ||
					strings.Contains(strings.ToLower(err.Error()), "duplicate") ||
					strings.Contains(strings.ToLower(err.Error()), "conflict") {
					log.InfoH2("Challenge %s already exists (created by another process), fetching existing challenge", challengeConf.Name)
					challengeData, err = config.Event.GetChallenge(challengeConf.Name)
					if err != nil {
						log.Error("Failed to get existing challenge %s after creation conflict: %v", challengeConf.Name, err)
						return fmt.Errorf("get existing challenge %s: %w", challengeConf.Name, err)
					}
					challengeData.CS = api
					log.InfoH3("Successfully fetched existing challenge %s after creation conflict", challengeConf.Name)
				} else {
					log.Error("Failed to create challenge %s: %v", challengeConf.Name, err)
					return fmt.Errorf("create challenge %s: %w", challengeConf.Name, err)
				}
			} else {
				challengeData.CS = api
				log.InfoH2("Successfully created challenge: %s (ID: %d)", challengeConf.Name, challengeData.Id)
			}
		} else {
			log.InfoH2("Challenge %s was created by another process, fetching existing challenge", challengeConf.Name)
			// Challenge was created by another goroutine, fetch it
			challengeData, err = config.Event.GetChallenge(challengeConf.Name)
			if err != nil {
				log.Error("Failed to get newly created challenge %s: %v", challengeConf.Name, err)
				return fmt.Errorf("get challenge %s: %w", challengeConf.Name, err)
			}
			log.InfoH3("Successfully fetched existing challenge %s", challengeConf.Name)
		}

		// Ensure the API client is properly set for newly created/fetched challenges
		challengeData.CS = api
	} else {
		log.InfoH2("Updating existing challenge: %s", challengeConf.Name)
		if err = getCache(challengeConf.Category+"/"+challengeConf.Name+"/challenge", &challengeData); err != nil {
			log.InfoH3("Cache miss for %s, fetching from API", challengeConf.Name)
			challengeData, err = config.Event.GetChallenge(challengeConf.Name)
			if err != nil {
				log.Error("Failed to get challenge %s from API: %v", challengeConf.Name, err)
				return fmt.Errorf("get challenge %s: %w", challengeConf.Name, err)
			}
			log.InfoH3("Successfully fetched challenge %s from API", challengeConf.Name)
		} else {
			log.InfoH3("Found challenge %s in cache", challengeConf.Name)
		}

		// fix bug nill pointer because cache didn't return gzapi
		challengeData.CS = api
		// fix bug isEnable always be false after sync
		challengeData.IsEnabled = nil
	}

	log.InfoH2("Processing attachments for %s", challengeConf.Name)
	err = HandleChallengeAttachments(challengeConf, challengeData, api)
	if err != nil {
		log.Error("Failed to handle attachments for %s: %v", challengeConf.Name, err)
		return fmt.Errorf("attachment handling failed for %s: %w", challengeConf.Name, err)
	}
	log.InfoH2("Attachments processed successfully for %s", challengeConf.Name)

	log.InfoH2("Updating flags for %s", challengeConf.Name)
	err = UpdateChallengeFlags(config, challengeConf, challengeData)
	if err != nil {
		log.Error("Failed to update flags for %s: %v", challengeConf.Name, err)
		return fmt.Errorf("update flags for %s: %w", challengeConf.Name, err)
	}
	log.InfoH2("Flags updated successfully for %s", challengeConf.Name)

	log.InfoH2("Merging challenge data for %s", challengeConf.Name)
	challengeData = MergeChallengeData(&challengeConf, challengeData)

	if IsConfigEdited(&challengeConf, challengeData, getCache) {
		log.InfoH2("Configuration changed for %s, updating...", challengeConf.Name)
		if challengeData, err = challengeData.Update(*challengeData); err != nil {
			log.Error("Update failed for %s: %v", challengeConf.Name, err.Error())
			if strings.Contains(err.Error(), "404") {
				log.InfoH3("Got 404 error, refreshing challenge data for %s", challengeConf.Name)
				challengeData, err = config.Event.GetChallenge(challengeConf.Name)
				if err != nil {
					log.Error("Failed to get challenge %s after 404: %v", challengeConf.Name, err)
					return fmt.Errorf("get challenge %s: %w", challengeConf.Name, err)
				}
				log.InfoH3("Retrying update for %s", challengeConf.Name)
				challengeData, err = challengeData.Update(*challengeData)
				if err != nil {
					log.Error("Update retry failed for %s: %v", challengeConf.Name, err)
					return fmt.Errorf("update challenge %s: %w", challengeConf.Name, err)
				}
			} else {
				return fmt.Errorf("update challenge %s: %w", challengeConf.Name, err)
			}
		}
		if challengeData == nil {
			log.Error("Update returned nil challenge data for %s", challengeConf.Name)
			return fmt.Errorf("update challenge failed for %s", challengeConf.Name)
		}
		log.InfoH2("Successfully updated challenge %s", challengeConf.Name)

		log.InfoH3("Caching updated challenge data for %s", challengeConf.Name)
		if err := setCache(challengeData.Category+"/"+challengeConf.Name+"/challenge", challengeData); err != nil {
			log.Error("Failed to cache challenge data for %s: %v", challengeConf.Name, err)
			return fmt.Errorf("cache error for %s: %w", challengeConf.Name, err)
		}
		log.InfoH3("Successfully cached challenge data for %s", challengeConf.Name)
	} else {
		log.InfoH2("Challenge %s is unchanged, skipping update", challengeConf.Name)
	}

	log.InfoH2("Successfully completed sync for challenge: %s", challengeConf.Name)
	return nil
}
