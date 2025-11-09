# Development Guide

This guide provides comprehensive information for developers who want to contribute to gzcli or understand its architecture, along with complete testing guidelines.

## Table of Contents

- [Development](#development)
  - [Architecture Overview](#architecture-overview)
  - [Project Structure](#project-structure)
  - [Development Setup](#development-setup)
  - [Building and Running](#building-and-running)
    - [Challenge Upload Server](#challenge-upload-server)
  - [Adding New Features](#adding-new-features)
  - [Debugging](#debugging)
  - [Performance Profiling](#performance-profiling)
  - [Test Environment](#test-environment)
  - [Available Make Targets](#available-make-targets)
- [Testing](#testing)
  - [Testing Philosophy](#testing-philosophy)
  - [Running Tests](#running-tests)
  - [Writing Tests](#writing-tests)
  - [Test Organization](#test-organization)
  - [Mocking Strategies](#mocking-strategies)
  - [Integration Tests](#integration-tests)
  - [Test Coverage](#test-coverage)
  - [Best Practices](#best-practices)
  - [Benchmarking](#benchmarking)
  - [Continuous Integration](#continuous-integration)
  - [Troubleshooting](#troubleshooting)
- [Additional Resources](#additional-resources)

## Development

### Architecture Overview

gzcli is built with a clean architecture pattern:

```
┌─────────────────────────────────────┐
│     CLI Layer (cmd/)                │
│   - Command definitions             │
│   - Flag parsing                    │
│   - User interaction                │
└────────────┬────────────────────────┘
             │
┌────────────▼────────────────────────┐
│     Business Logic (internal/)      │
│   - Core functionality              │
│   - API interactions                │
│   - File operations                 │
│   - Watcher system                  │
└────────────┬────────────────────────┘
             │
┌────────────▼────────────────────────┐
│     External Services               │
│   - GZ::CTF API                     │
│   - File system                     │
│   - Network operations              │
└─────────────────────────────────────┘
```

#### Key Components

1. **CLI Layer** (`cmd/`)
   - Uses [Cobra](https://github.com/spf13/cobra) for command-line interface
   - Each command in a separate file
   - Handles user input and validation

2. **Business Logic** (`internal/`)
   - `gzcli/` - Core CTF management logic
     - `gzapi/` - API client
     - `watcher/` - File watcher system
     - `challenge/` - Challenge management
     - `team/` - Team management
     - `config/` - Configuration handling
     - `event/` - Event handling
   - `log/` - Logging utilities
   - `utils/` - Helper functions
   - `template/` - Template system for challenge generation

3. **API Client** (`internal/gzcli/gzapi/`)
   - HTTP client for GZ::CTF API
   - Authentication and session management
   - Resource management (games, challenges, teams)

For detailed architecture documentation, see [architecture.md](architecture.md).

### Project Structure

```
gzcli/
├── cmd/                     # Command implementations
│   ├── root.go             # Root command
│   ├── init.go             # Init command
│   ├── sync.go             # Sync command
│   ├── watch.go            # Watch parent command
│   ├── watch_start.go      # Start watcher
│   ├── watch_stop.go       # Stop watcher
│   ├── watch_status.go     # Watcher status
│   ├── watch_logs.go       # View watcher logs
│   ├── team.go             # Team parent command
│   ├── team_create.go      # Create teams
│   ├── team_delete.go      # Delete teams
│   ├── script.go           # Script command
│   ├── scoreboard.go       # Scoreboard command
│   └── structure.go        # Structure command
│
├── internal/               # Private application code
│   ├── gzcli/             # Core logic
│   │   ├── gzapi/         # API client
│   │   ├── watcher/       # File watcher system
│   │   │   ├── core/      # Core watcher logic
│   │   │   └── ...
│   │   ├── challenge/     # Challenge management
│   │   ├── team/          # Team management
│   │   ├── config/        # Configuration
│   │   ├── event/         # Event handling
│   │   ├── structure/     # Structure management
│   │   ├── script/        # Script execution
│   │   ├── gzcli.go       # Main GZ type
│   │   └── ...
│   ├── log/               # Logging
│   ├── utils/             # Utilities
│   └── template/          # Templates
│
├── scripts/               # Development scripts
│   ├── setup.sh           # Environment setup
│   ├── test.sh            # Test runner
│   ├── lint.sh            # Linter runner
│   └── install-hooks.sh   # Git hooks installer
│
├── docs/                  # Documentation
│   ├── development.md     # Development and testing guide
│   ├── architecture.md    # System architecture
│   ├── api-reference.md   # Internal API documentation
│   ├── BINARY_OPTIMIZATION.md  # Binary size optimizations
│   ├── COMPLETION.md      # Shell completion guide
│   ├── MULTI_EVENT.md     # Multi-event management
│   ├── PERFORMANCE.md     # Performance guide
│   └── VERSIONING.md      # Automated versioning
│
├── Makefile               # Build automation
├── .golangci.yml          # Linter configuration
├── .goreleaser.yml        # Release configuration
├── go.mod                 # Go module
├── main.go                # Entry point
└── README.md              # Documentation
```

### Development Setup

#### Quick Start

```bash
# Clone the repository
git clone https://github.com/dimasma0305/gzcli.git
cd gzcli

# Complete setup with verification
make setup-complete

# Or run setup script
./scripts/setup.sh

# Verify environment
make doctor

# Or manually:
make deps    # Download dependencies
make tools   # Install dev tools
make build   # Build binary
```

#### IDE Setup

##### VS Code

The project includes pre-configured VS Code settings in `.vscode/`:
- `settings.json`: Go-specific settings and formatting
- `extensions.json`: Recommended extensions (auto-prompted)
- `launch.json`: Debug configurations
- `tasks.json`: Quick build/test tasks

Simply open the project in VS Code and install the recommended extensions when prompted.

##### GitHub Codespaces

Instantly develop in the cloud with everything pre-configured:
- Click "Code" → "Open with Codespaces" on GitHub
- Or use the badge in README.md

The dev container configuration (`.devcontainer/devcontainer.json`) ensures a consistent environment.

### Building and Running

#### Build Commands

```bash
# Build binary
make build

# Build with version info
make build VERSION=v1.0.0

# Install to $GOPATH/bin
make install

# Run in development mode (with hot reload using air)
make dev

# Clean build artifacts
make clean

# Build for multiple platforms (requires goreleaser)
make release
```

#### Optimized Builds

The project uses build optimizations to reduce binary size:

- **Optimized build:** ~18 MB (with `-trimpath`, `-s`, `-w` flags)

For more details on binary optimization, see [BINARY_OPTIMIZATION.md](BINARY_OPTIMIZATION.md).

The build flags strip debug information and file paths, reducing binary size by ~33% from the baseline build without affecting runtime performance.

#### Running

```bash
# Run directly
./gzcli --help

# Run with go run
go run . --help

# Run specific command
./gzcli init --help
```

#### Challenge Upload Server

Use the dedicated upload server when you need a simple web UI for packaging and ingesting challenges.

```bash
# Start the upload server on localhost:8090
gzcli upload-server

# Custom host/port
gzcli upload-server --host 0.0.0.0 --port 4000
```

- The server reads events from the current workspace (`events/<event>/`) and requires those events to exist locally.
- The home page lists built-in templates sourced from the project samples (e.g. Static Container, Static Attachment variants); download them at `/templates/<slug>.zip`.
- Uploads accept ZIP archives only; the server locates `challenge.yml`, validates it with the existing challenge checks, and ensures a `writeup/` directory is present.
- The selected event and category determine the destination (`events/<event>/<category>/<challenge-name>/`). If a challenge with the same name already exists, its contents are replaced.
- Authentication is intentionally not enforced—run the server only on trusted networks or wrap it with your own access controls if required.

### Adding New Features

#### Adding a New Command

1. Create command file `cmd/newcommand.go`:

```go
package cmd

import (
    "github.com/dimasma0305/gzcli/internal/gzcli"
    "github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
    Use:   "new",
    Short: "Short description",
    Long:  `Long description`,
    Run: func(cmd *cobra.Command, args []string) {
        // Implementation
    },
}

func init() {
    rootCmd.AddCommand(newCmd)

    // Add flags
    newCmd.Flags().StringP("flag", "f", "", "Flag description")
}
```

2. Add tests `cmd/newcommand_test.go`
3. Update documentation

#### Adding a New API Endpoint

1. Add method to `internal/gzcli/gzapi/`:

```go
func (api *GZAPI) NewEndpoint() error {
    url := api.Url + "/api/endpoint"
    _, err := api.Client.R().Get(url)
    return err
}
```

2. Add tests
3. Use in command

### Debugging

#### Enable Debug Mode

```bash
# Via flag
gzcli --debug init

# Via environment variable
export GZCLI_DEBUG=1
gzcli init
```

#### Using Delve Debugger

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug
dlv debug . -- init

# Set breakpoint
(dlv) break main.main
(dlv) continue
```

#### Logging

Add debug logs in code:

```go
log.Debug("Processing file: %s", filename)
log.DebugH2("  Size: %d bytes", size)
```

### Performance Profiling

#### CPU Profiling

```bash
# Build with profiling
go build -o gzcli .

# Run with CPU profile
gzcli --cpuprofile=cpu.prof sync

# Analyze
go tool pprof cpu.prof
```

#### Memory Profiling

```bash
# Run with memory profile
gzcli --memprofile=mem.prof sync

# Analyze
go tool pprof mem.prof
```

### Best Practices

1. **Error Handling**
   - Always handle errors
   - Provide context in error messages
   - Use `fmt.Errorf` with `%w` for wrapping errors

2. **Logging**
   - Use appropriate log levels
   - Include context in log messages
   - Don't log sensitive information

3. **Testing**
   - Aim for >80% code coverage
   - Test edge cases
   - Use table-driven tests

4. **Code Organization**
   - Keep functions small and focused
   - Use meaningful names
   - Add comments for exported functions

5. **Performance**
   - Avoid premature optimization
   - Profile before optimizing
   - Use `sync.Pool` for frequently allocated objects

### Test Environment

gzcli provides make targets for managing a test environment:

```bash
# Initialize test environment
make test-env-init

# Edit .test/.gzctf/conf.yaml with your settings

# Note: You'll need a running GZ::CTF instance for integration testing
# Configure the URL in .test/.gzctf/conf.yaml
```

### Available Make Targets

Run `make help` to see all available targets. Key targets include:

**Building:**
- `make build` - Build the binary
- `make install` - Install to $GOPATH/bin
- `make clean` - Clean build artifacts
- `make release` - Build for multiple platforms

**Development:**
- `make dev` - Run with hot reload
- `make fmt` - Format code
- `make lint` - Run linters
- `make vet` - Run go vet
- `make check` - Run all checks
- `make ci` - Run CI checks

**Testing:**
- `make test` - Run all tests
- `make test-unit` - Run unit tests
- `make test-integration` - Run integration tests
- `make test-coverage` - Generate coverage report
- `make test-race` - Run with race detector
- `make test-watcher` - Run watcher tests
- `make test-challenge` - Run challenge tests
- `make test-api` - Run API tests
- `make test-watch` - Watch and re-run tests
- `make bench` - Run benchmarks

**Dependencies:**
- `make deps` - Download dependencies
- `make deps-update` - Update dependencies
- `make tools` - Install dev tools

**Test Environment:**
- `make test-env-init` - Initialize test environment
- `make test-env-clean` - Clean test data

### Troubleshooting

#### Common Issues

**Build Fails:**
```bash
# Clear cache
go clean -cache -modcache
go mod download
go build
```

**Tests Fail:**
```bash
# Run verbose
go test -v ./...

# Run specific test
go test -v ./path -run TestName
```

**Import Issues:**
```bash
# Tidy modules
go mod tidy
```

**Development Tools Missing:**
```bash
# Install all development tools
make tools
```

## Testing

### Testing Philosophy

gzcli uses Go's built-in testing framework along with table-driven tests for comprehensive test coverage. We aim for:

- **>80% code coverage** for core packages
- **Fast unit tests:** <5 seconds for the full suite
- **Comprehensive integration tests** for critical paths
- **Clear test names** that describe what is being tested

#### Test Pyramid

We follow the testing pyramid approach:

```
        /\
       /  \      E2E Tests (Few)
      /----\
     /      \    Integration Tests (Some)
    /--------\
   /          \  Unit Tests (Many)
  /____________\
```

- **Unit Tests:** Test individual functions and methods in isolation
- **Integration Tests:** Test interactions between components
- **E2E Tests:** Test complete workflows (manual or CI-only)

#### Test Naming

Tests should clearly describe **what** is being tested and **what** the expected behavior is:

```go
func TestWatcher_HandleFileChange_RedeploysChallenge(t *testing.T)
func TestGZAPI_Login_WithInvalidCredentials_ReturnsError(t *testing.T)
func TestConfig_Load_WithMissingFile_ReturnsError(t *testing.T)
```

**Format:** `Test<Component>_<Method>_<Scenario>_<ExpectedResult>`

### Running Tests

#### All Tests

```bash
# Run all tests
make test

# Run with verbose output
go test -v ./...

# Run with race detector
make test-race
```

#### Unit Tests Only

```bash
# Short mode (skips slow integration tests)
make test-unit

# Or directly
go test -short ./...
```

#### Specific Tests

```bash
# Run tests in a specific package
go test -v ./internal/gzcli/watcher/...

# Run a specific test
go test -v ./internal/gzcli/watcher -run TestWatcher_Start

# Run tests matching a pattern
go test -v ./... -run "TestWatcher.*"
```

#### Component-Specific Tests

```bash
# Watcher tests
make test-watcher

# Challenge tests
make test-challenge

# API tests
make test-api
```

#### Test Coverage

```bash
# Generate coverage report
make test-coverage

# View in browser
make coverage-browse

# Check coverage for specific package
go test -cover ./internal/gzcli/watcher/...
```

#### Watch Mode

```bash
# Auto-run tests on file changes
make test-watch
```

### Writing Tests

#### Table-Driven Tests

Use table-driven tests for testing multiple scenarios:

```go
func TestParseChallenge(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *Challenge
        wantErr bool
    }{
        {
            name: "valid challenge",
            input: `name: Test
category: Web`,
            want: &Challenge{
                Name:     "Test",
                Category: "Web",
            },
            wantErr: false,
        },
        {
            name:    "invalid YAML",
            input:   "invalid: [",
            want:    nil,
            wantErr: true,
        },
        {
            name:    "empty input",
            input:   "",
            want:    nil,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseChallenge(tt.input)

            if (err != nil) != tt.wantErr {
                t.Errorf("ParseChallenge() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("ParseChallenge() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

#### Testing with Fixtures

Use test fixtures for complex test data:

```go
func TestLoadConfig(t *testing.T) {
    // Load fixture
    data, err := os.ReadFile("testdata/valid_config.yaml")
    if err != nil {
        t.Fatalf("Failed to load fixture: %v", err)
    }

    cfg, err := ParseConfig(data)
    if err != nil {
        t.Errorf("ParseConfig() failed: %v", err)
    }

    // Assertions
    if cfg.URL != "https://example.com" {
        t.Errorf("URL = %v, want https://example.com", cfg.URL)
    }
}
```

#### Using Subtests

Subtests provide better organization and allow running specific test cases:

```go
func TestAPI(t *testing.T) {
    t.Run("Login", func(t *testing.T) {
        t.Run("ValidCredentials", func(t *testing.T) {
            // Test valid login
        })

        t.Run("InvalidCredentials", func(t *testing.T) {
            // Test invalid login
        })
    })

    t.Run("GetGames", func(t *testing.T) {
        // Test getting games
    })
}
```

#### Error Testing

Always test both success and error cases:

```go
func TestDivide(t *testing.T) {
    t.Run("ValidDivision", func(t *testing.T) {
        result, err := Divide(10, 2)
        if err != nil {
            t.Errorf("unexpected error: %v", err)
        }
        if result != 5 {
            t.Errorf("got %v, want 5", result)
        }
    })

    t.Run("DivisionByZero", func(t *testing.T) {
        _, err := Divide(10, 0)
        if err == nil {
            t.Error("expected error, got nil")
        }
    })
}
```

### Test Organization

#### File Structure

```
package/
├── handler.go
├── handler_test.go                  # Unit tests
├── handler_integration_test.go      # Integration tests
├── testdata/                        # Test fixtures
│   ├── valid_input.json
│   └── invalid_input.json
└── testutil/                        # Test helpers
    └── mocks.go
```

#### Test Tags

Use build tags to separate test types:

```go
//go:build integration
// +build integration

package gzapi_test
```

Run with: `go test -tags=integration ./...`

### Mocking Strategies

#### Interface Mocking

Define interfaces for dependencies:

```go
// Interface for HTTP client
type HTTPClient interface {
    Do(req *http.Request) (*http.Response, error)
}

// Mock implementation
type MockHTTPClient struct {
    DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
    return m.DoFunc(req)
}

// Test using mock
func TestAPICall(t *testing.T) {
    mockClient := &MockHTTPClient{
        DoFunc: func(req *http.Request) (*http.Response, error) {
            return &http.Response{
                StatusCode: 200,
                Body:       io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
            }, nil
        },
    }

    api := &API{client: mockClient}
    // Test with mock client
}
```

#### Using testutil Package

See [testutil README](../internal/gzcli/testutil/README.md) for available test utilities.

### Integration Tests

Integration tests verify interactions between components:

```go
func TestSync_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // Setup test environment
    tmpDir := t.TempDir()

    // Create test files
    // ...

    // Run sync
    err := Sync(tmpDir)
    if err != nil {
        t.Fatalf("Sync failed: %v", err)
    }

    // Verify results
    // ...
}
```

Skip integration tests with: `go test -short ./...`

### Test Coverage

#### Measuring Coverage

```bash
# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out

# Show coverage by function
go tool cover -func=coverage.out
```

#### Coverage Goals

- **Critical packages** (watcher, challenge, API): >85%
- **Command handlers:** >70%
- **Utilities:** >80%
- **Overall project:** >80%

#### Improving Coverage

1. Identify uncovered code:
   ```bash
   go test -coverprofile=coverage.out ./...
   go tool cover -html=coverage.out
   ```

2. Add tests for uncovered paths
3. Focus on error paths and edge cases
4. Use table-driven tests to cover multiple scenarios

### Best Practices

#### DO

✅ Write tests before fixing bugs (TDD for bug fixes)
✅ Test public APIs and exported functions
✅ Use table-driven tests for multiple scenarios
✅ Test error conditions and edge cases
✅ Keep tests fast and focused
✅ Use meaningful test names
✅ Clean up resources in tests (use `t.Cleanup()`)
✅ Use subtests for better organization
✅ Mock external dependencies
✅ Test concurrent code with race detector

#### DON'T

❌ Test implementation details
❌ Write flaky tests (time-dependent, order-dependent)
❌ Ignore test failures
❌ Skip testing error paths
❌ Write tests that depend on external services (without mocks)
❌ Commit code with failing tests
❌ Use sleep() to wait for async operations
❌ Hard-code paths or configurations

#### Example: Clean Test Structure

```go
func TestFeature(t *testing.T) {
    // Setup
    tmpDir := t.TempDir()
    t.Cleanup(func() {
        // Cleanup if needed (tmpDir auto-cleaned)
    })

    // Execute
    result, err := Feature(tmpDir)

    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

### Benchmarking

#### Writing Benchmarks

```go
func BenchmarkParseChallenge(b *testing.B) {
    input := loadTestData()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        ParseChallenge(input)
    }
}
```

#### Running Benchmarks

```bash
# Run benchmarks
go test -bench=. ./...

# With memory stats
go test -bench=. -benchmem ./...

# Compare benchmarks
make bench > old.txt
# Make changes
make bench > new.txt
benchcmp old.txt new.txt
```

### Continuous Integration

Tests run automatically in CI on:
- Every push to `main` or `develop`
- Every pull request
- Multiple Go versions (1.23+)
- Multiple operating systems (Linux, macOS, Windows)

See `.github/workflows/ci.yml` for CI configuration.

### Troubleshooting

#### Tests Failing Locally

```bash
# Clean build cache
go clean -testcache

# Update dependencies
go mod tidy
go mod download

# Run with verbose output
go test -v ./...
```

#### Flaky Tests

If tests are flaky:
1. Remove timing dependencies
2. Use proper synchronization (channels, WaitGroups)
3. Increase timeouts if necessary
4. Ensure proper test isolation

#### Race Conditions

```bash
# Run with race detector
go test -race ./...

# Fix race conditions by using:
# - Mutexes
# - Proper channel usage
# - Avoiding shared state
```

## Additional Resources

### Documentation
- [Architecture](architecture.md) - System architecture and design decisions
- [API Reference](api-reference.md) - Internal API documentation
- [Contributing Guidelines](../CONTRIBUTING.md) - How to contribute
- [Binary Optimization](BINARY_OPTIMIZATION.md) - Binary size reduction techniques
- [Multi-Event Management](MULTI_EVENT.md) - Managing multiple CTF events
- [Performance Guide](PERFORMANCE.md) - Performance optimization details
- [Versioning Guide](VERSIONING.md) - Automated semantic versioning

### External Resources
- [Go Documentation](https://golang.org/doc/)
- [Cobra Documentation](https://github.com/spf13/cobra)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Testing Package](https://pkg.go.dev/testing)
- [Table-Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Air - Live Reload](https://github.com/air-verse/air)
- [GoReleaser](https://goreleaser.com/)

### Getting Help
- Open an [issue](https://github.com/dimasma0305/gzcli/issues) on GitHub
- Check existing [discussions](https://github.com/dimasma0305/gzcli/discussions)
- Ask in [GitHub Discussions](https://github.com/dimasma0305/gzcli/discussions)
