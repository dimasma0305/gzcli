package cmd

import (
	"github.com/spf13/cobra"

	"github.com/dimasma0305/gzcli/internal/log"
	"github.com/dimasma0305/gzcli/internal/template/other"
)

var (
	initURL            string
	initPublicEntry    string
	initDiscordWebhook string
)

var initCmd = &cobra.Command{
	Use:     "init",
	Aliases: []string{"i"},
	Short:   "Initialize a new CTF project structure",
	Long: `Initialize a new CTF project with configuration files and directory structure.

This command creates:
  - .gzctf/ directory with server configuration files
  - Makefile with helpful commands
  - .gitignore file

After initialization, create your first event with 'gzcli event create'.`,
	Example: `  # Initialize with required flags
  gzcli init --url https://ctf.example.com --public-entry https://public.example.com

  # With discord webhook
  gzcli init --url https://ctf.example.com --public-entry https://public.example.com --discord-webhook https://discord.com/api/webhooks/...

  # After init, create your first event
  gzcli event create my-ctf-2024`,
	Run: func(cmd *cobra.Command, _ []string) {
		// Validate required flags
		if initURL == "" {
			log.Error("--url flag is required")
			_ = cmd.Usage()
			return
		}
		if initPublicEntry == "" {
			log.Error("--public-entry flag is required")
			_ = cmd.Usage()
			return
		}

		initInfo := map[string]string{
			"url":            initURL,
			"publicEntry":    initPublicEntry,
			"discordWebhook": initDiscordWebhook,
		}

		if errors := other.CTFTemplate(".", initInfo); errors != nil {
			for _, err := range errors {
				if err != nil {
					log.Error("%s", err)
				}
			}
			return
		}

		log.Info("✅ CTF project initialized successfully!")
		log.Info("\nNext steps:")
		log.Info("  1. Review server configuration: .gzctf/conf.yaml")
		log.Info("  2. Create your first event: gzcli event create <name>")
		log.Info("  3. Start the platform: make platform-up")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVar(&initURL, "url", "", "URL for the CTF instance (required)")
	initCmd.Flags().StringVar(&initPublicEntry, "public-entry", "", "Public entry point for the CTF (required)")
	initCmd.Flags().StringVar(&initDiscordWebhook, "discord-webhook", "", "Discord webhook URL for notifications (optional)")
}
