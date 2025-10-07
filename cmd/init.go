package cmd

import (
	"github.com/dimasma0305/gzcli/internal/log"
	"github.com/dimasma0305/gzcli/internal/template/other"
	"github.com/spf13/cobra"
)

var (
	initUrl            string
	initPublicEntry    string
	initDiscordWebhook string
)

var initCmd = &cobra.Command{
	Use:     "init",
	Aliases: []string{"i"},
	Short:   "Initialize a new CTF project structure",
	Long: `Initialize a new CTF project with configuration files and directory structure.

This command creates:
  - .gzctf/ directory with configuration files
  - Challenge directory structure
  - Docker compose files
  - README and documentation

You can provide values via flags or be prompted for input interactively.`,
	Example: `  # Initialize with prompts
  gzcli init

  # Initialize with flags
  gzcli init --url https://ctf.example.com --public-entry https://public.example.com

  # With all options
  gzcli init --url https://ctf.example.com --public-entry https://public.example.com --discord-webhook https://discord.com/api/webhooks/...`,
	Run: func(cmd *cobra.Command, args []string) {
		initInfo := map[string]string{
			"url":            initUrl,
			"publicEntry":    initPublicEntry,
			"discordWebhook": initDiscordWebhook,
		}

		if errors := other.CTFTemplate(".", initInfo); errors != nil {
			for _, err := range errors {
				if err != nil {
					log.Error("%s", err)
				}
			}
		}

		log.Info("CTF project initialized successfully!")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVar(&initUrl, "url", "", "URL for the CTF instance")
	initCmd.Flags().StringVar(&initPublicEntry, "public-entry", "", "Public entry point for the CTF")
	initCmd.Flags().StringVar(&initDiscordWebhook, "discord-webhook", "", "Discord webhook URL for notifications")
}
