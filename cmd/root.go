/*
Copyright © 2023 dimas maulana dimasmaulana0305@gmail.com
*/

// Package cmd provides command-line interface commands for gzcli
package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/dimasma0305/gzcli/internal/log"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gzcli",
	Short: "High-performance CLI for GZ::CTF",
	Long: `gzcli - Modern command-line interface for GZ::CTF operations

A powerful tool for managing CTF challenges, teams, and automated deployments.

Features:
  • Initialize and sync CTF projects
  • Automatic file watching and redeployment
  • Team management and batch operations
  • Custom script execution
  • CTFTime scoreboard generation`,
	Example: `  # Initialize a new CTF project
  gzcli init

  # Synchronize challenges to server
  gzcli sync

  # Start file watcher
  gzcli watch start

  # Create teams from CSV
  gzcli team create teams.csv

  # Generate CTFTime scoreboard
  gzcli scoreboard > scoreboard.json`,
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		// Enable debug mode if flag is set
		if debug, _ := cmd.Flags().GetBool("debug"); debug {
			log.SetDebugMode(true)
			log.Debug("Debug mode enabled")
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Add debug flag to root command
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug logging")
}
