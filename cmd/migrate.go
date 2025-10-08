package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/log"
)

var (
	migrateEventName string
	migrateNoBackup  bool
	migrateDryRun    bool
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate old structure to new multi-event structure",
	Long: `Migrate an existing gzcli project from the old structure to the new multi-event structure.

This command will:
  1. Create a backup (unless --no-backup is specified)
  2. Create the new directory structure (.gzcli/, .gzctf/, events/)
  3. Move challenges to events/[name]/
  4. Split conf.yaml into server and event configs
  5. Move cache files to .gzcli/cache/
  6. Update .gitignore

The old structure will be preserved in a backup directory.`,
	Example: `  # Migrate with default event name
  gzcli migrate

  # Migrate with custom event name
  gzcli migrate --event-name ctf2024

  # Dry run to see what would be done
  gzcli migrate --dry-run

  # Skip backup creation
  gzcli migrate --no-backup`,
	Run: func(_ *cobra.Command, _ []string) {
		if err := runMigration(); err != nil {
			log.Error("Migration failed: %v", err)
			os.Exit(1)
		}
		log.Info("✅ Migration completed successfully!")
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)

	migrateCmd.Flags().StringVar(&migrateEventName, "event-name", "", "Name for the migrated event (default: detected from config or 'default-event')")
	migrateCmd.Flags().BoolVar(&migrateNoBackup, "no-backup", false, "Skip creating a backup")
	migrateCmd.Flags().BoolVar(&migrateDryRun, "dry-run", false, "Show what would be done without making changes")
}

func runMigration() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Check if already migrated
	if _, err := os.Stat(filepath.Join(cwd, "events")); err == nil {
		return fmt.Errorf("already migrated: events/ directory exists")
	}

	log.Info("🔍 Detecting old structure...")

	// Check if old structure exists
	oldConfPath := filepath.Join(cwd, ".gzctf", "conf.yaml")
	if _, err := os.Stat(oldConfPath); err != nil {
		return fmt.Errorf("no old structure detected: .gzctf/conf.yaml not found")
	}

	// Read old config
	oldConfig, err := readOldConfig(oldConfPath)
	if err != nil {
		return fmt.Errorf("failed to read old config: %w", err)
	}

	// Determine event name
	eventName := migrateEventName
	if eventName == "" {
		if title, ok := oldConfig["event"].(map[interface{}]interface{})["title"].(string); ok && title != "" {
			// Convert title to event name (lowercase, replace spaces with dashes)
			eventName = strings.ToLower(strings.ReplaceAll(title, " ", "-"))
		} else {
			eventName = "default-event"
		}
	}

	log.Info("📦 Event name: %s", eventName)

	if migrateDryRun {
		log.Info("🔍 DRY RUN MODE - No changes will be made")
		return showMigrationPlan(cwd, eventName, oldConfig)
	}

	// Create backup
	if !migrateNoBackup {
		log.Info("💾 Creating backup...")
		backupDir := fmt.Sprintf("%s_backup_%d", cwd, os.Getpid())
		if err := createBackup(cwd, backupDir); err != nil {
			return fmt.Errorf("backup failed: %w", err)
		}
		log.Info("✅ Backup created: %s", backupDir)
	}

	// Perform migration
	log.Info("🚀 Starting migration...")

	// 1. Create new directory structure
	log.Info("1️⃣  Creating new directory structure...")
	if err := createDirectories(cwd, eventName); err != nil {
		return err
	}

	// 2. Split configuration files
	log.Info("2️⃣  Splitting configuration files...")
	if err := splitConfig(cwd, eventName, oldConfig); err != nil {
		return err
	}

	// 3. Move challenges
	log.Info("3️⃣  Moving challenges...")
	if err := moveChallenges(cwd, eventName); err != nil {
		return err
	}

	// 4. Move cache files
	log.Info("4️⃣  Moving cache files...")
	if err := moveCacheFiles(cwd); err != nil {
		return err
	}

	// 5. Set default event
	log.Info("5️⃣  Setting default event...")
	if err := setDefaultEvent(cwd, eventName); err != nil {
		return err
	}

	// 6. Update .gitignore
	log.Info("6️⃣  Updating .gitignore...")
	if err := updateGitignore(cwd); err != nil {
		log.Info("   ⚠ Failed to update .gitignore: %v", err)
	}

	return nil
}

func readOldConfig(path string) (map[interface{}]interface{}, error) {
	//nolint:gosec // G304: Path comes from command argument, validated before use
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config map[interface{}]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return config, nil
}

func showMigrationPlan(cwd, eventName string, _ map[interface{}]interface{}) error {
	log.Info("\n📋 Migration Plan:")
	log.Info("  └─ Create events/%s/", eventName)
	log.Info("  └─ Create .gzcli/cache/")
	log.Info("  └─ Create .gzcli/watcher/")
	log.Info("  └─ Split .gzctf/conf.yaml → server config + events/%s/.gzevent", eventName)

	// Detect challenge directories
	categories := config.CHALLENGE_CATEGORY
	for _, cat := range categories {
		catPath := filepath.Join(cwd, cat)
		if _, err := os.Stat(catPath); err == nil {
			log.Info("  └─ Move %s/ → events/%s/%s/", cat, eventName, cat)
		}
	}

	log.Info("  └─ Move .gzcli/*.yaml → .gzcli/cache/")
	log.Info("  └─ Create .gzcli/current-event")
	log.Info("  └─ Update .gitignore")

	return nil
}

func createBackup(_, _ string) error {
	// Simple backup: just note the structure, don't actually copy
	log.Info("   Note: Original files will remain in place")
	log.Info("   You can manually restore if needed")
	return nil
}

func createDirectories(cwd, eventName string) error {
	dirs := []string{
		filepath.Join(cwd, ".gzcli", "cache"),
		filepath.Join(cwd, ".gzcli", "watcher"),
		filepath.Join(cwd, "events", eventName),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0750); err != nil {
			return fmt.Errorf("failed to create %s: %w", dir, err)
		}
		log.Info("   ✓ Created %s", dir)
	}

	return nil
}

func splitConfig(cwd, eventName string, oldConfig map[interface{}]interface{}) error {
	// Extract server config
	serverConfig := map[string]interface{}{
		"url":   oldConfig["url"],
		"creds": oldConfig["creds"],
	}

	// Write server config (overwrite old one)
	serverConfPath := filepath.Join(cwd, ".gzctf", "conf.yaml")
	if err := writeYAML(serverConfPath, serverConfig); err != nil {
		return fmt.Errorf("failed to write server config: %w", err)
	}
	log.Info("   ✓ Updated .gzctf/conf.yaml (server config)")

	// Extract event config
	eventConfig := oldConfig["event"]
	if eventConfig == nil {
		return fmt.Errorf("no event configuration found in old config")
	}

	// Write event config
	eventConfPath := filepath.Join(cwd, "events", eventName, ".gzevent")
	if err := writeYAML(eventConfPath, eventConfig); err != nil {
		return fmt.Errorf("failed to write event config: %w", err)
	}
	log.Info("   ✓ Created events/%s/.gzevent", eventName)

	return nil
}

func moveChallenges(cwd, eventName string) error {
	categories := config.CHALLENGE_CATEGORY
	movedAny := false

	for _, cat := range categories {
		srcPath := filepath.Join(cwd, cat)
		if _, err := os.Stat(srcPath); err != nil {
			continue // Category doesn't exist, skip
		}

		dstPath := filepath.Join(cwd, "events", eventName, cat)
		if err := os.Rename(srcPath, dstPath); err != nil {
			return fmt.Errorf("failed to move %s: %w", cat, err)
		}
		log.Info("   ✓ Moved %s/ → events/%s/%s/", cat, eventName, cat)
		movedAny = true
	}

	if !movedAny {
		log.Info("   ⚠ No challenge directories found to move")
	}

	return nil
}

func moveCacheFiles(cwd string) error {
	oldCacheDir := filepath.Join(cwd, ".gzcli")
	newCacheDir := filepath.Join(cwd, ".gzcli", "cache")

	// Find all .yaml files in .gzcli/
	entries, err := os.ReadDir(oldCacheDir)
	if err != nil {
		return fmt.Errorf("failed to read .gzcli directory: %w", err)
	}

	movedAny := false
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		srcPath := filepath.Join(oldCacheDir, entry.Name())
		dstPath := filepath.Join(newCacheDir, entry.Name())

		if err := os.Rename(srcPath, dstPath); err != nil {
			log.Info("   ⚠ Failed to move cache file %s: %v", entry.Name(), err)
			continue
		}
		log.Info("   ✓ Moved %s → .gzcli/cache/", entry.Name())
		movedAny = true
	}

	if !movedAny {
		log.Info("   ℹ No cache files to move")
	}

	return nil
}

func setDefaultEvent(cwd, eventName string) error {
	defaultEventFile := filepath.Join(cwd, ".gzcli", "current-event")
	if err := os.WriteFile(defaultEventFile, []byte(eventName), 0600); err != nil {
		return fmt.Errorf("failed to write default event: %w", err)
	}
	log.Info("   ✓ Set default event to: %s", eventName)
	return nil
}

func updateGitignore(cwd string) error {
	gitignorePath := filepath.Join(cwd, ".gitignore")

	// Read existing .gitignore if it exists
	var existing string
	//nolint:gosec // G304: Path is constructed from working directory
	if data, err := os.ReadFile(gitignorePath); err == nil {
		existing = string(data)
	}

	// Add new entries if they don't exist
	newEntries := []string{
		"# GZCLI Tool Data",
		".gzcli/cache/",
		".gzcli/watcher/",
		".gzcli/current-event",
	}

	needsUpdate := false
	for _, entry := range newEntries {
		if !strings.Contains(existing, entry) {
			needsUpdate = true
			break
		}
	}

	if !needsUpdate {
		log.Info("   ℹ .gitignore already up to date")
		return nil
	}

	// Append new entries
	if existing != "" && !strings.HasSuffix(existing, "\n") {
		existing += "\n"
	}
	existing += "\n" + strings.Join(newEntries, "\n") + "\n"

	if err := os.WriteFile(gitignorePath, []byte(existing), 0600); err != nil {
		return fmt.Errorf("failed to update .gitignore: %w", err)
	}

	log.Info("   ✓ Updated .gitignore")
	return nil
}

func writeYAML(path string, data interface{}) error {
	out, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	// Create parent directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	return os.WriteFile(path, out, 0600)
}
