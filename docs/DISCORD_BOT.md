# Discord Bot Integration

The `gzcli bot` command provides real-time Discord notifications for CTF events by monitoring the GZ::CTF database.

## Features

The bot monitors and sends Discord webhook notifications for:

- 🏆 **First Blood** - First team to solve a challenge
- 🥈 **Second Blood** - Second team to solve a challenge
- 🥉 **Third Blood** - Third team to solve a challenge
- 💡 **New Hint** - When a hint is published for a challenge
- 🎉 **New Challenge** - When a new challenge is published

## Requirements

1. **PostgreSQL Database Access**: The bot needs to connect to the GZ::CTF PostgreSQL database
2. **Discord Webhook URL**: A webhook URL from your Discord server

## Setup

### 1. Create Discord Webhook

1. In your Discord server, go to Server Settings → Integrations → Webhooks
2. Click "New Webhook"
3. Give it a name (e.g., "CTF Bot")
4. Select the channel where notifications should appear
5. Copy the webhook URL

### 2. Configure Database Connection

The bot needs access to the GZ::CTF PostgreSQL database. By default, it assumes:
- Host: `db` (Docker container name)
- Port: `5432`
- User: `postgres`
- Database: `gzctf`

These can be customized via command-line flags.

### 3. Set Environment Variables

For security, use environment variables for sensitive data:

```bash
export POSTGRES_PASSWORD=your_database_password
export GZCTF_DISCORD_WEBHOOK=https://discord.com/api/webhooks/...
```

## Usage

### Basic Usage

```bash
# With environment variables set
gzcli bot

# Or using flags
gzcli bot --webhook "https://discord.com/api/webhooks/..." --db-password "password"
```

### Docker Compose Example

Add the bot as a service in your `docker-compose.yml`:

```yaml
services:
  gzctf-bot:
    image: your-registry/gzcli:latest
    command: ["bot"]
    environment:
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
      - GZCTF_DISCORD_WEBHOOK=${DISCORD_WEBHOOK}
    depends_on:
      - db
    networks:
      - gzctf-network
```

### Custom Configuration

```bash
# Custom database connection
gzcli bot \
  --db-host localhost \
  --db-port 5432 \
  --db-user myuser \
  --db-password mypass \
  --webhook "https://discord.com/api/webhooks/..."

# With custom icon for embeds
gzcli bot \
  --icon-url "https://example.com/logo.png" \
  --webhook "$WEBHOOK_URL"
```

## Command-Line Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--db-host` | Database host | `db` |
| `--db-port` | Database port | `5432` |
| `--db-user` | Database user | `postgres` |
| `--db-password` | Database password | (from `POSTGRES_PASSWORD` env) |
| `--db-name` | Database name | `gzctf` |
| `-w, --webhook` | Discord webhook URL | (from `GZCTF_DISCORD_WEBHOOK` env) |
| `--icon-url` | Custom icon for embeds | TCP1P logo |

## Security Considerations

1. **Protect Your Webhook URL**: Never commit webhook URLs to version control
2. **Use Environment Variables**: Store sensitive data in environment variables
3. **Network Access**: Ensure the bot has network access to both the database and Discord
4. **Database Permissions**: The bot only needs `SELECT` and `UPDATE` permissions on specific tables

## Notification Format

Each notification type has a distinct color and format:

- **First Blood**: Red embed with team name, challenge, and event name
- **Second Blood**: Gold embed with team name, challenge, and event name
- **Third Blood**: Green embed with team name, challenge, and event name
- **New Hint**: Blue embed with challenge name and event name
- **New Challenge**: Purple embed with challenge name and event name

All notifications include the event name (from the `Games` table) to help distinguish between multiple CTF events running on the same server.

Team names are automatically sanitized to prevent Discord mentions (e.g., `@everyone` → `@everyon3`).

## Troubleshooting

### Bot won't connect to database

- Verify database host and port are correct
- Check if `POSTGRES_PASSWORD` is set
- Ensure the bot has network access to the database
- For Docker: verify the bot is on the same network as the database

### Webhook messages not appearing

- Verify the webhook URL is correct and not expired
- Check if the webhook's target channel still exists
- Ensure the bot has internet access to reach Discord

### Bot shows "Failed to unlock teams"

This is a warning, not an error. It attempts to unlock teams but won't stop the bot if it fails.

## Development

The bot implementation is located in:
- Command: `cmd/bot.go`
- Core logic: `internal/gzcli/bot/discord.go`

To add new notification types:
1. Add the case in `createEmbed()` function
2. Ensure the database query includes the new notice type
3. Update documentation

## Related Commands

- `gzcli sync` - Sync challenges to server
- `gzcli watch` - Watch for file changes
- `gzcli team` - Manage teams

## Credits

Bot integration developed by the TCP1P Community for CTF organizers worldwide.
