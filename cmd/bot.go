package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/dimasma0305/gzcli/internal/gzcli/bot"
	"github.com/dimasma0305/gzcli/internal/log"
)

var (
	botDBHost     string
	botDBPort     int
	botDBUser     string
	botDBPassword string
	botDBName     string
	botWebhookURL string
	botIconURL    string
)

var botCmd = &cobra.Command{
	Use:     "bot",
	Aliases: []string{"b"},
	Short:   "Start Discord bot for CTF event notifications",
	Long: `Start a Discord bot that monitors the GZ::CTF database for events and sends notifications.

The bot monitors the following events:
  • First Blood (🏆) - First team to solve a challenge
  • Second Blood (🥈) - Second team to solve a challenge
  • Third Blood (🥉) - Third team to solve a challenge
  • New Hint (💡) - When a hint is published for a challenge
  • New Challenge (🎉) - When a new challenge is published

The bot requires:
  1. PostgreSQL database connection (typically the GZ::CTF database)
  2. Discord webhook URL for sending notifications

Configuration can be provided via flags or environment variables:
  • POSTGRES_PASSWORD - Database password (recommended)
  • GZCTF_DISCORD_WEBHOOK - Discord webhook URL (required)`,
	Example: `  # Start bot with environment variables
  export POSTGRES_PASSWORD=mysecret
  export GZCTF_DISCORD_WEBHOOK=https://discord.com/api/webhooks/...
  gzcli bot

  # Start bot with custom database settings
  gzcli bot --db-host localhost --db-port 5432 --webhook $WEBHOOK_URL

  # Start bot with custom icon
  gzcli bot --icon-url https://example.com/logo.png`,
	PreRun: func(_ *cobra.Command, _ []string) {
		// Get webhook URL from environment if not provided
		if botWebhookURL == "" {
			botWebhookURL = os.Getenv("GZCTF_DISCORD_WEBHOOK")
		}

		// Get database password from environment if not provided
		if botDBPassword == "" {
			botDBPassword = os.Getenv("POSTGRES_PASSWORD")
		}
	},
	Run: func(_ *cobra.Command, _ []string) {
		// Validate required configuration
		if botWebhookURL == "" {
			log.Error("Discord webhook URL is required")
			log.Error("Set via --webhook flag or GZCTF_DISCORD_WEBHOOK environment variable")
			os.Exit(1)
		}

		if botDBPassword == "" {
			log.Info("Database password not set (POSTGRES_PASSWORD env var)")
			log.Info("This may cause connection issues if the database requires authentication")
		}

		// Create bot configuration
		config := &bot.Config{
			DBHost:     botDBHost,
			DBPort:     botDBPort,
			DBUser:     botDBUser,
			DBPassword: botDBPassword,
			DBName:     botDBName,
			WebhookURL: botWebhookURL,
			IconURL:    botIconURL,
		}

		// Create bot instance
		b, err := bot.New(config)
		if err != nil {
			log.Error("Failed to create bot: %v", err)
			os.Exit(1)
		}

		// Setup signal handling for graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		// Run bot in a goroutine
		errChan := make(chan error, 1)
		go func() {
			errChan <- b.Run()
		}()

		// Wait for completion or signal
		select {
		case err := <-errChan:
			if err != nil {
				log.Error("Bot stopped with error: %v", err)
				os.Exit(1)
			}
		case sig := <-sigChan:
			log.Info("Received signal %v, shutting down...", sig)
			if err := b.Close(); err != nil {
				log.Error("Error closing bot: %v", err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(botCmd)

	// Database connection flags
	botCmd.Flags().StringVar(&botDBHost, "db-host", "db", "Database host")
	botCmd.Flags().IntVar(&botDBPort, "db-port", 5432, "Database port")
	botCmd.Flags().StringVar(&botDBUser, "db-user", "postgres", "Database user")
	botCmd.Flags().StringVar(&botDBPassword, "db-password", "", "Database password (or set POSTGRES_PASSWORD env var)")
	botCmd.Flags().StringVar(&botDBName, "db-name", "gzctf", "Database name")

	// Discord configuration flags
	botCmd.Flags().StringVarP(&botWebhookURL, "webhook", "w", "", "Discord webhook URL (required, or set GZCTF_DISCORD_WEBHOOK env var)")
	botCmd.Flags().StringVar(&botIconURL, "icon-url", "", "Custom icon URL for Discord embeds")

	// Mark webhook as required (will be validated in PreRun)
	_ = botCmd.MarkFlagRequired("webhook")
}
