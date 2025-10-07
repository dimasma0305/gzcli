package filesystem

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/types"
	"github.com/dimasma0305/gzcli/internal/log"
	"github.com/fsnotify/fsnotify"
)

// EventHandler interface for handling file system events
type EventHandler interface {
	HandleFileChange(filePath string)
	HandleFileRemoval(filePath string)
	HandleChallengeRemovalByDir(removedDir string)
}

// ProcessEvent routes fsnotify events to change or removal handlers
func ProcessEvent(event fsnotify.Event, handler EventHandler) {
	// On Remove or Rename, handle potential deletion
	if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
		handler.HandleFileRemoval(event.Name)
		return
	}
	// For Create/Write, proceed with normal change handling if the file exists
	if _, err := os.Stat(event.Name); err == nil {
		handler.HandleFileChange(event.Name)
	}
}

// CheckFileRemoval determines if a file removal should trigger challenge removal
func CheckFileRemoval(path string, watchedDirs map[string]string) (bool, string, string) {
	absPath, _ := filepath.Abs(path)

	// Check if the removed path is a directory that might be a challenge directory
	for challengeName, challengeDir := range watchedDirs {
		absChallengeDir, _ := filepath.Abs(challengeDir)
		if absChallengeDir == absPath {
			return true, challengeName, absChallengeDir
		}
	}

	// If a challenge.yml or challenge.yaml is removed, infer which challenge it belonged to by path prefix
	base := filepath.Base(path)
	if base == "challenge.yml" || base == "challenge.yaml" {
		// The parent directory represents the challenge cwd
		dir := filepath.Dir(path)
		return true, "", dir
	}

	return false, "", ""
}

// IsChallengeDirectoryRemoved checks if a challenge directory was actually removed
func IsChallengeDirectoryRemoved(removedDir string) bool {
	absRemoved, _ := filepath.Abs(removedDir)

	// Check if the directory itself was removed (directory deletion scenario)
	if _, err := os.Stat(absRemoved); os.IsNotExist(err) {
		// Directory no longer exists, proceed with removal
		return true
	}

	// Directory still exists, check if challenge files are missing
	chalYml := filepath.Join(absRemoved, "challenge.yml")
	chalYaml := filepath.Join(absRemoved, "challenge.yaml")
	if os.IsNotExist(fileStat(chalYml)) && os.IsNotExist(fileStat(chalYaml)) {
		return true // Challenge files are gone
	}

	return false
}

// FindChallengeByPath finds a challenge name from watched directories by path
func FindChallengeByPath(removedDir string, watchedDirs map[string]string) string {
	absRemoved, _ := filepath.Abs(removedDir)

	for challengeName, challengeDir := range watchedDirs {
		absChallengeDir, _ := filepath.Abs(challengeDir)
		if absChallengeDir == absRemoved {
			return challengeName
		}
	}

	return ""
}

// ClearWatchedByPath clears local watched state using a removed directory path hint
func ClearWatchedByPath(removedDir string, watchedDirs map[string]string, clearCallback func(string)) {
	absRemoved, _ := filepath.Abs(removedDir)
	for name, dir := range watchedDirs {
		absDir, _ := filepath.Abs(dir)
		if strings.HasPrefix(absDir, absRemoved) || strings.HasPrefix(absRemoved, absDir) {
			clearCallback(name)
			log.InfoH3("Cleared watch state for removed challenge path: %s (%s)", name, dir)
		}
	}
}

// IsInChallengeCategory checks if a file path is within a challenge category directory
func IsInChallengeCategory(filePath string) bool {
	// Get the challenge categories
	challengeCategories := []string{
		"Misc", "Crypto", "Pwn",
		"Web", "Reverse", "Blockchain",
		"Forensics", "Hardware", "Mobile", "PPC",
		"OSINT", "Game Hacking", "AI", "Pentest",
	}

	// Normalize the file path
	normalizedPath := filepath.Clean(filePath)
	pathParts := strings.Split(normalizedPath, string(filepath.Separator))

	// Check if any part of the path matches a challenge category
	for _, part := range pathParts {
		for _, category := range challengeCategories {
			if strings.EqualFold(part, category) {
				return true
			}
		}
	}

	return false
}

// fileStat returns error if path does not exist; helper to simplify logic
func fileStat(path string) error {
	_, err := os.Stat(path)
	return err
}

// WatchLoop is the main event loop for file watching
func WatchLoop(watcher *fsnotify.Watcher, config types.WatcherConfig, handler EventHandler, ctx <-chan struct{}) {
	for {
		select {
		case <-ctx:
			return

		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if ShouldProcessEvent(event, config) {
				log.InfoH2("File change detected: %s (%s)", event.Name, event.Op.String())

				// Handle removal events immediately without debouncing
				if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
					log.InfoH2("Processing removal event immediately: %s", event.Name)
					ProcessEvent(event, handler)
				} else {
					// For other events, process immediately (debouncing will be handled at higher level if needed)
					ProcessEvent(event, handler)
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Error("Watcher error: %v", err)
		}
	}
}
