package challenge

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dimasma0305/gzcli/internal/log"
	"github.com/fsnotify/fsnotify"
)

// Manager manages challenge watch operations
type Manager struct {
	watcher    *fsnotify.Watcher
	challenges map[string]string // challengeName -> cwd
}

// NewManager creates a new challenge manager
func NewManager(watcher *fsnotify.Watcher) *Manager {
	return &Manager{
		watcher:    watcher,
		challenges: make(map[string]string),
	}
}

// AddChallenge adds a challenge directory to the watcher
func (m *Manager) AddChallenge(name, cwd string) error {
	// Check if already watching this challenge
	if _, exists := m.challenges[name]; exists {
		return nil // Already watching
	}

	// Add the challenge directory
	err := m.watcher.Add(cwd)
	if err != nil {
		return fmt.Errorf("failed to add directory %s: %w", cwd, err)
	}

	// Also watch subdirectories
	err = filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

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

// RemoveChallenge removes a challenge from the watcher
func (m *Manager) RemoveChallenge(name string) error {
	cwd, exists := m.challenges[name]
	if !exists {
		return nil
	}

	if err := m.watcher.Remove(cwd); err != nil {
		// Directory may no longer exist; log but don't fail
		log.DebugH3("Watcher remove for %s returned: %v", cwd, err)
	}

	delete(m.challenges, name)
	return nil
}

// GetChallenges returns the map of watched challenges
func (m *Manager) GetChallenges() map[string]string {
	return m.challenges
}

// FindChallengeForFile finds which challenge a file belongs to
func (m *Manager) FindChallengeForFile(filePath string) (string, string, error) {
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return "", "", err
	}

	log.DebugH3("Looking for challenge that contains file: %s", absFilePath)

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
		log.DebugH3("Best matching challenge is %s", bestMatch)
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
