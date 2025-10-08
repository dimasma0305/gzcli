# Multi-Event Management

gzcli supports managing multiple CTF events in a single workspace. This guide explains the multi-event structure and how to work with it.

## Directory Structure

```
root/
├── .gzcli/                     # Tool data (git-ignored)
│   ├── cache/                 # API cache files
│   ├── watcher/               # Watcher daemon files
│   │   ├── watcher.pid
│   │   ├── watcher.log
│   │   ├── watcher.db
│   │   └── watcher.sock
│   └── current-event          # Default event selection
│
├── .gzctf/                     # Server configuration (shared)
│   ├── conf.yaml              # Server URL and credentials
│   ├── appsettings.json       # Server settings
│   ├── docker-compose.yml     # Docker setup
│   ├── challenge.schema.yaml  # Challenge schema
│   ├── conf.schema.yaml       # Server config schema
│   ├── gzevent.schema.yaml    # Event config schema
│   └── ...
│
└── events/                     # Event management
    ├── ctf2024/
    │   ├── .git/              # Event-specific git repo
    │   ├── .gzevent           # Event configuration
    │   ├── web/               # Challenge categories
    │   ├── crypto/
    │   ├── pwn/
    │   └── ...
    └── ctf2025/
        ├── .git/
        ├── .gzevent
        └── ...
```

## Configuration Files

### `.gzctf/conf.yaml` (Server Config)

Shared across all events. Contains server connection information:

```yaml
url: "https://ctf.example.com"
creds:
  username: "admin"
  password: "password"
```

### `events/[name]/.gzevent` (Event Config)

Event-specific configuration:

```yaml
title: "Example CTF 2024"
start: "2024-10-11T12:00:00+00:00"
end: "2024-10-13T12:00:00+00:00"
poster: "../../.gzctf/favicon.ico"
hidden: false
summary: "example summary"
content: |
  Detailed description...
inviteCode: ""
organizations:
  - "Miku fans club"
teamMemberCountLimit: 0
containerCountLimit: 3
practiceMode: true
writeupRequired: false
```

## Event Selection

### Default Behavior

**Most commands now operate on ALL events by default.** This includes:
- `sync` - Synchronize all events
- `watch start` - Watch all events
- `script` - Run scripts for all events
- `structure` - Generate structures for all events

### Selection Methods

You can control which events are processed using:

1. **Specific events** (highest priority):
   ```bash
   gzcli sync --event ctf2024
   gzcli sync --event ctf2024 --event ctf2025
   ```

2. **Exclude events**:
   ```bash
   gzcli sync --exclude-event practice
   gzcli watch start --exclude-event test-event --exclude-event staging
   ```

3. **No flags** (default):
   ```bash
   gzcli sync  # Operates on ALL events
   ```

### Single-Event Commands

Some commands still work with single events (for backward compatibility):
- `team create` - Creates teams for one event
- Other management commands

For these commands, event selection follows this priority:

1. **Command-line flag**:
   ```bash
   gzcli team create teams.csv --event ctf2024
   ```

2. **Environment variable**:
   ```bash
   export GZCLI_EVENT=ctf2024
   gzcli team create teams.csv
   ```

3. **Default event file** (`.gzcli/current-event`):
   ```bash
   gzcli event switch ctf2024  # Sets default
   gzcli team create teams.csv # Uses default
   ```

4. **Auto-detection**:
   - If only one event exists, it's automatically selected
   - If multiple events exist without selection, an error is shown

### Event Commands

```bash
# List all events
gzcli event list

# Show current event
gzcli event current

# Switch default event
gzcli event switch ctf2024

# Create new event (coming soon)
gzcli event create ctf2025
```

## Working with Events

### Default Multi-Event Behavior

**As of the latest version, most commands operate on ALL events by default.** This includes:
- `gzcli sync` - Syncs all events
- `gzcli watch start` - Watches all events simultaneously
- `gzcli script <name>` - Runs scripts for all events
- `gzcli structure` - Generates structures for all events

### Syncing Challenges

```bash
# Sync all events (default)
gzcli sync

# Sync specific event(s)
gzcli sync --event ctf2024
gzcli sync --event ctf2024 --event ctf2025

# Sync all except specific event(s)
gzcli sync --exclude-event practice --exclude-event test
```

### File Watching

```bash
# Watch all events (default)
gzcli watch start

# Watch specific event(s)
gzcli watch start --event ctf2024 --event ctf2025

# Watch all except specific event(s)
gzcli watch start --exclude-event practice
```

### Running Scripts

```bash
# Run script for all events (default)
gzcli script deploy

# Run script for specific event(s)
gzcli script test --event ctf2024

# Run script for all except specific event(s)
gzcli script cleanup --exclude-event production
```

### Generating Structures

```bash
# Generate structure for all events (default)
gzcli structure

# Generate for specific event(s)
gzcli structure --event ctf2024

# Generate for all except specific event(s)
gzcli structure --exclude-event practice
```

### Team Management

```bash
# Create teams for specific event (single-event command)
gzcli team create teams.csv --event ctf2024

# Create teams for default event
gzcli team create teams.csv
```

## Cache Management

Caches are event-specific and stored in `.gzcli/cache/`:

- `config-[event].yaml` - Event configuration cache
- `teams-[event].yaml` - Team creation cache
- Other event-specific cache files

This allows you to work with multiple events without cache conflicts.

## Git Integration

Each event can have its own git repository in `events/[name]/.git/`. This allows:

- Separate version control for each event
- Independent collaboration per event
- Event-specific commit history

```bash
cd events/ctf2024
git init
git remote add origin https://github.com/your-org/ctf2024-challenges.git
git add .
git commit -m "Initial challenges"
git push -u origin main
```

## Migration from Single Event

If you have an existing gzcli project with challenges at the root level, you can migrate:

```bash
# Coming soon: gzcli migrate
```

Manual migration:

1. Create events directory:
   ```bash
   mkdir -p events/ctf2024
   ```

2. Move challenges:
   ```bash
   mv web crypto pwn misc events/ctf2024/
   ```

3. Split config:
   - Extract `url` and `creds` from `.gzctf/conf.yaml`
   - Move `event` section to `events/ctf2024/.gzevent`

4. Move cache:
   ```bash
   mkdir -p .gzcli/cache
   mv .gzcli/*.yaml .gzcli/cache/
   ```

## Best Practices

### 1. Use Event Names Consistently

```bash
# Good
events/ctf2024/
events/ctf2025/
events/training-2024/

# Avoid
events/CTF_2024/
events/ctf-2024-final/
```

### 2. Share Server Configuration

Keep `.gzctf/conf.yaml` consistent across events since they share the same server.

### 3. Separate Git Repositories

Initialize separate git repos for each event to maintain clean history:

```bash
cd events/ctf2024
git init
```

### 4. Use Default Event

Set a default event during active development:

```bash
gzcli event switch ctf2024
# Now all commands use ctf2024 by default
```

### 5. Document Event Purpose

Add a README in each event directory:

```bash
events/ctf2024/README.md
```

## Troubleshooting

### "Multiple events found" Error

```bash
$ gzcli sync
Error: multiple events found: [ctf2024, ctf2025]. Please specify with --event flag

# Solution 1: Use flag
$ gzcli sync --event ctf2024

# Solution 2: Set default
$ gzcli event switch ctf2024
$ gzcli sync
```

### Cache Conflicts

If you experience cache issues:

```bash
# Clear cache for specific event
rm .gzcli/cache/config-ctf2024.yaml

# Clear all cache
rm -rf .gzcli/cache/*
```

### Event Not Found

```bash
$ gzcli sync --event ctf2026
Error: event ctf2026 does not exist

# List available events
$ gzcli event list
```

## Advanced Usage

### Running Multiple Events Simultaneously

Use separate terminal sessions with different environment variables:

```bash
# Terminal 1
export GZCLI_EVENT=ctf2024
gzcli watch start

# Terminal 2
export GZCLI_EVENT=ctf2025
gzcli watch start
```

### Event-Specific Scripts

Add scripts to event directories:

```bash
events/ctf2024/scripts/deploy.sh
events/ctf2024/scripts/test.sh
```

### Shared Resources

For resources shared across events:

```bash
events/shared/docker-images/
events/shared/scripts/
```

Reference in challenges using relative paths:

```yaml
# events/ctf2024/web/challenge1/challenge.yaml
container:
  containerImage: ../../shared/docker-images/web-base
```

## See Also

- [Architecture Documentation](architecture.md)
- [Contributing Guidelines](../CONTRIBUTING.md)
- [Development Guide](../DEVELOPMENT.md)
