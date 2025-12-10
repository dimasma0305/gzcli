package gzcli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/dimasma0305/gzcli/internal/gzcli/challenge"
	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/event"
	"github.com/dimasma0305/gzcli/internal/gzcli/gzapi"
	"github.com/dimasma0305/gzcli/internal/gzcli/team"
	"github.com/dimasma0305/gzcli/internal/gzcli/watcher"
	"github.com/dimasma0305/gzcli/internal/log"
)

// ChallengeYaml is a type alias for backward compatibility with watcher.go
type ChallengeYaml = config.ChallengeYaml

// Container is a type alias for backward compatibility with watcher.go
type Container = config.Container

// ScriptValue is a type alias for backward compatibility with watcher.go
type ScriptValue = config.ScriptValue

// ScriptConfig is a type alias for backward compatibility with watcher.go
type ScriptConfig = config.ScriptConfig

// AppSettings is a type alias for backward compatibility with watcher.go
type AppSettings = config.AppSettings

// Dashboard is a type alias for backward compatibility with watcher.go
type Dashboard = config.Dashboard

// Watcher types for backward compatibility
type (
	// Watcher is the main watcher instance
	Watcher = watcher.Watcher

	// WatcherConfig holds configuration for the watcher
	WatcherConfig = watcher.WatcherConfig

	// WatcherClient provides client interface for the watcher daemon
	WatcherClient = watcher.WatcherClient
)

// DefaultWatcherConfig provides default watcher configuration
var DefaultWatcherConfig = watcher.DefaultWatcherConfig

// NewWatcher creates a new file watcher instance for backward compatibility
func NewWatcher(gz *GZ) (*Watcher, error) {
	return watcher.NewWatcher(gz.api)
}

// NewWatcherClient creates a new watcher client
func NewWatcherClient(socketPath string) *WatcherClient {
	return watcher.NewWatcherClient(socketPath)
}

// Config is a compatibility wrapper that allows lowercase appsettings field for watcher.go
type Config struct {
	Url         string       `yaml:"url"` //nolint:revive // Field name required for watcher.go compatibility
	Creds       gzapi.Creds  `yaml:"creds"`
	Event       gzapi.Game   `yaml:"event"`
	AppSettings *AppSettings `yaml:"-"`
}

// ToConfigPackage converts to config.Config
func (c *Config) ToConfigPackage() *config.Config {
	return &config.Config{
		Url:         c.Url,
		Creds:       c.Creds,
		Event:       c.Event,
		Appsettings: c.AppSettings,
	}
}

// FromConfigPackage converts from config.Config
func FromConfigPackage(conf *config.Config) *Config {
	return &Config{
		Url:         conf.Url,
		Creds:       conf.Creds,
		Event:       conf.Event,
		AppSettings: conf.Appsettings,
	}
}

// Compatibility functions for watcher.go

// GetConfig retrieves the application configuration
func GetConfig(api *gzapi.GZAPI) (*Config, error) {
	conf, err := getConfigWrapper(api)
	if err != nil {
		return nil, err
	}
	return FromConfigPackage(conf), nil
}

// GetChallengesYaml retrieves challenge configurations from YAML files
func GetChallengesYaml(conf *Config) ([]ChallengeYaml, error) {
	return config.GetChallengesYaml(conf.ToConfigPackage())
}

// DefaultScriptTimeout is the default timeout for script execution
const DefaultScriptTimeout = challenge.DefaultScriptTimeout

// GZ is the main application struct for GZCTF CLI operations
type GZ struct {
	api        *gzapi.GZAPI
	UpdateGame bool
	watcher    *watcher.Watcher
	eventName  string // Store the event name for this instance
}

// Cache frequently used paths and configurations
var (
	workDirOnce   sync.Once
	cachedWorkDir string
)

const (
	gzctfDir = ".gzctf"
)

// getWorkDir returns the cached working directory
func getWorkDir() string {
	workDirOnce.Do(func() {
		cachedWorkDir, _ = os.Getwd()
	})
	return cachedWorkDir
}

// Optimized database query execution with prepared command
var dbQueryCmd = exec.Command(
	"docker", "compose", "exec", "-T", "db", "psql",
	"--user", "postgres", "-d", "gzctf", "-c",
)

func runDBQuery(query string) error {
	cmd := *dbQueryCmd // Copy base command
	cmd.Args = append(cmd.Args, query)
	cmd.Dir = filepath.Join(getWorkDir(), gzctfDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Error("Database query failed: %v", err)
		return err
	}
	return nil
}

// Init initializes the GZ instance with configuration and API client
// Uses the default event selection mechanism
func Init() (*GZ, error) {
	return InitWithEvent("")
}

// InitWithEvent initializes the GZ instance with a specific event
// If eventName is empty, it will be auto-detected
func InitWithEvent(eventName string) (*GZ, error) {
	// Note: Since we're using memoization, we create fresh instances
	// This allows commands to work with different events
	conf, err := config.GetConfigWithEvent(&gzapi.GZAPI{}, eventName, GetCache, setCache, deleteCacheWrapper, createNewGameWrapper)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	api, err := gzapi.Init(conf.Url, &conf.Creds)
	if err == nil {
		return &GZ{api: api, eventName: conf.EventName}, nil
	}

	// Fallback to registration
	api, err = gzapi.Register(conf.Url, &gzapi.RegisterForm{
		Email:    "admin@localhost",
		Username: conf.Creds.Username,
		Password: conf.Creds.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("registration failed: %w", err)
	}

	if err := runDBQuery(fmt.Sprintf(
		`UPDATE "AspNetUsers" SET "Role"=3 WHERE "UserName"='%s';`,
		conf.Creds.Username,
	)); err != nil {
		return nil, err
	}

	return &GZ{api: api, eventName: conf.EventName}, nil
}

// GenerateStructure generates challenge directory structure from templates
func (gz *GZ) GenerateStructure() error {
	appsettings, err := config.GetAppSettings()
	if err != nil {
		return err
	}
	conf := &Config{
		AppSettings: appsettings,
	}
	challenges, err := config.GetChallengesYaml(conf.ToConfigPackage())
	if err != nil {
		return err
	}

	// Convert to interface for structure package
	challengeData := make([]challengeDataImpl, len(challenges))
	for i, c := range challenges {
		challengeData[i] = challengeDataImpl{c}
	}

	// Call genStructure with the provided challenges
	challengeInterfaces := make([]interface{ GetCwd() string }, len(challengeData))
	for i := range challengeData {
		challengeInterfaces[i] = challengeData[i]
	}

	return genStructureWrapper(challengeInterfaces)
}

// RemoveAllEvent removes all events/games with parallel execution
func (gz *GZ) RemoveAllEvent() error {
	return event.RemoveAllEvent(gz.api)
}

// Scoreboard2CTFTimeFeed converts scoreboard to CTFTime feed format
func (gz *GZ) Scoreboard2CTFTimeFeed() (*event.CTFTimeFeed, error) {
	conf, err := getConfigWrapper(gz.api)
	if err != nil {
		return nil, err
	}

	return event.Scoreboard2CTFTimeFeed(&conf.Event)
}

// Sync synchronizes challenges from local configuration to the GZCTF server
func (gz *GZ) Sync() error {
	return gz.syncWithRetry(0)
}

// syncWithRetry is the internal sync implementation with retry logic
func (gz *GZ) syncWithRetry(retryCount int) error {
	const maxRetries = 2 // Prevent infinite recursion

	// Step 1: Get configuration
	conf, err := config.GetConfigWithEvent(gz.api, gz.eventName, GetCache, setCache, deleteCacheWrapper, createNewGameWrapper)
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	// Step 2: Get local challenge configurations
	challengesConf, err := config.GetChallengesYaml(conf)
	if err != nil {
		return fmt.Errorf("challenges config error: %w", err)
	}

	// Step 3: Find the current game on the server
	games, err := gz.api.GetGames()
	if err != nil {
		return fmt.Errorf("games fetch error: %w", err)
	}

	currentGame := challenge.FindCurrentGame(games, conf.Event.Title, gz.api)
	if currentGame == nil {
		if retryCount >= maxRetries {
			log.Error("Game '%s' not found after %d retries", conf.Event.Title, maxRetries)
			return fmt.Errorf("game '%s' not found", conf.Event.Title)
		}
		_ = DeleteCache(fmt.Sprintf("config-%s", gz.eventName))
		return gz.syncWithRetry(retryCount + 1)
	}

	// Step 4: Update game if needed
	if gz.UpdateGame {
		if err := challenge.UpdateGameIfNeeded(conf, currentGame, gz.api, createPosterIfNotExistOrDifferent, setCache); err != nil {
			return fmt.Errorf("game update error: %w", err)
		}
	}

	// Step 5: Validate local challenges
	if err := challenge.ValidateChallenges(challengesConf); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Step 6: Get remote challenges
	conf.Event.CS = gz.api
	remoteChallenges, err := conf.Event.GetChallenges()
	if err != nil {
		return fmt.Errorf("API challenges fetch error: %w", err)
	}

	remoteChallenges, deleted, err := challenge.RemoveDuplicateChallenges(remoteChallenges, nil)
	if err != nil {
		return fmt.Errorf("duplicate challenge cleanup error: %w", err)
	}
	if deleted {
		remoteChallenges, err = conf.Event.GetChallenges()
		if err != nil {
			return fmt.Errorf("API challenges refetch error after cleanup: %w", err)
		}
	}

	// Step 7: Process all challenges concurrently
	return gz.processChallenges(conf, challengesConf, remoteChallenges)
}

// processChallenges handles the concurrent processing of challenges
func (gz *GZ) processChallenges(conf *config.Config, challengesConf []config.ChallengeYaml, remoteChallenges []gzapi.Challenge) error {
	log.Info("Syncing %d challenges...", len(challengesConf))
	var wg sync.WaitGroup
	errChan := make(chan error, len(challengesConf))
	var successCount, failureCount int32

	challengeMutexes := make(map[string]*sync.Mutex)
	var mutexesMu sync.RWMutex

	for _, localChallenge := range challengesConf {
		wg.Add(1)
		go func(c config.ChallengeYaml) {
			defer wg.Done()

			mutexesMu.Lock()
			if challengeMutexes[c.Name] == nil {
				challengeMutexes[c.Name] = &sync.Mutex{}
			}
			mutex := challengeMutexes[c.Name]
			mutexesMu.Unlock()

			mutex.Lock()
			defer mutex.Unlock()

			log.Info("Processing challenge: %s", c.Name)
			if err := challenge.SyncChallenge(conf, c, remoteChallenges, gz.api, GetCache, setCache); err != nil {
				log.Error("Failed to sync challenge %s: %v", c.Name, err)
				errChan <- fmt.Errorf("challenge sync failed for %s: %w", c.Name, err)
				atomic.AddInt32(&failureCount, 1)
			} else {
				log.Info("Successfully synced challenge: %s", c.Name)
				atomic.AddInt32(&successCount, 1)
			}
		}(localChallenge)
	}

	wg.Wait()
	close(errChan)

	log.Info("Sync completed. Success: %d, Failures: %d", successCount, failureCount)

	// Return the first error encountered, if any
	if err := <-errChan; err != nil {
		return err
	}

	log.Info("All challenges synced successfully!")
	return nil
}

// MustInit initializes GZ or fatally logs error
func MustInit() *GZ {
	gz, err := Init()
	if err != nil {
		log.Fatal("Initialization failed: ", err)
	}
	return gz
}

// MustSync synchronizes data or fatally logs error
func (gz *GZ) MustSync() {
	if err := gz.Sync(); err != nil {
		log.Fatal("Sync failed: ", err)
	}
}

// MustScoreboard2CTFTimeFeed converts scoreboard or fatally logs error
func (gz *GZ) MustScoreboard2CTFTimeFeed() *event.CTFTimeFeed {
	feed, err := gz.Scoreboard2CTFTimeFeed()
	if err != nil {
		log.Fatal("Scoreboard generation failed: ", err)
	}
	return feed
}

// MustCreateTeams creates teams or fatally logs error
func (gz *GZ) MustCreateTeams(url string, sendEmail bool) {
	if err := gz.CreateTeams(url, sendEmail); err != nil {
		log.Fatal("Team creation failed: ", err)
	}
}

// MustDeleteAllUser removes all users or fatally logs error
func (gz *GZ) MustDeleteAllUser() {
	if err := gz.DeleteAllUser(); err != nil {
		log.Fatal("User deletion failed: ", err)
	}
}

// StartWatcher starts the file watcher service
func (gz *GZ) StartWatcher(config watcher.WatcherConfig) error {
	if gz.watcher != nil && gz.watcher.IsWatching() {
		return fmt.Errorf("watcher is already running")
	}

	w, err := watcher.NewWatcher(gz.api)
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}

	if err := w.Start(config); err != nil {
		return fmt.Errorf("failed to start watcher: %w", err)
	}

	gz.watcher = w
	return nil
}

// StopWatcher stops the file watcher service
func (gz *GZ) StopWatcher() error {
	if gz.watcher == nil {
		return fmt.Errorf("no watcher is running")
	}

	if err := gz.watcher.Stop(); err != nil {
		return fmt.Errorf("failed to stop watcher: %w", err)
	}

	gz.watcher = nil
	return nil
}

// IsWatcherRunning returns true if the watcher is currently running
func (gz *GZ) IsWatcherRunning() bool {
	return gz.watcher != nil && gz.watcher.IsWatching()
}

// GetWatcherStatus returns the status of the watcher service
func (gz *GZ) GetWatcherStatus() map[string]interface{} {
	status := map[string]interface{}{
		"running": gz.IsWatcherRunning(),
	}

	if gz.watcher != nil {
		status["watched_challenges"] = gz.watcher.GetWatchedChallenges()
	} else {
		status["watched_challenges"] = []string{}
	}

	return status
}

// MustStartWatcher starts the watcher or fatally logs error
func (gz *GZ) MustStartWatcher(config watcher.WatcherConfig) {
	if err := gz.StartWatcher(config); err != nil {
		log.Fatal("Failed to start watcher: ", err)
	}
}

// MustStopWatcher stops the watcher or fatally logs error
func (gz *GZ) MustStopWatcher() {
	if err := gz.StopWatcher(); err != nil {
		log.Fatal("Failed to stop watcher: ", err)
	}
}

// CreateTeams creates teams from a CSV file
func (gz *GZ) CreateTeams(csvURL string, isSendEmail bool) error {
	// Step 1: Get configuration
	conf, err := getConfigWrapper(gz.api)
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	// Step 2: Get CSV data from URL
	csvData, err := team.GetData(csvURL)
	if err != nil {
		return fmt.Errorf("failed to get CSV data: %w", err)
	}

	// Step 3: Load existing team credentials from cache
	var teamsCredsCache []*team.TeamCreds
	if err := GetCache("teams_creds", &teamsCredsCache); err != nil {
		log.Info("Could not load team credentials cache: %v", err)
	}

	// Step 4: Parse CSV and create teams
	configAdapter := &teamConfigAdapter{conf: conf}
	if err := team.ParseCSV(csvData, configAdapter, teamsCredsCache, isSendEmail, team.CreateTeamAndUser, generateUsername, setCache); err != nil {
		return fmt.Errorf("failed to parse CSV and create teams: %w", err)
	}

	return nil
}

// Helper type adapters and wrappers
type challengeDataImpl struct {
	config.ChallengeYaml
}

func (c challengeDataImpl) GetCwd() string {
	return c.Cwd
}

type teamConfigAdapter struct {
	conf *config.Config
}

func (t *teamConfigAdapter) GetUrl() string { //nolint:revive // Method name required by team.ConfigInterface
	return t.conf.Url
}

func (t *teamConfigAdapter) GetEventId() int { //nolint:revive // Method name required by team.ConfigInterface
	return t.conf.Event.Id
}

func (t *teamConfigAdapter) GetEventTitle() string {
	return t.conf.Event.Title
}

func (t *teamConfigAdapter) GetInviteCode() string {
	return t.conf.Event.InviteCode
}

func (t *teamConfigAdapter) GetAppSettings() team.AppSettingsInterface {
	return &appSettingsAdapter{settings: t.conf.Appsettings}
}

type appSettingsAdapter struct {
	settings *config.AppSettings
}

func (a *appSettingsAdapter) GetEmailConfig() team.EmailConfig {
	return team.EmailConfig{
		UserName: a.settings.EmailConfig.UserName,
		Password: a.settings.EmailConfig.Password,
		SMTP: struct {
			Host string
			Port int
		}{
			Host: a.settings.EmailConfig.Smtp.Host,
			Port: a.settings.EmailConfig.Smtp.Port,
		},
	}
}
