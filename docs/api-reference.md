# Internal API Reference

This document provides reference documentation for gzcli's internal packages and APIs.

## Table of Contents

- [Core Packages](#core-packages)
- [gzcli Package](#gzcli-package)
- [gzapi Package](#gzapi-package)
- [watcher Package](#watcher-package)
- [challenge Package](#challenge-package)
- [team Package](#team-package)
- [config Package](#config-package)
- [event Package](#event-package)
- [utils Package](#utils-package)

## Core Packages

### Package Organization

```
internal/gzcli/
├── gzapi/          # API client for GZCTF
├── watcher/        # File watching system
├── challenge/      # Challenge management
├── team/           # Team operations
├── config/         # Configuration management
├── event/          # Event system (webhooks)
├── structure/      # Directory structure management
├── script/         # Script execution
└── utils/          # Utility functions
```

## gzcli Package

Main package that coordinates all operations.

### Type: GZ

The central type that manages application state:

```go
type GZ struct {
    Conf   *config.Config
    Api    *gzapi.GZAPI
    Event  *event.Event
    // ... other fields
}
```

#### Methods

**`NewGZ(configPath string) (*GZ, error)`**

Creates a new GZ instance with the specified configuration.

```go
gz, err := gzcli.NewGZ(".gzctf/conf.yaml")
if err != nil {
    log.Fatal(err)
}
```

**`Initialize() error`**

Initializes the GZ instance (loads config, creates API client).

```go
if err := gz.Initialize(); err != nil {
    log.Fatal(err)
}
```

**`Sync(options SyncOptions) error`**

Synchronizes challenges to the platform.

```go
err := gz.Sync(gzcli.SyncOptions{
    UpdateGame: true,
    DryRun:     false,
})
```

## gzapi Package

HTTP client for GZ::CTF API.

### Type: GZAPI

```go
type GZAPI struct {
    Url    string
    Client *resty.Client
    // ... session management fields
}
```

### Authentication

**`Login(username, password string) error`**

Authenticates with the GZCTF platform.

```go
api := gzapi.New("https://ctf.example.com")
err := api.Login("admin", "password")
```

**`Logout() error`**

Logs out and invalidates session.

```go
defer api.Logout()
```

### Game Operations

**`GetGames() ([]Game, error)`**

Retrieves all games from the platform.

```go
games, err := api.GetGames()
for _, game := range games {
    fmt.Printf("Game: %s (ID: %d)\n", game.Title, game.ID)
}
```

**`CreateGame(game *Game) (*Game, error)`**

Creates a new game.

```go
game := &gzapi.Game{
    Title:       "My CTF",
    Description: "CTF description",
}
created, err := api.CreateGame(game)
```

**`UpdateGame(gameID int, game *Game) error`**

Updates an existing game.

```go
err := api.UpdateGame(game.ID, updatedGame)
```

**`DeleteGame(gameID int) error`**

Deletes a game.

```go
err := api.DeleteGame(gameID)
```

### Challenge Operations

**`GetChallenges(gameID int) ([]Challenge, error)`**

Gets all challenges for a game.

```go
challenges, err := api.GetChallenges(gameID)
```

**`CreateChallenge(gameID int, challenge *Challenge) (*Challenge, error)`**

Creates a new challenge.

```go
challenge := &gzapi.Challenge{
    Title:    "XSS Challenge",
    Category: "Web",
    Tags:     []string{"web", "xss"},
}
created, err := api.CreateChallenge(gameID, challenge)
```

**`UpdateChallenge(challengeID int, challenge *Challenge) error`**

Updates an existing challenge.

```go
err := api.UpdateChallenge(challengeID, updated)
```

**`DeleteChallenge(challengeID int) error`**

Deletes a challenge.

```go
err := api.DeleteChallenge(challengeID)
```

**`UploadAttachment(challengeID int, file io.Reader, filename string) error`**

Uploads a file attachment to a challenge.

```go
f, _ := os.Open("challenge.zip")
defer f.Close()
err := api.UploadAttachment(challengeID, f, "challenge.zip")
```

### Team Operations

**`GetTeams() ([]Team, error)`**

Retrieves all teams.

```go
teams, err := api.GetTeams()
```

**`CreateTeam(team *Team) (*Team, error)`**

Creates a new team.

```go
team := &gzapi.Team{
    Name:  "Team Alpha",
    Bio:   "Team description",
}
created, err := api.CreateTeam(team)
```

**`DeleteTeam(teamID int) error`**

Deletes a team.

```go
err := api.DeleteTeam(teamID)
```

### User Operations

**`CreateUser(user *User) (*User, error)`**

Creates a new user.

```go
user := &gzapi.User{
    Username: "player1",
    Email:    "player1@example.com",
    Password: "secure_password",
}
created, err := api.CreateUser(user)
```

**`DeleteUser(userID int) error`**

Deletes a user.

```go
err := api.DeleteUser(userID)
```

## watcher Package

File watching and auto-sync functionality.

### Type: Watcher

```go
type Watcher struct {
    gz         *gzcli.GZ
    watcher    *fsnotify.Watcher
    debouncer  *time.Timer
    // ... other fields
}
```

### Methods

**`NewWatcher(gz *gzcli.GZ, options WatcherOptions) (*Watcher, error)`**

Creates a new file watcher.

```go
watcher, err := watcher.NewWatcher(gz, watcher.WatcherOptions{
    Debounce: 2 * time.Second,
    IgnorePatterns: []string{"*.tmp", ".git/"},
})
```

**`Start() error`**

Starts watching for file changes.

```go
if err := watcher.Start(); err != nil {
    log.Fatal(err)
}
```

**`Stop() error`**

Stops the watcher.

```go
defer watcher.Stop()
```

**`StartDaemon() error`**

Starts the watcher as a background daemon.

```go
err := watcher.StartDaemon()
```

**`StopDaemon() error`**

Stops the daemon.

```go
err := watcher.StopDaemon()
```

**`Status() (*WatcherStatus, error)`**

Gets the current daemon status.

```go
status, err := watcher.Status()
fmt.Printf("Running: %v\n", status.Running)
```

### Type: WatcherOptions

Configuration options for the watcher.

```go
type WatcherOptions struct {
    Debounce       time.Duration
    IgnorePatterns []string
    Foreground     bool
}
```

## challenge Package

Challenge parsing, validation, and synchronization.

### Type: Challenge

```go
type Challenge struct {
    Name        string
    Title       string
    Category    string
    Description string
    Tags        []string
    Hints       []Hint
    Flags       []Flag
    // ... other fields
}
```

### Functions

**`LoadChallenge(path string) (*Challenge, error)`**

Loads a challenge from a directory.

```go
challenge, err := challenge.LoadChallenge("./challenges/web/xss")
```

**`ParseYAML(data []byte) (*Challenge, error)`**

Parses challenge YAML configuration.

```go
data, _ := os.ReadFile("challenge.yaml")
challenge, err := challenge.ParseYAML(data)
```

**`Validate(challenge *Challenge) error`**

Validates a challenge configuration.

```go
if err := challenge.Validate(ch); err != nil {
    log.Fatal("Invalid challenge:", err)
}
```

**`SyncChallenge(gz *gzcli.GZ, challengePath string) error`**

Synchronizes a single challenge.

```go
err := challenge.SyncChallenge(gz, "./challenges/web/xss")
```

## team Package

Team management and bulk operations.

### Functions

**`CreateTeamsFromCSV(gz *gzcli.GZ, csvPath string, sendEmail bool) error`**

Creates teams from a CSV file.

```go
err := team.CreateTeamsFromCSV(gz, "teams.csv", true)
```

**`DeleteAllTeams(gz *gzcli.GZ) error`**

Deletes all teams and associated users.

```go
err := team.DeleteAllTeams(gz)
```

**`ParseCSV(csvPath string) ([]TeamRecord, error)`**

Parses team data from CSV.

```go
records, err := team.ParseCSV("teams.csv")
```

### Type: TeamRecord

```go
type TeamRecord struct {
    Name     string
    Email    string
    Password string
    Members  []string
}
```

## config Package

Configuration file parsing and management.

### Type: Config

```go
type Config struct {
    URL         string
    PublicEntry string
    Creds       Credentials
    Event       EventConfig
    Watcher     WatcherConfig
}
```

### Functions

**`Load(path string) (*Config, error)`**

Loads configuration from file.

```go
cfg, err := config.Load(".gzctf/conf.yaml")
```

**`Validate(cfg *Config) error`**

Validates configuration.

```go
if err := config.Validate(cfg); err != nil {
    log.Fatal("Invalid config:", err)
}
```

**`Save(cfg *Config, path string) error`**

Saves configuration to file.

```go
err := config.Save(cfg, ".gzctf/conf.yaml")
```

## event Package

Event handling and webhook notifications.

### Type: Event

```go
type Event struct {
    DiscordWebhook string
    // ... other webhooks
}
```

### Methods

**`NewEvent(config EventConfig) *Event`**

Creates a new event handler.

```go
event := event.NewEvent(config.Event)
```

**`OnSync(challenge *Challenge) error`**

Sends notification when challenge is synced.

```go
err := event.OnSync(challenge)
```

**`OnError(err error) error`**

Sends error notification.

```go
event.OnError(syncError)
```

**`SendDiscord(message string) error`**

Sends a Discord webhook message.

```go
err := event.SendDiscord("Challenge synced successfully!")
```

## utils Package

Common utility functions.

### File Operations

**`FileExists(path string) bool`**

Checks if a file exists.

```go
if utils.FileExists("challenge.yaml") {
    // File exists
}
```

**`DirExists(path string) bool`**

Checks if a directory exists.

```go
if utils.DirExists("challenges") {
    // Directory exists
}
```

**`CopyFile(src, dst string) error`**

Copies a file.

```go
err := utils.CopyFile("source.txt", "dest.txt")
```

**`CopyDir(src, dst string) error`**

Recursively copies a directory.

```go
err := utils.CopyDir("./challenges", "./backup")
```

### String Operations

**`Slugify(s string) string`**

Converts a string to a URL-safe slug.

```go
slug := utils.Slugify("My Challenge Name") // "my-challenge-name"
```

**`Truncate(s string, maxLen int) string`**

Truncates a string to maximum length.

```go
short := utils.Truncate("Long description...", 50)
```

### Path Operations

**`FindChallengeRoot(path string) (string, error)`**

Finds the challenge root directory by looking for challenge.yaml.

```go
root, err := utils.FindChallengeRoot("./challenges/web/xss/src")
// Returns: "./challenges/web/xss"
```

**`RelativePath(base, target string) (string, error)`**

Computes relative path from base to target.

```go
rel, err := utils.RelativePath("/home/user", "/home/user/challenges")
// Returns: "challenges"
```

## Error Handling

### Common Error Types

**`ErrNotFound`**

Resource not found error.

```go
if errors.Is(err, gzapi.ErrNotFound) {
    // Handle not found
}
```

**`ErrUnauthorized`**

Authentication failed error.

```go
if errors.Is(err, gzapi.ErrUnauthorized) {
    // Re-authenticate
}
```

**`ErrInvalidConfig`**

Configuration validation error.

```go
if errors.Is(err, config.ErrInvalidConfig) {
    // Fix configuration
}
```

### Error Wrapping

Errors are wrapped with context using `fmt.Errorf` with `%w`:

```go
if err != nil {
    return fmt.Errorf("syncing challenge %s: %w", name, err)
}
```

Unwrap with `errors.Unwrap()` or `errors.Is()`:

```go
if errors.Is(err, gzapi.ErrUnauthorized) {
    // Handle authorization error
}
```

## Best Practices

### Using the API Client

1. **Always check for errors**
   ```go
   if err := api.Login(user, pass); err != nil {
       return err
   }
   ```

2. **Reuse API client instances**
   ```go
   api := gzapi.New(url)
   // Use same instance for multiple operations
   ```

3. **Handle rate limiting**
   ```go
   // API client handles rate limiting automatically
   // But implement backoff for bulk operations
   ```

### Configuration Management

1. **Validate before use**
   ```go
   if err := config.Validate(cfg); err != nil {
       log.Fatal(err)
   }
   ```

2. **Use environment variables for sensitive data**
   ```go
   password := os.Getenv("GZCTF_PASSWORD")
   ```

### File Operations

1. **Always close files**
   ```go
   f, err := os.Open("file.txt")
   if err != nil {
       return err
   }
   defer f.Close()
   ```

2. **Check file existence before operations**
   ```go
   if !utils.FileExists(path) {
       return fmt.Errorf("file not found: %s", path)
   }
   ```

## Examples

### Complete Sync Example

```go
// Initialize
gz, err := gzcli.NewGZ(".gzctf/conf.yaml")
if err != nil {
    log.Fatal(err)
}

if err := gz.Initialize(); err != nil {
    log.Fatal(err)
}

// Sync challenges
err = gz.Sync(gzcli.SyncOptions{
    UpdateGame: true,
})
if err != nil {
    log.Fatal(err)
}
```

### Watch Example

```go
// Create watcher
w, err := watcher.NewWatcher(gz, watcher.WatcherOptions{
    Debounce: 2 * time.Second,
})
if err != nil {
    log.Fatal(err)
}

// Start watching
if err := w.Start(); err != nil {
    log.Fatal(err)
}

// Stop on interrupt
defer w.Stop()
```

### Team Creation Example

```go
// Load config
gz, err := gzcli.NewGZ(".gzctf/conf.yaml")
if err != nil {
    log.Fatal(err)
}

// Create teams from CSV
err = team.CreateTeamsFromCSV(gz, "teams.csv", true)
if err != nil {
    log.Fatal(err)
}
```

## See Also

- [Architecture Documentation](architecture.md)
- [Testing Guide](../TESTING.md)
- [Development Guide](../DEVELOPMENT.md)
- [Contributing Guidelines](../CONTRIBUTING.md)
