# gzcli

[![CI](https://github.com/dimasma0305/gzcli/actions/workflows/ci.yml/badge.svg)](https://github.com/dimasma0305/gzcli/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dimasma0305/gzcli)](https://goreportcard.com/report/github.com/dimasma0305/gzcli)
[![GoDoc](https://godoc.org/github.com/dimasma0305/gzcli?status.svg)](https://godoc.org/github.com/dimasma0305/gzcli)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/dimasma0305/gzcli)](https://github.com/dimasma0305/gzcli/releases)

A high-performance command-line interface for [GZ::CTF](https://github.com/GZTimeWalker/GZCTF) operations with multi-event management and file watching capabilities.

## Description

gzcli is a standalone CLI tool for managing GZ::CTF challenges, providing features such as:

- **Multi-event management** - Manage multiple CTF events in one workspace
- **Challenge launcher server** - Web-based control panel with voting system
- **Discord bot integration** - Real-time CTF event notifications via webhooks
- Challenge synchronization
- File watching with automatic redeployment
- Team management and batch operations
- CTFTime scoreboard generation
- Custom script execution
- Git integration with automatic pull

## Installation

### Quick Install (Recommended)

Use the install script, which will:
- Automatically download the latest pre-built binary for your platform
- Fall back to building from source if a binary is not available
- Detect your shell and set up autocompletion

**Bash:**
```sh
bash <(curl -s https://raw.githubusercontent.com/dimasma0305/gzcli/main/install.sh)
```

**Zsh:**
```sh
zsh <(curl -s https://raw.githubusercontent.com/dimasma0305/gzcli/main/install.sh)
```

**Fish:**
```fish
curl -s https://raw.githubusercontent.com/dimasma0305/gzcli/main/install.sh | bash
```

Or download and run manually:

```sh
wget https://raw.githubusercontent.com/dimasma0305/gzcli/main/install.sh
chmod +x install.sh
./install.sh
```

### Binary Downloads

Pre-built binaries are available for multiple platforms:
- **Linux**: amd64, arm64, armv6, armv7, 386
- **macOS**: Universal Binary (Intel & Apple Silicon)
- **Windows**: amd64, 386

Download from the [releases page](https://github.com/dimasma0305/gzcli/releases/latest).

**Size:** Binaries are optimized with build flags (`-trimpath`, `-s`, `-w`), resulting in ~18 MB downloads across all platforms. See [Binary Optimization](docs/BINARY_OPTIMIZATION.md) for details.

### Uninstallation

To completely remove gzcli and all shell completions:

**Bash:**
```sh
bash <(curl -s https://raw.githubusercontent.com/dimasma0305/gzcli/main/uninstall.sh)
```

**Zsh:**
```sh
zsh <(curl -s https://raw.githubusercontent.com/dimasma0305/gzcli/main/uninstall.sh)
```

**Fish:**
```fish
curl -s https://raw.githubusercontent.com/dimasma0305/gzcli/main/uninstall.sh | bash
```

Or download and run manually:

```sh
wget https://raw.githubusercontent.com/dimasma0305/gzcli/main/uninstall.sh
chmod +x uninstall.sh
./uninstall.sh
```

The uninstall script will:
- Remove the gzcli binary from all standard installation locations
- Remove shell completions for Bash, Zsh, Fish, and PowerShell
- Clean up shell configuration files (with automatic backups)
- Provide clear feedback about what was removed

## Usage

### Quick Start

```sh
# Initialize a new CTF project
gzcli init

# Synchronize challenges to server
gzcli sync

# Start file watcher (auto-redeploy on changes)
gzcli watch start
```

### Initialize a new CTF

```sh
# Interactive mode (prompts for input)
gzcli init

# With flags
gzcli init --url https://ctf.example.com --public-entry https://public.example.com

# With all options
gzcli init --url https://ctf.example.com \
  --public-entry https://public.example.com \
  --discord-webhook https://discord.com/api/webhooks/...
```

### Synchronize challenges

```sh
# Sync challenges
gzcli sync

# Sync and update game configuration
gzcli sync --update-game
```

### File Watcher

The file watcher automatically redeploys challenges when files change.

```sh
# Start watcher as a daemon
gzcli watch start

# Start in foreground (view logs in terminal)
gzcli watch start --foreground

# Check watcher status
gzcli watch status

# View watcher logs
gzcli watch logs

# Stop watcher daemon
gzcli watch stop

# Custom configuration
gzcli watch start --debounce 5s --ignore "*.tmp" --ignore "*.log"
```

### Challenge Launcher Server

Start a web server for managing challenge launchers with real-time control and voting system.

```sh
# Start server on default localhost:8080
gzcli serve

# Start server on custom host and port
gzcli serve --host 0.0.0.0 --port 3000

# Short flags
gzcli serve -H 0.0.0.0 -p 3000
```

**Features:**
- **Real-time WebSocket communication** - Instant status updates and control
- **IP-based user tracking** - Track unique users by IP address
- **Voting system** - 50% threshold voting for challenge restarts
- **Auto-stop** - Automatically stop challenges when no users are connected
- **Restart cooldown** - Prevent restart spam with configurable cooldown periods
- **Rate limiting** - Protect against abuse with per-IP rate limits
- **Health monitoring** - Automatic health checks every 30 seconds
- **Browser notifications** - Get notified when challenges are ready

**Supported Launcher Types:**
- **Docker Compose** - Multi-container applications
- **Dockerfile** - Single container deployments
- **Kubernetes** - Advanced orchestration

**Challenge Configuration Examples:**

Docker Compose:
```yaml
# challenge.yml
dashboard:
  type: "compose"
  config: "./docker-compose.yml"
```

Dockerfile:
```yaml
# challenge.yml
dashboard:
  type: "dockerfile"
  config: "./Dockerfile"
```

Kubernetes:
```yaml
# challenge.yml
dashboard:
  type: "kubernetes"
  config: "./k8s-manifest.yaml"
```

**Port Discovery**: Ports are automatically parsed from configuration files:
- Docker Compose: Reads `ports` and `expose` from services
- Dockerfile: Parses `EXPOSE` directives
- Kubernetes: Extracts from Service port/nodePort definitions

Once the server is running, access your challenges at:
```
http://localhost:8080/<event>_<category>_<challenge_name>
```

**Note:** Challenge URLs are kept secret (not listed on homepage) for security.

### Team Management

```sh
# Create teams from CSV file
gzcli team create teams.csv

# Create teams and send registration emails
gzcli team create teams.csv --send-email

# Delete all teams and users
gzcli team delete --all
```

### Scripts

Execute custom scripts defined in challenge.yaml files:

```sh
# Run 'deploy' script for all challenges
gzcli script deploy

# Run 'test' script
gzcli script test
```

### Discord Bot

Run a Discord bot that monitors CTF events and sends real-time notifications:

```sh
# Start the bot with environment variables
export POSTGRES_PASSWORD=your_password
export GZCTF_DISCORD_WEBHOOK=https://discord.com/api/webhooks/...
gzcli bot

# Or with flags
gzcli bot --webhook $WEBHOOK_URL --db-password $DB_PASS
```

The bot monitors and sends notifications for:
- 🏆 First Blood - First team to solve a challenge
- 🥈 Second Blood - Second team to solve
- 🥉 Third Blood - Third team to solve
- 💡 New Hint - When hints are published
- 🎉 New Challenge - When challenges are published

Each notification includes the event name to distinguish between multiple CTF events. See [Discord Bot Documentation](docs/DISCORD_BOT.md) for detailed setup instructions.

### Other Commands

```sh
# Generate CTFTime scoreboard feed
gzcli scoreboard

# Generate challenge directory structure
gzcli structure
```

### Command Aliases

Save time with short aliases:

```sh
gzcli s          # same as: gzcli sync
gzcli w start    # same as: gzcli watch start
gzcli t create   # same as: gzcli team create
gzcli i          # same as: gzcli init
gzcli b          # same as: gzcli bot
```

## Configuration

### Multi-Event Structure

gzcli now supports managing multiple CTF events in a single workspace:

```
root/
├── .gzcli/          # Tool data (cache, watcher state) - git-ignored
├── .gzctf/          # Server configuration (shared)
│   └── conf.yaml    # Server URL and credentials
└── events/          # Your CTF events
    ├── ctf2024/
    │   ├── .gzevent # Event-specific configuration
    │   ├── web/     # Challenge categories
    │   ├── crypto/
    │   └── ...
    └── ctf2025/
        └── ...
```

### Server Configuration (`.gzctf/conf.yaml`)

```yaml
url: https://ctf.example.com
creds:
  username: admin
  password: your_password
```

### Event Configuration (`events/[name]/.gzevent`)

```yaml
title: "Your CTF 2024"
start: "2024-10-11T12:00:00+00:00"
end: "2024-10-13T12:00:00+00:00"
# ... more event settings
```

### Event Selection

**By default, most commands operate on ALL events.** You can control which events are processed:

```bash
# Operate on all events (default)
gzcli sync
gzcli watch start
gzcli script deploy
gzcli structure

# Operate on specific event(s)
gzcli sync --event ctf2024
gzcli sync --event ctf2024 --event ctf2025

# Exclude specific event(s)
gzcli sync --exclude-event practice
gzcli watch start --exclude-event test-event

# Set default event (for single-event commands)
gzcli event switch ctf2024
```

For detailed information about multi-event management, see [docs/MULTI_EVENT.md](docs/MULTI_EVENT.md).

### Migration from Old Structure

If you have an existing gzcli project, migrate it to the new structure:

```bash
gzcli migrate
```

For more configuration options, see the examples in the repository.

## Documentation

- [Contributing Guidelines](CONTRIBUTING.md) - How to contribute to the project
- [Development & Testing Guide](docs/development.md) - Setup, development workflow, and testing guide
- [Binary Optimization](docs/BINARY_OPTIMIZATION.md) - Binary size optimizations and compression details
- [Versioning Guide](docs/VERSIONING.md) - Automated semantic versioning
- [Performance Guide](docs/PERFORMANCE.md) - Performance optimizations
- [Architecture](docs/architecture.md) - System architecture and design
- [API Reference](docs/api-reference.md) - Internal API documentation
- [Multi-Event Management](docs/MULTI_EVENT.md) - Managing multiple CTF events

## Development

### Quick Start for Contributors

1. **Fork and clone the repository:**
   ```bash
   git clone https://github.com/YOUR_USERNAME/gzcli.git
   cd gzcli
   ```

2. **Run the setup script:**
   ```bash
   ./scripts/setup.sh
   ```

3. **Verify your environment:**
   ```bash
   make doctor
   ```

4. **Start developing:**
   ```bash
   make dev  # Hot reload mode
   ```

### Development Commands

```bash
# Setup
make setup-complete    # Complete setup with verification
make tools            # Install development tools
make doctor           # Diagnose environment issues

# Building
make build            # Build binary
make install          # Install to $GOPATH/bin

# Testing
make test             # Run all tests
make test-unit        # Run unit tests only
make test-coverage    # Generate coverage report
make coverage-browse  # Open coverage in browser
make quick-test       # Fast smoke tests

# Code Quality
make fmt              # Format code
make lint             # Run linters
make vet              # Run go vet
make check            # Run all checks

# Development
make dev              # Run with hot reload
make clean            # Clean build artifacts

# View all commands
make help
```

### Development Environment

#### VS Code

The project includes VS Code configuration for optimal Go development:
- Automatic formatting on save
- Integrated debugging
- Test explorer
- Recommended extensions

Open the project in VS Code and install the recommended extensions when prompted.

#### GitHub Codespaces

[![Open in GitHub Codespaces](https://github.com/codespaces/badge.svg)](https://github.com/codespaces/new?hide_repo_select=true&ref=main&repo=dimasma0305/gzcli)

Everything is pre-configured and ready for coding in the cloud.

### Testing

```bash
# Run all tests
make test

# Run with race detector
make test-race

# Component-specific tests
make test-watcher
make test-challenge
make test-api

# Generate coverage report
make test-coverage

# Open coverage in browser
make coverage-browse
```

See [Development & Testing Guide](docs/development.md#testing) for a comprehensive testing guide.

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

### Quick Contribution Guide

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes
4. Run tests: `make test`
5. Commit your changes: `git commit -m 'feat: add amazing feature'`
6. Push to the branch: `git push origin feature/amazing-feature`
7. Open a Pull Request

## Community

- **Issues:** [GitHub Issues](https://github.com/dimasma0305/gzcli/issues)
- **Discussions:** [GitHub Discussions](https://github.com/dimasma0305/gzcli/discussions)
- **Security:** Report vulnerabilities privately via [Security Advisories](https://github.com/dimasma0305/gzcli/security/advisories/new)

## License

Copyright © 2023 Dimas Maulana <dimasmaulana0305@gmail.com>

Licensed under the MIT License. See [LICENSE](LICENSE) for details.
