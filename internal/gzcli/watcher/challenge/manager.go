package challenge

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"

	"github.com/dimasma0305/gzcli/internal/log"
)

// Manager manages challenge watch operations with optimized path lookups
type Manager struct {
	watcher    *fsnotify.Watcher
	challenges map[string]string          // challengeName -> cwd
	pathIndex  map[string]*pathIndexEntry // path -> challenge info (for O(1) lookups)
	mu         sync.RWMutex
}

// pathIndexEntry stores challenge information for a specific path
type pathIndexEntry struct {
	challengeName string
	challengeCwd  string
	pathLength    int // Used for finding the most specific match
}

// NewManager creates a new challenge manager with path indexing
func NewManager(watcher *fsnotify.Watcher) *Manager {
	return &Manager{
		watcher:    watcher,
		challenges: make(map[string]string),
		pathIndex:  make(map[string]*pathIndexEntry, 1000), // Pre-allocate for performance
	}
}

// AddChallenge adds a challenge directory to the watcher with path indexing
func (m *Manager) AddChallenge(name, cwd string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already watching this challenge
	if _, exists := m.challenges[name]; exists {
		return nil // Already watching
	}

	// Get absolute path once
	absCwd, err := filepath.Abs(cwd)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for %s: %w", cwd, err)
	}

	// Add the challenge directory
	err = m.watcher.Add(cwd)
	if err != nil {
		return fmt.Errorf("failed to add directory %s: %w", cwd, err)
	}

	// Build path index while walking subdirectories
	err = filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil
		}

		// Index this path for fast lookups
		m.indexPath(absPath, name, absCwd)

		if info.IsDir() && !shouldIgnoreDir(path) {
			if err := m.watcher.Add(path); err != nil {
				log.Error("Failed to watch directory %s: %v", path, err)
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk directory %s: %w", cwd, err)
	}

	// Mark as watched
	m.challenges[name] = cwd
	log.InfoH2("Now watching: %s (%s)", name, cwd)
	return nil
}

// indexPath adds a path to the index for O(1) lookups
func (m *Manager) indexPath(absPath, challengeName, challengeCwd string) {
	// Normalize challenge directory path
	normCwd := challengeCwd
	if !strings.HasSuffix(normCwd, string(filepath.Separator)) {
		normCwd += string(filepath.Separator)
	}

	// Only index if path is within challenge directory
	if strings.HasPrefix(absPath, normCwd) {
		entry := &pathIndexEntry{
			challengeName: challengeName,
			challengeCwd:  challengeCwd,
			pathLength:    len(normCwd),
		}
		m.pathIndex[absPath] = entry
	}
}

// RemoveChallenge removes a challenge from the watcher and path index
func (m *Manager) RemoveChallenge(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cwd, exists := m.challenges[name]
	if !exists {
		return nil
	}

	if err := m.watcher.Remove(cwd); err != nil {
		// Directory may no longer exist; log but don't fail
		log.DebugH3("Watcher remove for %s returned: %v", cwd, err)
	}

	// Remove from path index
	m.removeFromIndex(name)

	delete(m.challenges, name)
	return nil
}

// removeFromIndex removes all paths associated with a challenge from the index
func (m *Manager) removeFromIndex(challengeName string) {
	// Remove all entries for this challenge
	for path, entry := range m.pathIndex {
		if entry.challengeName == challengeName {
			delete(m.pathIndex, path)
		}
	}
}

// GetChallenges returns the map of watched challenges
func (m *Manager) GetChallenges() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent concurrent modification
	result := make(map[string]string, len(m.challenges))
	for k, v := range m.challenges {
		result[k] = v
	}
	return result
}

// FindChallengeForFile finds which challenge a file belongs to using O(1) index lookup
func (m *Manager) FindChallengeForFile(filePath string) (string, string, error) {
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return "", "", err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	log.DebugH3("Looking for challenge that contains file: %s", absFilePath)

	// First, try direct index lookup
	if entry, found := m.pathIndex[absFilePath]; found {
		log.DebugH3("Found via index: %s", entry.challengeName)
		return entry.challengeName, entry.challengeCwd, nil
	}

	// If not found, walk up the directory tree to find the nearest match
	// This handles newly created files that aren't in the index yet
	dir := filepath.Dir(absFilePath)
	bestEntry := (*pathIndexEntry)(nil)

	for dir != "" && dir != "." && dir != "/" {
		if entry, found := m.pathIndex[dir]; found {
			// Found a parent directory in the index
			if bestEntry == nil || entry.pathLength > bestEntry.pathLength {
				bestEntry = entry
			}
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached root
		}
		dir = parent
	}

	if bestEntry != nil {
		log.DebugH3("Found via parent directory: %s", bestEntry.challengeName)
		// Add this path to index for future lookups
		m.pathIndex[absFilePath] = bestEntry
		return bestEntry.challengeName, bestEntry.challengeCwd, nil
	}

	// Fallback to linear search (for compatibility)
	return m.findChallengeLinear(absFilePath)
}

// findChallengeLinear performs a linear search as fallback (for edge cases)
//
//nolint:unparam // error return kept for interface consistency with future enhancements
func (m *Manager) findChallengeLinear(absFilePath string) (string, string, error) {
	var bestMatch string
	var bestMatchCwd string
	var longestMatch int

	for name, cwd := range m.challenges {
		absChallengeDir, err := filepath.Abs(cwd)
		if err != nil {
			log.DebugH3("Failed to get absolute path for challenge %s: %v", name, err)
			continue
		}

		// Ensure the challenge directory path ends with a separator to avoid partial matches
		if !strings.HasSuffix(absChallengeDir, string(filepath.Separator)) {
			absChallengeDir += string(filepath.Separator)
		}

		if strings.HasPrefix(absFilePath, absChallengeDir) {
			// Found a match, but check if it's more specific than previous matches
			matchLength := len(absChallengeDir)

			if matchLength > longestMatch {
				longestMatch = matchLength
				bestMatch = name
				bestMatchCwd = cwd
			}
		}
	}

	if bestMatch != "" {
		log.DebugH3("Best matching challenge (linear search) is %s", bestMatch)
		return bestMatch, bestMatchCwd, nil
	}

	return "", "", nil
}

// shouldIgnoreDir determines if a directory should be ignored
func shouldIgnoreDir(path string) bool {
	dirName := filepath.Base(path)
	if strings.HasPrefix(dirName, ".") && dirName != "." && dirName != ".." {
		return true
	}
	return false
}
