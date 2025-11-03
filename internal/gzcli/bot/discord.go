// Package bot provides Discord webhook integration for CTF event notifications
package bot

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/webhook"
	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/dimasma0305/gzcli/internal/log"
)

// Config holds the bot configuration
type Config struct {
	// Database connection settings
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string

	// Discord webhook URL
	WebhookURL string

	// Optional icon URL for embeds
	IconURL string
}

// GameNotice represents a notification from the GZ::CTF database
type GameNotice struct {
	ID             int
	NoticeType     int
	Values         []string
	PublishTimeUtc time.Time
	GameID         int
	GameTitle      string // Game title from JOIN with Games table
}

// Bot manages Discord notifications for CTF events
type Bot struct {
	config *Config
	db     *sql.DB
	client webhook.Client
}

// New creates a new Discord bot instance
func New(config *Config) (*Bot, error) {
	if config.WebhookURL == "" {
		return nil, fmt.Errorf("webhook URL is required")
	}

	// Set defaults
	if config.DBHost == "" {
		config.DBHost = "db"
	}
	if config.DBPort == 0 {
		config.DBPort = 5432
	}
	if config.DBUser == "" {
		config.DBUser = "postgres"
	}
	if config.DBName == "" {
		config.DBName = "gzctf"
	}

	// Create webhook client
	client, err := webhook.NewWithURL(config.WebhookURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook client: %w", err)
	}

	return &Bot{
		config: config,
		client: client,
	}, nil
}

// Connect establishes a connection to the PostgreSQL database
func (b *Bot) Connect() error {
	psqlconn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		b.config.DBHost,
		b.config.DBPort,
		b.config.DBUser,
		b.config.DBPassword,
		b.config.DBName,
	)

	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	b.db = db
	log.Info("Connected to the database!")
	return nil
}

// Close closes the database connection
func (b *Bot) Close() error {
	if b.db != nil {
		return b.db.Close()
	}
	return nil
}

// getLastNoticeID retrieves the highest notice ID from the database
func (b *Bot) getLastNoticeID() (int, error) {
	var lastNoticeID int
	err := b.db.QueryRow(`SELECT COALESCE(MAX("Id"), 0) FROM "GameNotices"`).Scan(&lastNoticeID)
	if err != nil {
		return 0, fmt.Errorf("failed to get last notice ID: %w", err)
	}
	return lastNoticeID, nil
}

// unlockAllTeams unlocks all teams in the database
func (b *Bot) unlockAllTeams() error {
	_, err := b.db.Exec(`UPDATE "Teams" SET "Locked" = false`)
	if err != nil {
		return fmt.Errorf("failed to unlock teams: %w", err)
	}
	return nil
}

// sendWelcomeMessage sends an initial welcome message to Discord
func (b *Bot) sendWelcomeMessage() error {
	embed := discord.NewEmbedBuilder().
		SetTitle("✅ Successfully Integrated").
		SetDescription("Connected to **CTFIFY GZCLI BOT** ~ by TCP1P Community.\n\nThank you for using our service 🚀").
		SetColor(0x00ff00).
		SetTimestamp(time.Now()).
		SetFooter("CTF Updates • by TCP1P Community", "").
		Build()

	_, err := b.client.CreateEmbeds([]discord.Embed{embed})
	if err != nil {
		return fmt.Errorf("failed to send welcome message: %w", err)
	}

	log.Info("Sent welcome message to Discord")
	return nil
}

// sanitizeTeamName removes characters that are not typically allowed in team names,
// although the specific filtering rules might need adjustment based on the platform's requirements.
func sanitizeTeamName(name string) string {
	// Remove invalid characters
	sanitized := regexp.MustCompile(`[^a-zA-Z0-9!@#$%^&*()_+\-={}\[\]:"';<>,.?/\\]`).ReplaceAllString(name, "")
	return sanitized
}

// sanitizeNoticeValues sanitizes notice values to prevent Discord pings
func sanitizeNoticeValues(notice *GameNotice) {
	replacements := map[string]string{
		"@everyone": "@everyon3",
		"@here":     "@her3",
	}

	// Sanitize team name if this is a team-related notice
	if notice.NoticeType >= 1 && notice.NoticeType <= 3 && len(notice.Values) > 0 {
		sanitized := sanitizeTeamName(notice.Values[0])
		if sanitized != notice.Values[0] {
			log.Debug("Sanitized team name from '%s' to '%s'", notice.Values[0], sanitized)
			notice.Values[0] = sanitized
		}
	}

	// Replace Discord mentions
	for i := range notice.Values {
		for old, new := range replacements {
			notice.Values[i] = strings.ReplaceAll(notice.Values[i], old, new)
		}
	}
}

// createEmbed creates a Discord embed for a game notice
func (b *Bot) createEmbed(notice *GameNotice) *discord.Embed {
	icon := b.config.IconURL
	if icon == "" {
		icon = "https://tcp1p.team/_next/static/media/TCP1P%20_Main%20White%20Red.89fd023d.svg"
	}

	var embed discord.Embed
	switch notice.NoticeType {
	case 1: // First Blood
		embed = discord.NewEmbedBuilder().
			SetTitle("🏆 First Blood!").
			SetDescription(fmt.Sprintf("Team **%s** was the first to solve:", notice.Values[0])).
			AddField("Challenge", fmt.Sprintf("`%s`", notice.Values[1]), false).
			AddField("Event", fmt.Sprintf("`%s`", notice.GameTitle), false).
			SetColor(0xE74C3C). // red
			SetFooter("CTF Updates • by TCP1P Community", icon).
			SetTimestamp(time.Now()).
			Build()

	case 2: // Second Blood
		embed = discord.NewEmbedBuilder().
			SetTitle("🥈 Second Blood").
			SetDescription(fmt.Sprintf("Team **%s** claimed the second solve:", notice.Values[0])).
			AddField("Challenge", fmt.Sprintf("`%s`", notice.Values[1]), false).
			AddField("Event", fmt.Sprintf("`%s`", notice.GameTitle), false).
			SetColor(0xF1C40F). // gold
			SetFooter("CTF Updates • by TCP1P Community", icon).
			SetTimestamp(time.Now()).
			Build()

	case 3: // Third Blood
		embed = discord.NewEmbedBuilder().
			SetTitle("🥉 Third Blood").
			SetDescription(fmt.Sprintf("Team **%s** secured the third solve:", notice.Values[0])).
			AddField("Challenge", fmt.Sprintf("`%s`", notice.Values[1]), false).
			AddField("Event", fmt.Sprintf("`%s`", notice.GameTitle), false).
			SetColor(0x2ECC71). // green
			SetFooter("CTF Updates • by TCP1P Community", icon).
			SetTimestamp(time.Now()).
			Build()

	case 4: // New Hint
		embed = discord.NewEmbedBuilder().
			SetTitle("💡 New Hint Released").
			SetDescription("A hint was just published for a challenge:").
			AddField("Challenge", fmt.Sprintf("`%s`", notice.Values[0]), false).
			AddField("Event", fmt.Sprintf("`%s`", notice.GameTitle), false).
			SetColor(0x3498DB). // blue
			SetFooter("CTF Updates • by TCP1P Community", icon).
			SetTimestamp(time.Now()).
			Build()

	case 5: // New Challenge
		embed = discord.NewEmbedBuilder().
			SetTitle("🎉 New Challenge Available!").
			SetDescription("A new challenge has just been published:").
			AddField("Challenge", fmt.Sprintf("`%s`", notice.Values[0]), false).
			AddField("Event", fmt.Sprintf("`%s`", notice.GameTitle), false).
			SetColor(0x9B59B6). // purple
			SetFooter("CTF Updates • by TCP1P Community", icon).
			SetTimestamp(time.Now()).
			Build()

	default:
		return nil
	}

	return &embed
}

// fetchNewNotices retrieves new notices from the database with game title
func (b *Bot) fetchNewNotices(lastNoticeID int) ([]GameNotice, error) {
	query := `
		SELECT n."Id", n."Type", n."Values", n."PublishTimeUtc", n."GameId", g."Title"
		FROM "GameNotices" n
		INNER JOIN "Games" g ON n."GameId" = g."Id"
		WHERE n."Id" > $1
		ORDER BY n."Id" ASC
	`

	rows, err := b.db.Query(query, lastNoticeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query notices: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var notices []GameNotice
	for rows.Next() {
		var notice GameNotice
		var values string

		err := rows.Scan(&notice.ID, &notice.NoticeType, &values, &notice.PublishTimeUtc, &notice.GameID, &notice.GameTitle)
		if err != nil {
			log.Error("Error scanning row: %v", err)
			continue
		}

		err = json.Unmarshal([]byte(values), &notice.Values)
		if err != nil {
			log.Error("Error unmarshalling values: %v", err)
			continue
		}

		notices = append(notices, notice)
	}

	return notices, nil
}

// processNotices processes and sends Discord notifications for notices
func (b *Bot) processNotices(notices []GameNotice) {
	for _, notice := range notices {
		sanitizeNoticeValues(&notice)

		embed := b.createEmbed(&notice)
		if embed == nil {
			log.Debug("Skipping unknown notice type: %d", notice.NoticeType)
			continue
		}

		_, err := b.client.CreateEmbeds([]discord.Embed{*embed})
		if err != nil {
			log.Error("Error sending webhook: %v", err)
		} else {
			log.Debug("Sent notification for notice ID %d (type %d)", notice.ID, notice.NoticeType)
		}
	}
}

// Run starts the bot's main loop
func (b *Bot) Run() error {
	// Connect with retry
	for {
		err := b.Connect()
		if err != nil {
			log.Error("Failed to connect to database: %v", err)
			log.Info("Retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}
	defer func() {
		_ = b.Close()
	}()

	// Get initial notice ID
	lastNoticeID, err := b.getLastNoticeID()
	if err != nil {
		return fmt.Errorf("failed to get initial notice ID: %w", err)
	}
	log.Info("Starting from notice ID: %d", lastNoticeID)

	// Send welcome message
	if err := b.sendWelcomeMessage(); err != nil {
		log.Info("Failed to send welcome message: %v", err)
	}

	// Main monitoring loop
	log.Info("Bot is now monitoring for CTF events...")
	pollInterval := 1 * time.Second

	for {
		// Fetch new notices
		newNotices, err := b.fetchNewNotices(lastNoticeID)
		if err != nil {
			log.Error("Error fetching notices: %v", err)
			time.Sleep(pollInterval)
			continue
		}

		// Unlock all teams
		if err := b.unlockAllTeams(); err != nil {
			log.Info("Failed to unlock teams: %v", err)
		}

		// Process notices
		if len(newNotices) > 0 {
			log.Info("Processing %d new notice(s)", len(newNotices))
			b.processNotices(newNotices)
			lastNoticeID = newNotices[len(newNotices)-1].ID
		}

		time.Sleep(pollInterval)
	}
}
