package gzcli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/dimasma0305/gzcli/internal/gzcli/challenge"
	"github.com/dimasma0305/gzcli/internal/gzcli/common"
	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/container"
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
	appsettings *AppSettings `yaml:"-"`
	Appsettings *AppSettings `yaml:"-"` // Public field for external access
}

// ToConfigPackage converts to config.Config
func (c *Config) ToConfigPackage() *config.Config {
	settings := c.Appsettings
	if settings == nil {
		settings = c.appsettings
	}
	return &config.Config{
		Url:         c.Url,
		Creds:       c.Creds,
		Event:       c.Event,
		Appsettings: settings,
	}
}

// FromConfigPackage converts from config.Config
func FromConfigPackage(conf *config.Config) *Config {
	settings := conf.Appsettings
	return &Config{
		Url:         conf.Url,
		Creds:       conf.Creds,
		Event:       conf.Event,
		appsettings: settings,
		Appsettings: settings,
	}
}

// SetAppSettings sets both appsettings fields
func (c *Config) SetAppSettings(settings *AppSettings) {
	c.appsettings = settings
	c.Appsettings = settings
}

// GetAppSettingsField returns the settings
func (c *Config) GetAppSettingsField() *AppSettings {
	if c.Appsettings != nil {
		return c.Appsettings
	}
	return c.appsettings
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
	conf := &Config{}
	conf.SetAppSettings(appsettings)
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
	log.Info("Starting sync process...")

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

	// Create container with dependencies
	cnt := container.NewContainer(container.ContainerConfig{
		Config:      conf,
		API:         gz.api,
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
		return gz.Sync()
	}
	log.Info("Found current game: %s (ID: %d)", currentGame.Title, currentGame.Id)

	if gz.UpdateGame {
		log.Info("Updating game configuration...")
		ctx := context.Background()
		if err := gameSvc.UpdateGameIfNeeded(ctx, conf, currentGame, createPosterIfNotExistOrDifferent, setCache); err != nil {
			log.Error("Failed to update game: %v", err)
			return fmt.Errorf("game update error: %w", err)
		}
		log.Info("Game updated successfully")
	}

	log.Info("Validating challenges...")
	validator := common.NewValidator()
	for _, challengeConf := range challengesConf {
		if err := validator.ValidateChallenge(challengeConf); err != nil {
			log.Error("Challenge validation failed for %s: %v", challengeConf.Name, err)
			return fmt.Errorf("validation error: %w", err)
		}
	}
	log.Info("All challenges validated successfully")

	// Get fresh challenges list
	log.Info("Fetching existing challenges from API...")
	conf.Event.CS = gz.api
	challenges, err := conf.Event.GetChallenges()
	if err != nil {
		log.Error("Failed to get challenges from API: %v", err)
		return fmt.Errorf("API challenges fetch error: %w", err)
	}
	log.Info("Found %d existing challenges in API", len(challenges))

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
			ctx := context.Background()
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

	// Return first error if any
	select {
	case err := <-errChan:
		return err
	default:
		log.Info("All challenges synced successfully!")
		return nil
	}
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

// MustRunScripts executes scripts or fatally logs error
// Deprecated: Use RunScripts directly with event parameter
func MustRunScripts(script string, eventName string) {
	if err := RunScripts(script, eventName); err != nil {
		log.Fatal("Script execution failed: ", err)
	}
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
	conf, err := getConfigWrapper(gz.api)
	if err != nil {
		return fmt.Errorf("failed to get config")
	}

	csvData, err := team.GetData(csvURL)
	if err != nil {
		return fmt.Errorf("failed to get CSV data")
	}

	// Load existing team credentials from cache
	var teamsCredsCache []*team.TeamCreds
	if err := GetCache("teams_creds", &teamsCredsCache); err != nil {
		log.Error("%s", err.Error())
	}

	// Create config adapter
	configAdapter := &teamConfigAdapter{conf: conf}

	err = team.ParseCSV(csvData, configAdapter, teamsCredsCache, isSendEmail,
		team.CreateTeamAndUser,
		generateUsername,
		setCache)
	if err != nil {
		return err
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
	return &appSettingsAdapter{settings: t.conf.GetAppSettingsField()}
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
