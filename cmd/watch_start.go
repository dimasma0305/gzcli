package cmd

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/dimasma0305/gzcli/internal/gzcli"
	"github.com/dimasma0305/gzcli/internal/log"
)

var (
	watchForeground    bool
	watchPidFile       string
	watchLogFile       string
	watchDebounce      time.Duration
	watchPollInterval  time.Duration
	watchIgnore        []string
	watchPatterns      []string
	watchGitPull       bool
	watchGitInterval   time.Duration
	watchGitRepo       string
	watchEvents        []string // Multiple events to watch
	watchExcludeEvents []string // Events to exclude from watching
)

var watchStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the file watcher daemon",
	Long: `Start the file watcher daemon for automatic challenge redeployment.

By default, watches all events. Use --event to specify specific events,
or --exclude-event to exclude certain events.

The watcher runs as a daemon by default. Use --foreground to run in the current terminal.`,
	Example: `  # Start as daemon for all events
  gzcli watch start

  # Start in foreground
  gzcli watch start --foreground

  # Watch specific events only
  gzcli watch start --event ctf2024 --event ctf2025

  # Watch all except specific events
  gzcli watch start --exclude-event practice

  # Start with custom debounce time
  gzcli watch start --debounce 5s

  # Start with custom ignore patterns
  gzcli watch start --ignore "*.tmp" --ignore "*.log"`,
	Run: func(_ *cobra.Command, _ []string) {
		// Determine which events to watch
		eventsToWatch, err := ResolveTargetEvents(watchEvents, watchExcludeEvents)
		if err != nil {
			log.Error("Failed to resolve target events: %v", err)
			os.Exit(1)
		}

		log.InfoH2("Watching %d event(s): %v", len(eventsToWatch), eventsToWatch)

		// Initialize GZAPI without event context (we'll handle events in the watcher)
		gz, err := gzcli.InitWithEvent("")
		if err != nil {
			log.Error("Failed to initialize: %v", err)
			os.Exit(1)
		}

		config := gzcli.WatcherConfig{
			Events:                    eventsToWatch,
			PollInterval:              watchPollInterval,
			DebounceTime:              watchDebounce,
			IgnorePatterns:            gzcli.DefaultWatcherConfig.IgnorePatterns,
			WatchPatterns:             gzcli.DefaultWatcherConfig.WatchPatterns,
			NewChallengeCheckInterval: gzcli.DefaultWatcherConfig.NewChallengeCheckInterval,
			DaemonMode:                !watchForeground,
			PidFile:                   gzcli.DefaultWatcherConfig.PidFile,
			LogFile:                   gzcli.DefaultWatcherConfig.LogFile,
			GitPullEnabled:            watchGitPull,
			GitPullInterval:           watchGitInterval,
			GitRepository:             watchGitRepo,
			DatabaseEnabled:           true,
			SocketEnabled:             true,
		}

		if watchPidFile != "" {
			config.PidFile = watchPidFile
		}
		if watchLogFile != "" {
			config.LogFile = watchLogFile
		}
		if len(watchIgnore) > 0 {
			config.IgnorePatterns = append(config.IgnorePatterns, watchIgnore...)
		}
		if len(watchPatterns) > 0 {
			config.WatchPatterns = watchPatterns
		}

		if config.DaemonMode {
			log.Info("Starting file watcher as daemon...")
		} else {
			log.Info("Starting file watcher in foreground...")
		}

		if err := gz.StartWatcher(config); err != nil {
			log.Fatal("Failed to start watcher: ", err)
		}

		if !config.DaemonMode {
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			log.Info("File watcher running in foreground. Press Ctrl+C to stop.")
			<-sigChan

			log.Info("Shutting down file watcher...")
			if err := gz.StopWatcher(); err != nil {
				log.Error("Error stopping watcher: %v", err)
			}
			log.Info("File watcher stopped.")
		}
	},
}

func init() {
	watchCmd.AddCommand(watchStartCmd)

	watchStartCmd.Flags().StringSliceVarP(&watchEvents, "event", "e", []string{}, "Specific event(s) to watch (can be specified multiple times)")
	watchStartCmd.Flags().StringSliceVar(&watchExcludeEvents, "exclude-event", []string{}, "Event(s) to exclude from watching (can be specified multiple times)")
	watchStartCmd.Flags().BoolVarP(&watchForeground, "foreground", "f", false, "Run in foreground instead of daemon mode")
	watchStartCmd.Flags().StringVar(&watchPidFile, "pid-file", "", "Custom PID file location (default: /tmp/gzctf-watcher.pid)")
	watchStartCmd.Flags().StringVar(&watchLogFile, "log-file", "", "Custom log file location (default: /tmp/gzctf-watcher.log)")
	watchStartCmd.Flags().DurationVar(&watchDebounce, "debounce", 2*time.Second, "Debounce time for file changes")
	watchStartCmd.Flags().DurationVar(&watchPollInterval, "poll-interval", 5*time.Second, "Polling interval for file changes")
	watchStartCmd.Flags().StringSliceVar(&watchIgnore, "ignore", []string{}, "Additional patterns to ignore")
	watchStartCmd.Flags().StringSliceVar(&watchPatterns, "patterns", []string{}, "File patterns to watch (overrides default)")
	watchStartCmd.Flags().BoolVar(&watchGitPull, "git-pull", true, "Enable automatic git pull")
	watchStartCmd.Flags().DurationVar(&watchGitInterval, "git-interval", 1*time.Minute, "Git pull interval")
	watchStartCmd.Flags().StringVar(&watchGitRepo, "git-repo", ".", "Git repository path")

	// Register completion for --event flag
	_ = watchStartCmd.RegisterFlagCompletionFunc("event", validEventNames)
}
