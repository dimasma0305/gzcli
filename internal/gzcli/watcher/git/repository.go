package git

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindGitRepoRoot walks up from startPath to find a directory containing a .git folder
func FindGitRepoRoot(startPath string) (string, error) {
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path for %s: %w", startPath, err)
	}

	current := absPath
	for {
		gitDir := filepath.Join(current, ".git")
		if info, statErr := os.Stat(gitDir); statErr == nil && info.IsDir() {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("no .git directory found from %s up to filesystem root", absPath)
		}
		current = parent
	}
}
