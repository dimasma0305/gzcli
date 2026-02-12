//nolint:revive // Exported functions follow project conventions
package challenge

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/log"
)

// DeleteFunc abstracts challenge deletion to allow testing without API calls.
type DeleteFunc func(*gzapi.Challenge) error

func deleteChallenge(c *gzapi.Challenge) error {
	return c.Delete()
}

// RemoveDuplicateChallenges deletes duplicate challenges (same title) from the remote list.
// It keeps the lowest-ID challenge for each title and deletes the rest using the provided deleteFunc.
// Returns the deduplicated slice that should be used for subsequent sync operations, and a flag indicating if deletions occurred.
func RemoveDuplicateChallenges(challenges []gzapi.Challenge, deleteFunc DeleteFunc) ([]gzapi.Challenge, bool, error) {
	if deleteFunc == nil {
		deleteFunc = deleteChallenge
	}

	if len(challenges) == 0 {
		return challenges, false, nil
	}

	byTitle := make(map[string]gzapi.Challenge, len(challenges))
	var duplicates []*gzapi.Challenge

	for i := range challenges {
		current := challenges[i] // create stable reference
		if keep, ok := byTitle[current.Title]; ok {
			if current.Id < keep.Id {
				dup := keep
				duplicates = append(duplicates, &dup)
				byTitle[current.Title] = current
			} else {
				dup := current
				duplicates = append(duplicates, &dup)
			}
		} else {
			byTitle[current.Title] = current
		}
	}

	if len(duplicates) > 0 {
		log.Info("Found %d duplicate challenges; deleting extras", len(duplicates))
	}

	var deleteErrs []string
	for _, dup := range duplicates {
		if dup == nil {
			continue
		}
		if err := deleteFunc(dup); err != nil {
			log.Error("Failed to delete duplicate challenge %s (id %d): %v", dup.Title, dup.Id, err)
			deleteErrs = append(deleteErrs, fmt.Sprintf("%s(%d): %v", dup.Title, dup.Id, err))
		} else {
			log.Info("Deleted duplicate challenge %s (id %d)", dup.Title, dup.Id)
		}
	}

	if len(deleteErrs) > 0 {
		return nil, true, fmt.Errorf("duplicate cleanup errors: %s", strings.Join(deleteErrs, "; "))
	}

	deduped := make([]gzapi.Challenge, 0, len(byTitle))
	for _, c := range byTitle {
		deduped = append(deduped, c)
	}

	return deduped, len(duplicates) > 0, nil
}

func IsChallengeExist(challengeName string, challenges []gzapi.Challenge) bool {
	for i := range challenges {
		if challenges[i].Title == challengeName {
			return true
		}
	}
	return false
}

// findChallengeByTitle returns a copy of the challenge with the given title, if present.
func findChallengeByTitle(challenges []gzapi.Challenge, title string) *gzapi.Challenge {
	for i := range challenges {
		if challenges[i].Title == title {
			// return a copy to avoid mutating the shared slice
			c := challenges[i]
			return &c
		}
	}
	return nil
}

type attachmentHandler func(config.ChallengeYaml, *gzapi.Challenge, *gzapi.GZAPI) error
type flagHandler func(*config.Config, config.ChallengeYaml, *gzapi.Challenge) error
type challengeRefresher func() (*gzapi.Challenge, error)

func IsExistInArray(value string, array []string) bool {
	for _, v := range array {
		if v == value {
			return true
		}
	}
	return false
}

func ensureFreshChallengeData(conf *config.Config, challengeConf config.ChallengeYaml, api *gzapi.GZAPI, challengeData *gzapi.Challenge, refresher challengeRefresher) (*gzapi.Challenge, error) {
	if challengeData != nil && challengeData.CS != nil && challengeData.GameId == conf.Event.Id && challengeData.Id != 0 {
		return challengeData, nil
	}

	if refresher == nil {
		return nil, fmt.Errorf("challenge refresher is nil")
	}

	refreshed, err := refresher()
	if err != nil {
		log.Error("Failed to refresh challenge %s: %v", challengeConf.Name, err)
		return nil, fmt.Errorf("refresh challenge %s: %w", challengeConf.Name, err)
	}

	refreshed.CS = api
	refreshed.GameId = conf.Event.Id
	refreshed.IsEnabled = nil

	return refreshed, nil
}

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errLower := strings.ToLower(err.Error())
	return strings.Contains(errLower, "404") || strings.Contains(errLower, "not found")
}

// buildChallengeCacheKey constructs the cache key for a challenge
// Format: <eventname>/<category>/<challenge>/challenge
func buildChallengeCacheKey(eventName, category, challengeName string) string {
	return fmt.Sprintf("%s/%s/%s/challenge", eventName, category, challengeName)
}

func IsConfigEdited(conf *config.Config, challengeConf *config.ChallengeYaml, challengeData *gzapi.Challenge, getCache func(string, interface{}) error) bool {
	var cacheChallenge gzapi.Challenge
	cacheKey := buildChallengeCacheKey(conf.EventName, challengeConf.Category, challengeConf.Name)
	if err := getCache(cacheKey, &cacheChallenge); err != nil {
		return true
	}
	// Some cache providers may return zero-value data on misses without an error.
	if cacheChallenge.Title == "" && challengeData.Title != "" {
		return true
	}

	if challengeData.Hints == nil {
		challengeData.Hints = []string{}
	}
	return !cmp.Equal(toComparableChallenge(*challengeData), toComparableChallenge(cacheChallenge))
}

type comparableChallenge struct {
	Title                string
	Content              string
	Category             string
	Type                 string
	Hints                []string
	FlagTemplate         string
	IsEnabled            *bool
	ContainerImage       string
	MemoryLimit          int
	CpuCount             int
	StorageLimit         int
	ContainerExposePort  int
	NetworkMode          string
	EnableTrafficCapture bool
	DisableBloodBonus    bool
	DeadlineUtc          int64
	SubmissionLimit      int
	OriginalScore        int
	MinScoreRate         float64
}

func toComparableChallenge(c gzapi.Challenge) comparableChallenge {
	if c.Hints == nil {
		c.Hints = []string{}
	}
	return comparableChallenge{
		Title:                c.Title,
		Content:              c.Content,
		Category:             c.Category,
		Type:                 c.Type,
		Hints:                c.Hints,
		FlagTemplate:         c.FlagTemplate,
		IsEnabled:            c.IsEnabled,
		ContainerImage:       c.ContainerImage,
		MemoryLimit:          c.MemoryLimit,
		CpuCount:             c.CpuCount,
		StorageLimit:         c.StorageLimit,
		ContainerExposePort:  c.ContainerExposePort,
		NetworkMode:          c.NetworkMode,
		EnableTrafficCapture: c.EnableTrafficCapture,
		DisableBloodBonus:    c.DisableBloodBonus,
		DeadlineUtc:          c.DeadlineUtc,
		SubmissionLimit:      c.SubmissionLimit,
		OriginalScore:        c.OriginalScore,
		MinScoreRate:         c.MinScoreRate,
	}
}

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

	if challengeConf.Container.NetworkMode != "" {
		challengeData.NetworkMode = challengeConf.Container.NetworkMode
	} else {
		challengeData.NetworkMode = "Open" // Default network mode
	}

	challengeData.EnableTrafficCapture = challengeConf.Container.EnableTrafficCapture
	challengeData.DisableBloodBonus = challengeConf.DisableBloodBonus
	challengeData.DeadlineUtc = challengeConf.DeadlineUtc
	challengeData.SubmissionLimit = challengeConf.SubmissionLimit
	challengeData.OriginalScore = challengeConf.Value

	if challengeData.OriginalScore >= 100 {
		challengeData.MinScoreRate = 0.10
	} else {
		challengeData.MinScoreRate = 1
	}

	return challengeData
}

// isDuplicateError checks if an error is a duplicate creation error
func isDuplicateError(err error) bool {
	errLower := strings.ToLower(err.Error())
	return strings.Contains(errLower, "already exists") ||
		strings.Contains(errLower, "duplicate") ||
		strings.Contains(errLower, "conflict")
}

// createChallengeWithRetry attempts to create a challenge and handles duplicate errors
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

// handleNewChallenge handles creation or fetching of a new challenge
func handleNewChallenge(conf *config.Config, challengeConf config.ChallengeYaml, challenges []gzapi.Challenge, api *gzapi.GZAPI) (*gzapi.Challenge, error) {
	// If we already know it exists, fetch it directly.
	// Otherwise create; duplicate conflicts are handled by createChallengeWithRetry.
	if IsChallengeExist(challengeConf.Name, challenges) {
		challengeData, err := conf.Event.GetChallenge(challengeConf.Name)
		if err != nil {
			log.Error("Failed to get existing challenge %s: %v", challengeConf.Name, err)
			return nil, fmt.Errorf("get challenge %s: %w", challengeConf.Name, err)
		}
		challengeData.CS = api
		challengeData.GameId = conf.Event.Id
		return challengeData, nil
	}

	return createChallengeWithRetry(conf, challengeConf, api)
}

// handleExistingChallenge handles fetching an existing challenge from cache or API
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
	// ensure GameId is always set for downstream API calls (e.g., attachments)
	challengeData.GameId = conf.Event.Id
	// fix bug isEnable always be false after sync
	challengeData.IsEnabled = nil

	return challengeData, nil
}

// SyncOrchestrator manages the challenge synchronization process.
type SyncOrchestrator struct {
	conf              *config.Config
	challengeConf     config.ChallengeYaml
	challenges        []gzapi.Challenge
	api               *gzapi.GZAPI
	getCache          func(string, interface{}) error
	setCache          func(string, interface{}) error
	existingChallenge *gzapi.Challenge
	challengeData     *gzapi.Challenge
	err               error
}

// NewSyncOrchestrator creates a new orchestrator for syncing a challenge.
func NewSyncOrchestrator(conf *config.Config, challengeConf config.ChallengeYaml, challenges []gzapi.Challenge, api *gzapi.GZAPI, getCache func(string, interface{}) error, setCache func(string, interface{}) error, existingChallenge *gzapi.Challenge) *SyncOrchestrator {
	return &SyncOrchestrator{
		conf:              conf,
		challengeConf:     challengeConf,
		challenges:        challenges,
		api:               api,
		getCache:          getCache,
		setCache:          setCache,
		existingChallenge: existingChallenge,
	}
}

// Execute runs the synchronization process.
func (s *SyncOrchestrator) Execute() error {
	s.handle("determining sync path", s.determineSyncPath)
	s.handle("processing attachments and flags", s.processAttachmentsAndFlags)
	s.handle("building/pushing container image", s.prepareContainerImage)
	s.handle("merging and updating challenge", s.mergeAndupdate)

	if s.err != nil {
		log.Error("Failed to sync challenge '%s': %v", s.challengeConf.Name, s.err)
		return s.err
	}

	log.Info("✓ %s", s.challengeConf.Name)
	return nil
}

// handle wraps a function call with error checking.
func (s *SyncOrchestrator) handle(step string, fn func() error) {
	if s.err != nil {
		return
	}
	if err := fn(); err != nil {
		s.err = fmt.Errorf("step '%s' failed: %w", step, err)
	}
}

// determineSyncPath determines whether to create a new challenge or update an existing one.
func (s *SyncOrchestrator) determineSyncPath() error {
	var err error
	switch {
	case s.existingChallenge != nil:
		s.challengeData = s.existingChallenge
		s.challengeData.CS = s.api
		s.challengeData.IsEnabled = nil
	case !IsChallengeExist(s.challengeConf.Name, s.challenges):
		s.challengeData, err = handleNewChallenge(s.conf, s.challengeConf, s.challenges, s.api)
	default:
		if remote := findChallengeByTitle(s.challenges, s.challengeConf.Name); remote != nil {
			remote.CS = s.api
			remote.GameId = s.conf.Event.Id
			remote.IsEnabled = nil
			s.challengeData = remote
			return nil
		}
		s.challengeData, err = handleExistingChallenge(s.conf, s.challengeConf, s.api, s.getCache)
	}
	return err
}

// processAttachmentsAndFlags handles attachments and flags for the challenge.
func (s *SyncOrchestrator) processAttachmentsAndFlags() error {
	var err error
	s.challengeData, err = processAttachmentsAndFlags(s.conf, s.challengeConf, s.challengeData, s.api)
	return err
}

// mergeAndupdate merges challenge data and updates the challenge if needed.
func (s *SyncOrchestrator) mergeAndupdate() error {
	s.challengeData = MergeChallengeData(&s.challengeConf, s.challengeData)
	return updateChallengeIfNeeded(s.conf, &s.challengeConf, s.challengeData, s.getCache, s.setCache)
}

func (s *SyncOrchestrator) prepareContainerImage() error {
	// Only do anything when a registry is configured in appsettings.json.
	// This avoids making docker a hard requirement for sync in environments
	// that just reference prebuilt remote images.
	if s.conf == nil || s.conf.Appsettings == nil {
		return nil
	}
	reg := s.conf.Appsettings.RegistryConfig.ServerAddress
	if strings.TrimSpace(reg) == "" {
		return nil
	}

	// Only container challenges use container images.
	if !isContainerChallengeType(s.challengeConf.Type) {
		return nil
	}

	// If containerImage is already a remote-ish image reference (contains a '/')
	// and it doesn't resolve to a local path, leave it alone.
	ci := strings.TrimSpace(s.challengeConf.Container.ContainerImage)
	if ci != "" && strings.Contains(ci, "/") && !containerImageResolvesToLocalPath(s.challengeConf.Cwd, ci) {
		return nil
	}

	slug := config.GenerateSlug(s.conf.EventName, s.challengeConf.Category, s.challengeConf.Name)
	localTag := fmt.Sprintf("%s:latest", slug)

	// Build local image first.
	buildDir, dockerfile := resolveDockerBuildContext(s.challengeConf.Cwd, s.challengeConf.Container.ContainerImage)
	log.InfoH3("Building image for %s: %s (context=%s)", s.challengeConf.Name, localTag, buildDir)
	buildCtx, cancelBuild := context.WithTimeout(context.Background(), getDockerBuildTimeout())
	defer cancelBuild()
	if err := dockerBuild(buildCtx, buildDir, dockerfile, localTag); err != nil {
		return err
	}

	// Push to registry, then make containerImage point at the pushed tag.
	repoPrefix, loginServer := parseRegistryServerAddress(reg)
	if repoPrefix == "" || loginServer == "" {
		return fmt.Errorf("invalid registry server address in appsettings.json: %q", reg)
	}

	remoteTag := fmt.Sprintf("%s/%s:latest", repoPrefix, slug)

	if strings.TrimSpace(s.conf.Appsettings.RegistryConfig.UserName) != "" {
		log.InfoH3("Logging in to registry: %s", loginServer)
		loginCtx, cancelLogin := context.WithTimeout(context.Background(), getDockerLoginTimeout())
		defer cancelLogin()
		if err := dockerLoginOnce(loginCtx, loginServer, s.conf.Appsettings.RegistryConfig.UserName, s.conf.Appsettings.RegistryConfig.Password); err != nil {
			return err
		}
	}

	log.InfoH3("Tagging image: %s -> %s", localTag, remoteTag)
	tagCtx, cancelTag := context.WithTimeout(context.Background(), getDockerTagTimeout())
	defer cancelTag()
	if err := dockerTag(tagCtx, localTag, remoteTag); err != nil {
		return err
	}

	log.InfoH3("Pushing image: %s", remoteTag)
	pushCtx, cancelPush := context.WithTimeout(context.Background(), getDockerPushTimeout())
	defer cancelPush()
	if err := dockerPush(pushCtx, remoteTag); err != nil {
		return err
	}

	// Ensure the challenge config synced to the API points at the registry image.
	s.challengeConf.Container.ContainerImage = remoteTag
	return nil
}

// SyncChallenge synchronizes a single challenge.
func SyncChallenge(conf *config.Config, challengeConf config.ChallengeYaml, challenges []gzapi.Challenge, api *gzapi.GZAPI, getCache func(string, interface{}) error, setCache func(string, interface{}) error) error {
	return NewSyncOrchestrator(conf, challengeConf, challenges, api, getCache, setCache, nil).Execute()
}

// SyncChallengeWithExisting syncs a challenge with an optional existing challenge to force update mode.
func SyncChallengeWithExisting(conf *config.Config, challengeConf config.ChallengeYaml, challenges []gzapi.Challenge, api *gzapi.GZAPI, getCache func(string, interface{}) error, setCache func(string, interface{}) error, existingChallenge *gzapi.Challenge) error {
	return NewSyncOrchestrator(conf, challengeConf, challenges, api, getCache, setCache, existingChallenge).Execute()
}

// processAttachmentsAndFlags handles attachments and flags for a challenge
func processAttachmentsAndFlags(conf *config.Config, challengeConf config.ChallengeYaml, challengeData *gzapi.Challenge, api *gzapi.GZAPI) (*gzapi.Challenge, error) {
	refresher := func() (*gzapi.Challenge, error) {
		return conf.Event.GetChallenge(challengeConf.Name)
	}
	return processAttachmentsAndFlagsWithHandlers(conf, challengeConf, challengeData, api, refresher, HandleChallengeAttachments, UpdateChallengeFlags)
}

func processAttachmentsAndFlagsWithHandlers(conf *config.Config, challengeConf config.ChallengeYaml, challengeData *gzapi.Challenge, api *gzapi.GZAPI, refresher challengeRefresher, attach attachmentHandler, updateFlags flagHandler) (*gzapi.Challenge, error) {
	current, err := ensureFreshChallengeData(conf, challengeConf, api, challengeData, refresher)
	if err != nil {
		return nil, err
	}

	if err := attach(challengeConf, current, api); err != nil {
		log.Error("Failed to handle attachments for %s (game %d challenge %d): %v", challengeConf.Name, current.GameId, current.Id, err)
		if isNotFoundError(err) {
			refreshed, refreshErr := refresher()
			if refreshErr != nil {
				return nil, fmt.Errorf("attachment handling failed for %s: %w", challengeConf.Name, err)
			}
			refreshed.CS = api
			refreshed.GameId = conf.Event.Id
			refreshed.IsEnabled = nil

			if retryErr := attach(challengeConf, refreshed, api); retryErr != nil {
				log.Error("Retry attachment failed for %s (game %d challenge %d): %v", challengeConf.Name, refreshed.GameId, refreshed.Id, retryErr)
				return nil, fmt.Errorf("attachment handling failed for %s after refresh: %w", challengeConf.Name, retryErr)
			}
			current = refreshed
		} else {
			return nil, fmt.Errorf("attachment handling failed for %s: %w", challengeConf.Name, err)
		}
	}

	if err := updateFlags(conf, challengeConf, current); err != nil {
		log.Error("Failed to update flags for %s: %v", challengeConf.Name, err)
		return nil, fmt.Errorf("update flags for %s: %w", challengeConf.Name, err)
	}

	return current, nil
}

// updateChallengeWithRetry attempts to update a challenge and retries on 404
func updateChallengeWithRetry(conf *config.Config, challengeConf *config.ChallengeYaml, challengeData *gzapi.Challenge) (*gzapi.Challenge, error) {
	updatedData, err := challengeData.Update(*challengeData)
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

// updateChallengeIfNeeded updates the challenge if configuration has changed
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
