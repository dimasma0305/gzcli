package filesystem

import (
	"path/filepath"
	"strings"

	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/types"
	"github.com/dimasma0305/gzcli/internal/log"
	"github.com/fsnotify/fsnotify"
)

// ShouldProcessEvent determines if we should process a file system event
func ShouldProcessEvent(event fsnotify.Event, config types.WatcherConfig) bool {
	// Process Write, Create, Remove, and Rename events
	if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
		return false
	}

	filename := filepath.Base(event.Name)

	// Skip common editor temporary files that cause loops
	if strings.HasPrefix(filename, ".") && (strings.HasSuffix(filename, ".swp") ||
		strings.HasSuffix(filename, ".tmp") ||
		strings.HasSuffix(filename, "~") ||
		strings.Contains(filename, ".sw")) {
		return false
	}

	// Skip VSCode temporary files
	if strings.HasPrefix(filename, ".vscode") || strings.Contains(event.Name, ".vscode") {
		return false
	}

	// Check ignore patterns (if any)
	for _, pattern := range config.IgnorePatterns {
		if matched, _ := filepath.Match(pattern, filename); matched {
			return false
		}
		if strings.Contains(event.Name, pattern) {
			return false
		}
	}

	// Check if it matches watch patterns (if specified)
	if len(config.WatchPatterns) > 0 {
		for _, pattern := range config.WatchPatterns {
			if matched, _ := filepath.Match(pattern, filename); matched {
				return true
			}
		}
		return false
	}

	return true
}

// ShouldIgnoreDir determines if a directory should be ignored
func ShouldIgnoreDir(path string) bool {
	// Only ignore hidden directories that start with dot (except current dir)
	dirName := filepath.Base(path)
	if strings.HasPrefix(dirName, ".") && dirName != "." && dirName != ".." {
		return true
	}
	return false
}

// DetermineUpdateType determines what type of update is needed based on the changed file
func DetermineUpdateType(filePath string, challengeCwd string) types.UpdateType {
	// Get relative path from challenge directory
	absChallengePath, err := filepath.Abs(challengeCwd)
	if err != nil {
		log.Error("Failed to get absolute challenge path: %v", err)
		return types.UpdateFullRedeploy // Default to full redeploy on error
	}

	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		log.Error("Failed to get absolute file path: %v", err)
		return types.UpdateFullRedeploy // Default to full redeploy on error
	}

	relPath, err := filepath.Rel(absChallengePath, absFilePath)
	if err != nil {
		log.Error("Failed to get relative path: %v", err)
		return types.UpdateFullRedeploy // Default to full redeploy on error
	}

	// Check if it's in solver directory - no update needed
	if strings.HasPrefix(relPath, "solver/") || strings.HasPrefix(relPath, "writeup/") {
		log.InfoH3("File is in solver/writeup directory, skipping update")
		return types.UpdateNone
	}

	// Check if it's challenge.yml or challenge.yaml - metadata update only
	base := filepath.Base(relPath)
	if base == "challenge.yml" || base == "challenge.yaml" {
		log.InfoH3("Challenge configuration file changed, updating metadata and attachment")
		return types.UpdateMetadata
	}

	// Check if it's in dist directory - attachment update only
	if strings.HasPrefix(relPath, "dist/") {
		log.InfoH3("File in dist directory changed, updating attachment only")
		return types.UpdateAttachment
	}

	// Check if it's in src directory - full redeploy needed
	if strings.HasPrefix(relPath, "src/") {
		log.InfoH3("Source file changed, full redeploy needed")
		return types.UpdateFullRedeploy
	}

	// Check for other important files that need full redeploy
	fileName := filepath.Base(relPath)
	if fileName == "Dockerfile" || fileName == "docker-compose.yml" || fileName == "Makefile" {
		log.InfoH3("Infrastructure file changed (%s), full redeploy needed", fileName)
		return types.UpdateFullRedeploy
	}

	// Only listen to src/, dist/ and challenge.yml/yaml. Ignore any other paths.
	log.InfoH3("Change outside allowed paths (src/, dist/, challenge.yml/.yaml); ignoring")
	return types.UpdateNone
}
