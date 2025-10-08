# gzcli Architecture

This document provides an overview of gzcli's architecture, design decisions, and component interactions.

## Table of Contents

- [Overview](#overview)
- [Architecture Layers](#architecture-layers)
- [Core Components](#core-components)
- [Data Flow](#data-flow)
- [File Watcher System](#file-watcher-system)
- [API Client Architecture](#api-client-architecture)
- [Configuration Management](#configuration-management)
- [Extension Points](#extension-points)
- [Design Decisions](#design-decisions)

## Overview

gzcli follows a clean architecture pattern with clear separation between:
- **CLI layer**: User interaction and command handling
- **Business logic**: Core functionality and operations
- **External services**: API calls, file system, etc.

```
┌─────────────────────────────────────────────────────────┐
│                     CLI Layer                            │
│                    (cmd/)                                │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐  ┌─────────┐     │
│  │  init   │ │  sync   │ │  watch  │  │  team   │     │
│  └─────────┘ └─────────┘ └─────────┘  └─────────┘     │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│              Business Logic Layer                        │
│               (internal/gzcli/)                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐             │
│  │ Watcher  │  │Challenge │  │   Team   │             │
│  └──────────┘  └──────────┘  └──────────┘             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐             │
│  │  Config  │  │  Event   │  │ Structure│             │
│  └──────────┘  └──────────┘  └──────────┘             │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│              External Services Layer                     │
│  ┌───────────────┐  ┌──────────────┐  ┌─────────────┐ │
│  │   GZCTF API   │  │ File System  │  │   Network   │ │
│  └───────────────┘  └──────────────┘  └─────────────┘ │
└─────────────────────────────────────────────────────────┘
```

## Architecture Layers

### 1. CLI Layer (`cmd/`)

Responsible for:
- Command-line interface and user interaction
- Flag parsing and validation
- Output formatting and user feedback
- Command execution coordination

**Key Files:**
- `root.go` - Root command and global configuration
- `init.go`, `sync.go`, `watch.go`, etc. - Individual commands
- Each command follows Cobra's command pattern

### 2. Business Logic Layer (`internal/gzcli/`)

Contains core application logic:

#### Core Types (`gzcli.go`)
- Main `GZ` struct that coordinates operations
- Configuration management
- API client initialization

#### Packages

**`gzapi/` - API client for GZ::CTF**
- Authentication and session management
- Resource operations (games, challenges, teams)
- Request/response handling

**`watcher/` - File watcher system**
- File change detection
- Debouncing logic
- Challenge redeployment

**`challenge/` - Challenge management**
- Challenge parsing and validation
- Synchronization logic
- File operations

**`team/` - Team management**
- Team creation and deletion
- CSV parsing
- Bulk operations

**`config/` - Configuration handling**
- YAML parsing
- Validation
- Environment-specific configs

**`event/` - Event system**
- Webhooks (Discord, etc.)
- Event notifications
- Logging

### 3. External Services Layer

Abstractions for external dependencies:
- HTTP client for API calls
- File system operations
- Network operations

## Core Components

### GZ Core Type

The main `GZ` struct coordinates all operations:

```go
type GZ struct {
    Conf   *config.Config
    Api    *gzapi.GZAPI
    Event  *event.Event
    // ... other fields
}
```

**Responsibilities:**
- Initialize and manage API client
- Load and validate configuration
- Coordinate between components
- Handle cross-cutting concerns

### API Client (`gzapi/`)

HTTP client for GZ::CTF API:

```
┌──────────────────────┐
│      GZAPI           │
│  ┌────────────────┐  │
│  │ Authentication │  │
│  └────────────────┘  │
│  ┌────────────────┐  │
│  │ Games          │  │
│  └────────────────┘  │
│  ┌────────────────┐  │
│  │ Challenges     │  │
│  └────────────────┘  │
│  ┌────────────────┐  │
│  │ Teams/Users    │  │
│  └────────────────┘  │
└──────────────────────┘
```

**Features:**
- Session management
- Rate limiting
- Retry logic
- Error handling

### Challenge System

Challenge lifecycle:

```
  Read Files
      │
      ▼
  Parse YAML
      │
      ▼
  Validate
      │
      ▼
   Upload
      │
      ▼
   Deploy
```

### Team Management

Team operations flow:

```
 CSV File
     │
     ▼
 Parse CSV
     │
     ▼
 Validate
     │
     ▼
Create Users
     │
     ▼
Create Teams
     │
     ▼
Assign Users
```

## Data Flow

### Sync Command Flow

```
 User runs
 'gzcli sync'
      │
      ▼
 Load Config
      │
      ▼
 Login to API
      │
      ▼
 Scan Challenges
      │
      ▼
┌─────┴─────┐
│           │
▼           ▼
Create      Update
New         Existing
│           │
└─────┬─────┘
      │
      ▼
 Upload Files
      │
      ▼
   Deploy
      │
      ▼
 Send Events
```

### Watch Command Flow

```
 Start Watcher
      │
      ▼
 Watch Files ◄────┐
      │           │
      ▼           │
 File Changed     │
      │           │
      ▼           │
  Debounce        │
      │           │
      ▼           │
 Identify Chall   │
      │           │
      ▼           │
 Sync Challenge   │
      │           │
      ▼           │
   Success? ──────┘
```

## File Watcher System

The watcher is a critical component for development workflow:

### Architecture

```
┌────────────────────────────────────────┐
│           Watcher Core                 │
│                                        │
│  ┌──────────────────────────────────┐ │
│  │     File System Monitor          │ │
│  │     (fsnotify)                   │ │
│  └────────────┬─────────────────────┘ │
│               │                        │
│  ┌────────────▼─────────────────────┐ │
│  │     Event Queue                  │ │
│  │     (Channel-based)              │ │
│  └────────────┬─────────────────────┘ │
│               │                        │
│  ┌────────────▼─────────────────────┐ │
│  │     Debouncer                    │ │
│  │     (Time-based)                 │ │
│  └────────────┬─────────────────────┘ │
│               │                        │
│  ┌────────────▼─────────────────────┐ │
│  │     Handler                      │ │
│  │     (Sync Challenge)             │ │
│  └──────────────────────────────────┘ │
└────────────────────────────────────────┘
```

### Key Features

1. **Recursive Watching:** Monitors all subdirectories
2. **Debouncing:** Groups rapid changes to avoid redundant syncs
3. **Ignore Patterns:** Filters out unwanted file changes
4. **Error Recovery:** Continues watching after sync failures
5. **Daemon Mode:** Runs in background with logging

### Implementation Details

**Debouncing Algorithm:**
```
File Change Event
      │
      ▼
  Add to Queue
      │
      ▼
 Start Timer (configurable, default 2s)
      │
      ▼
 More Events? ──Yes──► Reset Timer
      │
      No
      ▼
Process Event
```

**Challenge Detection:**
```
File Changed: /challenges/web/xss/src/index.html
      │
      ▼
 Walk up directory tree
      │
      ▼
 Find challenge.yaml
      │
      ▼
 Sync /challenges/web/xss
```

## API Client Architecture

### Request Flow

```
 User Code
      │
      ▼
  API Method
      │
      ▼
 Build Request
      │
      ▼
 Add Auth Headers
      │
      ▼
  Send Request ◄──── Retry Logic
      │
      ▼
 Parse Response
      │
      ▼
 Handle Errors
      │
      ▼
 Return Result
```

### Session Management

```
┌─────────────────────┐
│  Session Manager    │
│                     │
│  ┌───────────────┐  │
│  │  Login        │  │
│  └───────────────┘  │
│  ┌───────────────┐  │
│  │  Token Store  │  │
│  └───────────────┘  │
│  ┌───────────────┐  │
│  │  Refresh      │  │
│  └───────────────┘  │
└─────────────────────┘
```

**Features:**
- Automatic login on first request
- Token storage and reuse
- Automatic refresh on expiry
- Logout on exit

## Configuration Management

### Configuration Hierarchy

```
1. Default Values
        │
        ▼
2. Config File (.gzctf/conf.yaml)
        │
        ▼
3. Environment Variables
        │
        ▼
4. Command-line Flags
        │
        ▼
   Final Config
```

### Configuration Structure

```yaml
url: https://ctf.example.com
public_entry: https://public.example.com

creds:
  username: admin
  password: password

event:
  discord_webhook: https://discord.com/...

watcher:
  debounce: 2s
  ignore_patterns:
    - "*.tmp"
    - ".git/"
```

## Extension Points

### 1. Custom Event Handlers

Implement custom event handlers:

```go
type CustomHandler struct {
    // fields
}

func (h *CustomHandler) OnSync(challenge *Challenge) {
    // Custom logic
}
```

### 2. Custom Validators

Add custom validation:

```go
func RegisterValidator(name string, fn ValidatorFunc) {
    validators[name] = fn
}
```

### 3. Custom Scripts

Execute custom scripts via `scripts` section in `challenge.yaml`:

```yaml
scripts:
  build: ./build.sh
  deploy: ./deploy.sh
  test: ./test.sh
```

## Design Decisions

### Why Cobra for CLI?

- **Pros:** Rich feature set, widely used, good documentation
- **Cons:** Some overhead for simple commands
- **Decision:** Benefits outweigh complexity for multi-command CLI

### Why fsnotify for File Watching?

- **Pros:** Cross-platform, efficient, well-tested
- **Cons:** Platform-specific quirks
- **Decision:** Best available option for Go

### Why No External DB?

- **Decision:** File-based configuration is simpler
- **Benefits:** Easy deployment, no dependencies
- **Trade-off:** Not suitable for multi-user scenarios

### Channel-Based Concurrency

- **Pattern:** Use channels for event communication
- **Benefit:** Idiomatic Go, easier to reason about
- **Alternative:** Callbacks/mutexes (more complex)

### Daemon vs Foreground Watch

- **Support Both:** Different use cases
- **Daemon:** Production use, CI/CD
- **Foreground:** Development, debugging

## Error Handling Strategy

### Principle

1. **Fail Fast**: Validate early, return errors immediately
2. **Contextual Errors**: Wrap errors with context
3. **User-Friendly Messages**: Translate technical errors
4. **Logging**: Log at appropriate levels

### Example

```go
if err := ValidateChallenge(ch); err != nil {
    return fmt.Errorf("validating challenge %s: %w", ch.Name, err)
}
```

## Performance Considerations

### File Operations

- Use buffered I/O for large files
- Minimize disk reads/writes
- Cache parsed configurations

### API Calls

- Batch operations when possible
- Reuse HTTP connections
- Implement rate limiting

### Memory Usage

- Stream large files instead of loading entirely
- Clean up resources promptly
- Use appropriate data structures

## Testing Strategy

### Unit Tests

- Test individual components in isolation
- Mock external dependencies
- Focus on business logic

### Integration Tests

- Test component interactions
- Use test doubles for external services
- Verify end-to-end flows

### Worst-Case Tests

- Test edge cases and error conditions
- Simulate failures (network, disk, etc.)
- Verify recovery behavior

See [Development & Testing Guide](development.md#testing) for detailed testing guide.

## Security Considerations

### Credentials

- Never log credentials
- Store securely (file permissions)
- Support environment variables
- Clear from memory after use

### File Operations

- Validate paths to prevent traversal
- Check file permissions
- Sanitize user input

### API Communication

- Always use HTTPS
- Validate certificates
- Implement timeouts
- Rate limit requests

## Future Enhancements

### Potential Improvements

1. **Plugin System:** Support external plugins
2. **Database Backend:** Optional DB for large deployments
3. **Web UI:** Optional web interface for management
4. **Multi-tenancy:** Support multiple CTF instances
5. **Advanced Caching:** Reduce API calls
6. **Parallel Operations:** Speed up bulk operations

### Migration Paths

- Maintain backward compatibility
- Provide migration tools
- Document breaking changes
- Support gradual upgrades

## Resources

- [Development & Testing Guide](development.md): Development setup, workflow, and testing guide
- [Contributing Guidelines](../CONTRIBUTING.md): Contribution guidelines
- [API Reference](api-reference.md): Internal API documentation
