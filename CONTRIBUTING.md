# Contributing to gzcli

Thank you for your interest in contributing to gzcli! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Development Workflow](#development-workflow)
- [Code Style](#code-style)
- [Testing](#testing)
- [Test Environment](#test-environment)
- [Submitting Changes](#submitting-changes)
- [Commit Message Guidelines](#commit-message-guidelines)

## Getting Started

### Prerequisites

- Go 1.23 or later
- Git
- Make (optional but recommended)

### Environment Check

Before starting, verify your development environment:

```bash
make doctor
```

This will check all prerequisites and provide guidance for any missing tools.

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/gzcli.git
   cd gzcli
   ```
3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/dimasma0305/gzcli.git
   ```

## Development Setup

### Quick Setup

```bash
# Complete automated setup
make setup-complete

# Or step by step:

# Install dependencies
make deps

# Install development tools (golangci-lint, goimports, goreleaser, air)
make tools

# Install git hooks
make hooks

# Build the project
make build

# Verify everything works
make doctor
```

### Manual Setup

```bash
# Download dependencies
go mod download

# Install linters and tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/goreleaser/goreleaser@latest
go install github.com/air-verse/air@latest

# Build
go build -o gzcli .
```

## Development Workflow

### 1. Create a Branch

```bash
# Update your main branch
git checkout main
git pull upstream main

# Create a feature branch
git checkout -b feature/your-feature-name
```

### 2. Make Changes

- Write your code
- Add tests for new functionality
- Ensure existing tests pass
- Update documentation as needed

### 3. Run Checks Locally

```bash
# Format code
make fmt

# Run linters
make lint

# Run go vet
make vet

# Run tests
make test

# Run all checks (fmt, vet, lint, test)
make check

# Run CI checks (vet, lint, test, test-race)
make ci
```

### 4. Commit Your Changes

Follow our [commit message guidelines](#commit-message-guidelines)

```bash
git add .
git commit -m "feat: add new feature"
```

### 5. Push and Create PR

```bash
git push origin feature/your-feature-name
```

Then create a Pull Request on GitHub.

## Code Style

### Go Code Style

- Follow standard Go conventions and idioms
- Use `gofmt` and `goimports` for formatting (run `make fmt`)
- Keep functions small and focused
- Write self-documenting code with clear variable names
- Add comments for exported functions and complex logic

### Project Structure

```
gzcli/
├── cmd/                # Command implementations
│   ├── root.go        # Root command
│   ├── init.go        # Init command
│   ├── sync.go        # Sync command
│   ├── watch*.go      # Watch commands
│   ├── team*.go       # Team commands
│   ├── script.go      # Script command
│   ├── scoreboard.go  # Scoreboard command
│   ├── structure.go   # Structure command
│   └── ...
├── internal/          # Private application code
│   ├── gzcli/         # Core logic
│   │   ├── gzapi/     # API client
│   │   ├── watcher/   # File watcher system
│   │   ├── challenge/ # Challenge management
│   │   ├── team/      # Team management
│   │   ├── config/    # Configuration
│   │   ├── event/     # Event handling
│   │   └── ...
│   ├── log/           # Logging utilities
│   ├── utils/         # Utility functions
│   └── template/      # Template system
├── scripts/           # Development scripts
│   ├── setup.sh       # Environment setup
│   ├── test.sh        # Test runner
│   ├── lint.sh        # Linter runner
│   └── install-hooks.sh # Git hooks installer
├── main.go            # Application entry point
├── Makefile           # Build automation
└── ...
```

### Adding New Commands

1. Create a new file in `cmd/` directory (e.g., `cmd/newcmd.go`)
2. Define the command using cobra:
   ```go
   var newCmd = &cobra.Command{
       Use:   "newcmd",
       Short: "Short description",
       Long:  `Long description`,
       Run: func(cmd *cobra.Command, args []string) {
           // Implementation
       },
   }

   func init() {
       rootCmd.AddCommand(newCmd)
   }
   ```
3. Add tests in `cmd/newcmd_test.go`
4. Update documentation

## Testing

For comprehensive testing guidelines, see [TESTING.md](TESTING.md).

### Writing Tests

- Write table-driven tests for complex logic
- Use meaningful test names that describe what is being tested
- Test both success and failure cases
- Mock external dependencies
- Use test utilities in `internal/gzcli/testutil/` (see [testutil/README.md](internal/gzcli/testutil/README.md))

Example test:

```go
func TestMyFunction(t *testing.T) {
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
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := MyFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("MyFunction() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("MyFunction() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

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
go test -v ./internal/gzcli/... -run TestSpecificFunction
```

## Test Environment

For integration testing, you can set up a local GZCTF test environment:

```bash
# Initialize test environment
make test-env-init

# Configure .test/.gzctf/conf.yaml with your platform settings
# Point the URL to your running GZCTF instance

# Clean test data when done
make test-env-clean
```

Note: You'll need a running GZCTF instance for integration testing. Configure the URL in `.test/.gzctf/conf.yaml` to point to your test server.

## Submitting Changes

### Pull Request Process

1. **Ensure all checks pass**
   - All tests pass
   - Code is formatted
   - No linter errors
   - Coverage is maintained or improved

2. **Write a clear PR description**
   - What changes were made?
   - Why were these changes needed?
   - How were they tested?

3. **Link related issues**
   - Reference issue numbers (e.g., "Fixes #123")

4. **Request review**
   - Tag relevant maintainers
   - Be responsive to feedback

### PR Title Format

Use conventional commit format:
- `feat: add new feature`
- `fix: resolve bug`
- `docs: update documentation`
- `refactor: improve code structure`
- `test: add tests`
- `chore: update dependencies`

## Commit Message Guidelines

We follow [Conventional Commits](https://www.conventionalcommits.org/) specification.

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- **feat**: New feature
- **fix**: Bug fix
- **docs**: Documentation changes
- **style**: Code style changes (formatting, etc.)
- **refactor**: Code refactoring
- **perf**: Performance improvements
- **test**: Adding or updating tests
- **chore**: Maintenance tasks
- **ci**: CI/CD changes

### Examples

```
feat(watch): add support for custom ignore patterns

Add ability to specify custom file patterns to ignore in the watcher.
This allows users to exclude temporary files and build artifacts.

Closes #123
```

```
fix(sync): handle empty challenge directories

Previously, the sync command would fail when encountering empty
challenge directories. Now it skips them with a warning.

Fixes #456
```

## Code of Conduct

### Our Standards

- Be respectful and inclusive
- Welcome newcomers
- Focus on constructive feedback
- Accept responsibility for mistakes
- Prioritize community well-being

### Unacceptable Behavior

- Harassment or discrimination
- Trolling or insulting comments
- Personal or political attacks
- Publishing others' private information

## Additional Resources

- **[Testing Guide](TESTING.md)**: Comprehensive testing documentation
- **[Architecture Documentation](docs/architecture.md)**: System design and architecture
- **[API Reference](docs/api-reference.md)**: Internal API documentation
- **[Development Guide](DEVELOPMENT.md)**: Detailed development information

## Getting Help

- **Questions**: Open a [GitHub Discussion](https://github.com/dimasma0305/gzcli/discussions)
- **Bugs**: File a [GitHub Issue](https://github.com/dimasma0305/gzcli/issues)
- **Security**: Report via [Security Advisories](https://github.com/dimasma0305/gzcli/security/advisories/new)

## Recognition

Contributors will be recognized in:
- GitHub contributors list
- Release notes
- Project README

Thank you for contributing to gzcli! 🚀
