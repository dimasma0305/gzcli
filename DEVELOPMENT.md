# Development Guide

This guide provides information for developers who want to contribute to gzcli or understand its architecture.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Project Structure](#project-structure)
- [Development Setup](#development-setup)
- [Building and Running](#building-and-running)
- [Testing](#testing)
- [Adding New Features](#adding-new-features)
- [Debugging](#debugging)
- [Performance Profiling](#performance-profiling)
- [Test Environment](#test-environment)
- [Available Make Targets](#available-make-targets)

## Architecture Overview

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

### Key Components

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

## Project Structure

```
gzcli/
├── cmd/                    # Command implementations
│   ├── root.go            # Root command
│   ├── init.go            # Init command
│   ├── sync.go            # Sync command
│   ├── watch.go           # Watch parent command
│   ├── watch_start.go     # Start watcher
│   ├── watch_stop.go      # Stop watcher
│   ├── watch_status.go    # Watcher status
│   ├── watch_logs.go      # View watcher logs
│   ├── team.go            # Team parent command
│   ├── team_create.go     # Create teams
│   ├── team_delete.go     # Delete teams
│   ├── script.go          # Script command
│   ├── scoreboard.go      # Scoreboard command
│   └── structure.go       # Structure command
│
├── internal/              # Private application code
│   ├── gzcli/            # Core logic
│   │   ├── gzapi/        # API client
│   │   ├── watcher/      # File watcher system
│   │   │   ├── core/     # Core watcher logic
│   │   │   └── ...
│   │   ├── challenge/    # Challenge management
│   │   ├── team/         # Team management
│   │   ├── config/       # Configuration
│   │   ├── event/        # Event handling
│   │   ├── structure/    # Structure management
│   │   ├── script/       # Script execution
│   │   ├── gzcli.go      # Main GZ type
│   │   └── ...
│   ├── log/              # Logging
│   ├── utils/            # Utilities
│   └── template/         # Templates
│
├── scripts/              # Development scripts
│   ├── setup.sh          # Environment setup
│   ├── test.sh           # Test runner
│   ├── lint.sh           # Linter runner
│   └── install-hooks.sh  # Git hooks installer
│
├── Makefile              # Build automation
├── .golangci.yml         # Linter configuration
├── .goreleaser.yml       # Release configuration
├── go.mod                # Go module
├── main.go               # Entry point
└── README.md             # Documentation
```

## Development Setup

### Quick Start

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

### IDE Setup

#### VS Code

The project includes pre-configured VS Code settings in `.vscode/`:
- `settings.json`: Go-specific settings and formatting
- `extensions.json`: Recommended extensions (auto-prompted)
- `launch.json`: Debug configurations
- `tasks.json`: Quick build/test tasks

Simply open the project in VS Code and install recommended extensions when prompted.

#### GitHub Codespaces

Instantly develop in the cloud with everything pre-configured:
- Click "Code" → "Open with Codespaces" on GitHub
- Or use the badge in README.md

The dev container configuration (`.devcontainer/devcontainer.json`) ensures a consistent environment.

## Building and Running

### Build Commands

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

### Running

```bash
# Run directly
./gzcli --help

# Run with go run
go run . --help

# Run specific command
./gzcli init --help
```

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run unit tests only (short mode)
make test-unit

# Run integration tests
make test-integration

# Run tests with coverage
make test-coverage

# Run tests with race detector
make test-race

# Run specific component tests
make test-watcher       # Watcher tests
make test-challenge     # Challenge tests
make test-api           # API client tests

# Run tests with auto-reload (requires air)
make test-watch

# Run benchmarks
make bench

# Run specific tests
go test -v ./internal/gzcli -run TestSpecificFunction
```

### Writing Tests

Follow table-driven test pattern:

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:    "valid input",
            input:   "test",
            want:    "result",
            wantErr: false,
        },
        // More cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Function(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Adding New Features

### Adding a New Command

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

### Adding a New API Endpoint

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

## Debugging

### Enable Debug Mode

```bash
# Via flag
gzcli --debug init

# Via environment variable
export GZCLI_DEBUG=1
gzcli init
```

### Using Delve Debugger

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug
dlv debug . -- init

# Set breakpoint
(dlv) break main.main
(dlv) continue
```

### Logging

Add debug logs in code:

```go
log.Debug("Processing file: %s", filename)
log.DebugH2("  Size: %d bytes", size)
```

## Performance Profiling

### CPU Profiling

```bash
# Build with profiling
go build -o gzcli .

# Run with CPU profile
gzcli --cpuprofile=cpu.prof sync

# Analyze
go tool pprof cpu.prof
```

### Memory Profiling

```bash
# Run with memory profile
gzcli --memprofile=mem.prof sync

# Analyze
go tool pprof mem.prof
```

### Benchmarking

```go
func BenchmarkFunction(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Function()
    }
}
```

Run benchmarks:
```bash
go test -bench=. -benchmem ./...
```

## Best Practices

1. **Error Handling**
   - Always handle errors
   - Provide context in error messages
   - Use `fmt.Errorf` with `%w` for wrapping

2. **Logging**
   - Use appropriate log levels
   - Include context in log messages
   - Don't log sensitive information

3. **Testing**
   - Aim for >70% coverage
   - Test edge cases
   - Use table-driven tests

4. **Code Organization**
   - Keep functions small and focused
   - Use meaningful names
   - Add comments for exported functions

5. **Performance**
   - Avoid premature optimization
   - Profile before optimizing
   - Use sync.Pool for frequently allocated objects

## Test Environment

gzcli provides make targets for managing a test environment:

```bash
# Initialize test environment
make test-env-init

# Edit .test/.gzctf/conf.yaml with your settings

# Note: You'll need a running GZCTF instance for integration testing
# Configure the URL in .test/.gzctf/conf.yaml
```

## Troubleshooting

### Common Issues

**Build Fails**
```bash
# Clear cache
go clean -cache -modcache
go mod download
go build
```

**Tests Fail**
```bash
# Run verbose
go test -v ./...

# Run specific test
go test -v ./path -run TestName
```

**Import Issues**
```bash
# Tidy modules
go mod tidy
```

**Development Tools Missing**
```bash
# Install all development tools
make tools
```

## Available Make Targets

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

## Additional Documentation

- **[TESTING.md](TESTING.md)**: Comprehensive testing guide
- **[Architecture](docs/architecture.md)**: System architecture and design decisions
- **[API Reference](docs/api-reference.md)**: Internal API documentation
- **[CONTRIBUTING.md](CONTRIBUTING.md)**: Contribution guidelines

## Resources

- [Go Documentation](https://golang.org/doc/)
- [Cobra Documentation](https://github.com/spf13/cobra)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Air - Live Reload](https://github.com/air-verse/air)
- [GoReleaser](https://goreleaser.com/)

## Getting Help

- Open an [issue](https://github.com/dimasma0305/gzcli/issues) on GitHub
- Check existing [discussions](https://github.com/dimasma0305/gzcli/discussions)
- Read the [CONTRIBUTING.md](CONTRIBUTING.md) guide
