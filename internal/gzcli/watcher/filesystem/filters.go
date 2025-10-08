package filesystem

import (
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"

	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/watchertypes"
	"github.com/dimasma0305/gzcli/internal/log"
)

// FilterMatcher provides optimized pattern matching with pre-compiled regex
type FilterMatcher struct {
	ignoreRegexes []*regexp.Regexp
	watchRegexes  []*regexp.Regexp
	mu            sync.RWMutex
}

// globalFilterMatcher is a singleton for compiled patterns
var (
	globalFilterMatcher *FilterMatcher
	filterOnce          sync.Once
)

// getFilterMatcher returns the global filter matcher, initializing if needed
func getFilterMatcher(config watchertypes.WatcherConfig) *FilterMatcher {
	filterOnce.Do(func() {
		globalFilterMatcher = newFilterMatcher(config)
	})
	return globalFilterMatcher
}

// newFilterMatcher creates a new filter matcher with pre-compiled patterns
func newFilterMatcher(config watchertypes.WatcherConfig) *FilterMatcher {
	fm := &FilterMatcher{
		ignoreRegexes: make([]*regexp.Regexp, 0, len(config.IgnorePatterns)),
		watchRegexes:  make([]*regexp.Regexp, 0, len(config.WatchPatterns)),
	}

	// Compile ignore patterns to regex
	for _, pattern := range config.IgnorePatterns {
		// Convert glob pattern to regex
		regex := globToRegex(pattern)
		if compiled, err := regexp.Compile(regex); err == nil {
			fm.ignoreRegexes = append(fm.ignoreRegexes, compiled)
		}
	}

	// Compile watch patterns to regex
	for _, pattern := range config.WatchPatterns {
		regex := globToRegex(pattern)
		if compiled, err := regexp.Compile(regex); err == nil {
			fm.watchRegexes = append(fm.watchRegexes, compiled)
		}
	}

	return fm
}

// globToRegex converts a glob pattern to a regular expression
func globToRegex(glob string) string {
	// Escape special regex characters except * and ?
	regex := regexp.QuoteMeta(glob)
	// Replace \* with .* (match any characters)
	regex = strings.ReplaceAll(regex, `\*`, ".*")
	// Replace \? with . (match single character)
	regex = strings.ReplaceAll(regex, `\?`, ".")
	// Anchor the pattern
	return "^" + regex + "$"
}

// matchesIgnorePattern checks if a filename/path matches any ignore pattern
func (fm *FilterMatcher) matchesIgnorePattern(filename, fullPath string) bool {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	for _, regex := range fm.ignoreRegexes {
		if regex.MatchString(filename) || regex.MatchString(fullPath) {
			return true
		}
	}
	return false
}

// matchesWatchPattern checks if a filename matches any watch pattern
func (fm *FilterMatcher) matchesWatchPattern(filename string) bool {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	if len(fm.watchRegexes) == 0 {
		return true // No watch patterns means watch everything
	}

	for _, regex := range fm.watchRegexes {
		if regex.MatchString(filename) {
			return true
		}
	}
	return false
}

// ShouldProcessEvent determines if we should process a file system event using optimized matching
func ShouldProcessEvent(event fsnotify.Event, config watchertypes.WatcherConfig) bool {
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

	// Use optimized regex-based matching if patterns exist
	if len(config.IgnorePatterns) > 0 || len(config.WatchPatterns) > 0 {
		matcher := getFilterMatcher(config)

		// Check ignore patterns first
		if matcher.matchesIgnorePattern(filename, event.Name) {
			return false
		}

		// Check watch patterns
		if len(config.WatchPatterns) > 0 {
			return matcher.matchesWatchPattern(filename)
		}
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
func DetermineUpdateType(filePath string, challengeCwd string) watchertypes.UpdateType {
	// Get relative path from challenge directory
	absChallengePath, err := filepath.Abs(challengeCwd)
	if err != nil {
		log.Error("Failed to get absolute challenge path: %v", err)
		return watchertypes.UpdateFullRedeploy // Default to full redeploy on error
	}

	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		log.Error("Failed to get absolute file path: %v", err)
		return watchertypes.UpdateFullRedeploy // Default to full redeploy on error
	}

	relPath, err := filepath.Rel(absChallengePath, absFilePath)
	if err != nil {
		log.Error("Failed to get relative path: %v", err)
		return watchertypes.UpdateFullRedeploy // Default to full redeploy on error
	}

	// Normalize path separators for consistent matching on Windows and Unix
	// Convert backslashes to forward slashes
	relPath = filepath.ToSlash(relPath)

	// Check if it's in solver directory - no update needed
	if strings.HasPrefix(relPath, "solver/") || strings.HasPrefix(relPath, "writeup/") {
		log.InfoH3("File is in solver/writeup directory, skipping update")
		return watchertypes.UpdateNone
	}

	// Check if it's challenge.yml or challenge.yaml - metadata update only
	base := filepath.Base(relPath)
	if base == "challenge.yml" || base == "challenge.yaml" {
		log.InfoH3("Challenge configuration file changed, updating metadata and attachment")
		return watchertypes.UpdateMetadata
	}

	// Check if it's in dist directory - attachment update only
	if strings.HasPrefix(relPath, "dist/") {
		log.InfoH3("File in dist directory changed, updating attachment only")
		return watchertypes.UpdateAttachment
	}

	// Check if it's in src directory - full redeploy needed
	if strings.HasPrefix(relPath, "src/") {
		log.InfoH3("Source file changed, full redeploy needed")
		return watchertypes.UpdateFullRedeploy
	}

	// Check for other important files that need full redeploy
	fileName := filepath.Base(relPath)
	if fileName == "Dockerfile" || fileName == "docker-compose.yml" || fileName == "Makefile" {
		log.InfoH3("Infrastructure file changed (%s), full redeploy needed", fileName)
		return watchertypes.UpdateFullRedeploy
	}

	// Only listen to src/, dist/ and challenge.yml/yaml. Ignore any other paths.
	log.InfoH3("Change outside allowed paths (src/, dist/, challenge.yml/.yaml); ignoring")
	return watchertypes.UpdateNone
}
