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

// Version information variables.
// These can be overridden at build time using ldflags:
//
//	go build -ldflags "-X github.com/dimasma0305/gzcli/cmd.Version=x.y.z \
//	  -X github.com/dimasma0305/gzcli/cmd.GitCommit=abc123 \
//	  -X github.com/dimasma0305/gzcli/cmd.BuildTime=2025-10-07_12:34:56"
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gzcli",
	Short: "High-performance CLI for GZ::CTF",
	Version: func() string {
		if GitCommit != "unknown" && BuildTime != "unknown" {
			return Version + "\nCommit: " + GitCommit + "\nBuilt: " + BuildTime
		}
		return Version
	}(),
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

var (
	// Global event flag - shared across all commands
	globalEventFlag string
)

func init() {
	// Add debug flag to root command
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug logging")

	// Add global event selection flag
	rootCmd.PersistentFlags().StringVarP(&globalEventFlag, "event", "e", "", "Specify which event to use (overrides GZCLI_EVENT env var)")

	// Register completion for global --event flag
	_ = rootCmd.RegisterFlagCompletionFunc("event", validEventNames)
}

// GetEventFlag returns the value of the global --event flag.
// This function provides a clean way for other commands to access the globally specified event.
func GetEventFlag() string {
	return globalEventFlag
}
