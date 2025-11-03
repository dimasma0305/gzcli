package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
)

// validEventNames returns a list of valid event names for shell completion.
// It is used by cobra to provide suggestions for commands that take an event name as an argument.
func validEventNames(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	eventNames, err := getAvailableEvents()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return eventNames, cobra.ShellCompDirectiveNoFileComp
}

// getAvailableEvents scans the events directory and returns a list of available event names.
// An event is considered available if it has a .gzevent file in its directory.
func getAvailableEvents() ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	eventsDir := filepath.Join(cwd, config.EVENTS_DIR)

	// Check if events directory exists
	if _, err := os.Stat(eventsDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	entries, err := os.ReadDir(eventsDir)
	if err != nil {
		return nil, err
	}

	var eventNames []string
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if .gzevent file exists
			gzeventPath := filepath.Join(eventsDir, entry.Name(), config.GZEVENT_FILE)
			if _, err := os.Stat(gzeventPath); err == nil {
				eventNames = append(eventNames, entry.Name())
			}
		}
	}

	return eventNames, nil
}

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion script for gzcli.

To load completions:

Bash:

  $ source <(gzcli completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ gzcli completion bash > /etc/bash_completion.d/gzcli
  # macOS:
  $ gzcli completion bash > $(brew --prefix)/etc/bash_completion.d/gzcli

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ gzcli completion zsh > "${fpath[1]}/_gzcli"

  # You will need to start a new shell for this setup to take effect.

Fish:

  $ gzcli completion fish | source

  # To load completions for each session, execute once:
  $ gzcli completion fish > ~/.config/fish/completions/gzcli.fish

PowerShell:

  PS> gzcli completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> gzcli completion powershell > gzcli.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		switch args[0] {
		case "bash":
			err = cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			err = cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			err = cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			err = cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
		if err != nil {
			// Error is logged but not fatal for completion generation
			cmd.PrintErrf("Error generating completion: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
