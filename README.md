# gzcli

[![CI](https://github.com/dimasma0305/gzcli/actions/workflows/ci.yml/badge.svg)](https://github.com/dimasma0305/gzcli/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dimasma0305/gzcli)](https://goreportcard.com/report/github.com/dimasma0305/gzcli)
[![GoDoc](https://godoc.org/github.com/dimasma0305/gzcli?status.svg)](https://godoc.org/github.com/dimasma0305/gzcli)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/dimasma0305/gzcli)](https://github.com/dimasma0305/gzcli/releases)

A high-performance command-line interface for [GZ::CTF](https://github.com/GZTimeWalker/GZCTF) with multi-event management and file watching capabilities.

## Overview

gzcli is a powerful, standalone CLI tool designed to streamline the entire lifecycle of managing a GZ::CTF-based competition. It simplifies everything from initial setup to real-time event monitoring, making it an indispensable tool for CTF organizers.

### Key Features:
- **Multi-Event Workspace**: Manage multiple, distinct CTF events from a single, organized workspace.
- **Challenge Launcher Server**: A web-based control panel for starting, stopping, and managing challenge containers, complete with a voting system for restarts.
- **Discord Bot Integration**: Keep participants informed with real-time notifications for key events like First Bloods, new challenges, and hint releases.
- **Automated Synchronization**: Seamlessly sync local challenge configurations with the GZ::CTF server.
- **File Watcher with Hot-Reload**: Automatically redeploy challenges and apply updates when files are modified.
- **Comprehensive Team Management**: Easily handle batch operations, including creating teams from CSV files.
- **CTFTime Integration**: Generate a CTFTime-compatible scoreboard feed in JSON format.
- **Custom Scripting**: Execute custom scripts defined within your challenge configurations for flexible automation.
- **Git Integration**: Keep your challenges up-to-date with automatic git pull functionality.

## Installation

### Quick Install (Recommended)

Our intelligent install script automates the download of the latest pre-built binary for your platform, with a fallback to building from source if needed. It also detects your shell to set up autocompletion.

**For Bash:**
```sh
bash <(curl -s https://raw.githubusercontent.com/dimasma0305/gzcli/main/install.sh)
```

**For Zsh:**
```sh
zsh <(curl -s https://raw.githubusercontent.com/dimasma0305/gzcli/main/install.sh)
```

**For Fish:**
```fish
curl -s https://raw.githubusercontent.com/dimasma0305/gzcli/main/install.sh | bash
```

Alternatively, you can download and run the script manually:
```sh
wget https://raw.githubusercontent.com/dimasma0305/gzcli/main/install.sh
chmod +x install.sh
./install.sh
```

### Manual Installation (Pre-built Binaries)

We provide pre-built binaries for a wide range of platforms:
- **Linux**: amd64, arm64, armv6, armv7, 386
- **macOS**: Universal Binary (Intel & Apple Silicon)
- **Windows**: amd64, 386

You can download the appropriate binary for your system from our [Releases Page](https://github.com/dimasma0305/gzcli/releases/latest). These binaries are optimized for size, typically around 18 MB.

### Uninstallation

To completely remove gzcli, including all shell completions and configuration files, use our uninstall script. It provides clear feedback on the actions taken and creates automatic backups of any modified shell configurations.

**For Bash:**
```sh
bash <(curl -s https://raw.githubusercontent.com/dimasma0305/gzcli/main/uninstall.sh)
```

**For Zsh:**
```sh
zsh <(curl -s https://raw.githubusercontent.com/dimasma0305/gzcli/main/uninstall.sh)
```

**For Fish:**
```fish
curl -s https://raw.githubusercontent.com/dimasma0305/gzcli/main/uninstall.sh | bash
```

## Getting Started

### Initializing Your Project

To start a new CTF project, run the `init` command. You can do this interactively or by providing flags for a non-interactive setup.

```sh
# Interactively set up your project
gzcli init

# Set up using flags
gzcli init --url https://ctf.example.com --public-entry https://public.example.com

# Include a Discord webhook for notifications
gzcli init --url https://ctf.example.com \
  --public-entry https://public.example.com \
  --discord-webhook https://discord.com/api/webhooks/...
```

### Synchronizing Challenges

The `sync` command is used to push your local challenge configurations to the GZ::CTF server.

```sh
# Synchronize all challenges
gzcli sync

# Sync and update the game configuration on the server
gzcli sync --update-game
```

### Automating with the File Watcher

The file watcher is a powerful feature that monitors your project for changes and automatically redeploys challenges as you work.

```sh
# Start the watcher as a background daemon
gzcli watch start

# For debugging, run in the foreground to see live logs
gzcli watch start --foreground

# Check the current status of the watcher
gzcli watch status

# Tail the watcher logs
gzcli watch logs

# Stop the watcher daemon
gzcli watch stop

# Customize the watcher with a 5-second debounce and ignore patterns
gzcli watch start --debounce 5s --ignore "*.tmp" --ignore "*.log"
```

## Advanced Features

### Challenge Launcher Server

gzcli includes a built-in web server that provides a user-friendly interface for managing challenge launchers. It features real-time controls and a voting system for restarts.

```sh
# Start the server on the default address (localhost:8080)
gzcli serve

# Bind the server to a different host and port
gzcli serve --host 0.0.0.0 --port 3000

# Use short flags for convenience
gzcli serve -H 0.0.0.0 -p 3000
```

#### Launcher Features:
- **Real-time Control**: Instant status updates and actions via WebSockets.
- **User Tracking**: Monitor unique users interacting with challenges by their IP address.
- **Voting System**: A democratic restart system requiring a 50% vote threshold.
- **Resource Management**: Automatically stops challenges when they are no longer in use.
- **Cooldowns & Rate Limiting**: Protect against spam and abuse with configurable limits.
- **Health Monitoring**: Continuous health checks ensure challenges are running correctly.
- **Notifications**: Receive browser notifications when challenges are ready.

#### Supported Launcher Types:
- **Docker Compose**: For multi-container challenge setups.
- **Dockerfile**: For single-container challenges.
- **Kubernetes**: For challenges requiring advanced orchestration.

Once the server is active, you can access your challenges at `http://localhost:8080/<event>_<category>_<challenge_name>`. Note that challenge URLs are not publicly listed for security.

### Team Management

Manage your teams efficiently with batch operations.

```sh
# Create multiple teams at once from a CSV file
gzcli team create teams.csv

# Create teams and automatically send registration emails
gzcli team create teams.csv --send-email

# A failsafe command to delete all teams and users
gzcli team delete --all
```

### Custom Scripts

Define and execute custom scripts within your `challenge.yaml` files for enhanced automation.

```sh
# Run the 'deploy' script for all defined challenges
gzcli script deploy

# Run a 'test' script
gzcli script test
```

### Discord Bot

Keep your community engaged with a Discord bot that announces key CTF events.

```sh
# Start the bot using environment variables for configuration
export POSTGRES_PASSWORD=your_password
export GZCTF_DISCORD_WEBHOOK=https://discord.com/api/webhooks/...
gzcli bot

# Alternatively, configure using flags
gzcli bot --webhook $WEBHOOK_URL --db-password $DB_PASS
```

The bot will post notifications for:
- 🏆 First Blood
- 🥈 Second Blood
- 🥉 Third Blood
- 💡 New Hints
- 🎉 New Challenges

For more details, refer to the [Discord Bot Documentation](docs/DISCORD_BOT.md).

### Additional Commands

```sh
# Generate a CTFTime-compatible scoreboard feed
gzcli scoreboard

# Automatically generate your challenge directory structure
gzcli structure
```

### Command Aliases

For a faster workflow, use these short aliases:
- `gzcli s`: `gzcli sync`
- `gzcli w start`: `gzcli watch start`
- `gzcli t create`: `gzcli team create`
- `gzcli i`: `gzcli init`
- `gzcli b`: `gzcli bot`

## Configuration

### Multi-Event Workspace Structure

gzcli is designed to manage multiple CTF events within a single, organized workspace.

```
my-ctf-workspace/
├── .gzcli/          # Tool-specific data (cache, watcher state) - should be git-ignored.
├── .gzctf/          # Shared server configuration.
│   └── conf.yaml    # Server URL and admin credentials.
└── events/          # Directory for all your CTF events.
    ├── ctf2024/
    │   ├── .gzevent # Configuration specific to this event.
    │   ├── web/     # Challenge categories for this event.
    │   ├── crypto/
    │   └── ...
    └── my-other-ctf/
        └── ...
```

### Server Configuration (`.gzctf/conf.yaml`)
This file contains the shared configuration for your GZ::CTF server.
```yaml
url: https://ctf.example.com
creds:
  username: admin
  password: your_admin_password
```

### Event Configuration (`events/[name]/.gzevent`)
Each event has its own configuration file.
```yaml
title: "My Awesome CTF 2024"
start: "2024-12-01T00:00:00Z"
end: "2024-12-03T00:00:00Z"
# Additional event settings...
```

### Event Selection

By default, commands that can operate on multiple challenges (like `sync`, `watch`, `script`, and `structure`) will run on **all events**. You can target specific events using flags.

```bash
# The default is to operate on all events
gzcli sync

# Target one or more specific events
gzcli sync --event ctf2024
gzcli sync --event ctf2024 --event my-other-ctf

# Exclude certain events
gzcli watch start --exclude-event practice-ctf

# Set a default event for commands that operate on a single event
gzcli event switch ctf2024
```

For a more in-depth guide on this feature, see our [Multi-Event Management Documentation](docs/MULTI_EVENT.md).

### Migrating from an Old Structure

If you have a project created with an older version of gzcli, you can easily migrate it to the new multi-event structure with a single command:
```bash
gzcli migrate
```

## Contributing

We welcome contributions of all kinds! Please read our [Contributing Guidelines](CONTRIBUTING.md) to get started.

### Quick Start for Developers

1.  **Fork and clone the repository:**
    ```bash
    git clone https://github.com/YOUR_USERNAME/gzcli.git
    cd gzcli
    ```
2.  **Run the setup script:**
    ```bash
    ./scripts/setup.sh
    ```
3.  **Verify your setup:**
    ```bash
    make doctor
    ```
4.  **Start developing with hot-reload:**
    ```bash
    make dev
    ```

### Development Commands

Our `Makefile` provides a comprehensive set of commands for development and testing.
```bash
# Setup and Verification
make setup-complete  # Run the full setup and verification process.
make tools           # Install all necessary development tools.
make doctor          # Diagnose any issues with your environment.

# Building and Installation
make build           # Build the binary.
make install         # Install the binary to $GOPATH/bin.

# Testing
make test            # Run all tests.
make test-unit       # Run only unit tests.
make test-coverage   # Generate a test coverage report.
make coverage-browse # Open the coverage report in your browser.

# Code Quality
make fmt             # Format all code.
make lint            # Run all linters.
make vet             # Run go vet.
make check           # Run all code quality checks.

# Development Workflow
make dev             # Run in development mode with hot-reloading.
make clean           # Remove all build artifacts.

# Get a list of all available commands
make help
```

### Development Environments

#### VS Code

This project is optimized for development in VS Code. It includes configurations for automatic formatting, integrated debugging, a test explorer, and a list of recommended extensions. Simply open the project in VS Code and follow the prompts to install the recommended extensions.

#### GitHub Codespaces

[![Open in GitHub Codespaces](https://github.com/codespaces/badge.svg)](https://github.com/codespaces/new?hide_repo_select=true&ref=main&repo=dimasma0305/gzcli)

For a zero-config setup, you can use GitHub Codespaces, which provides a fully pre-configured development environment in the cloud.

## Community and Support

- **Bug Reports & Feature Requests**: [GitHub Issues](https://github.com/dimasma0305/gzcli/issues)
- **General Discussion**: [GitHub Discussions](https://github.com/dimasma0305/gzcli/discussions)
- **Security**: To report a security vulnerability, please use [GitHub Security Advisories](https://github.com/dimasma0305/gzcli/security/advisories/new) for private disclosure.

## License

Copyright © 2023 Dimas Maulana <dimasmaulana0305@gmail.com>

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
