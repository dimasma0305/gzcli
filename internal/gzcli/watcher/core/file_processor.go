package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/dimasma0305/gzcli/internal/gzcli/config"
	"github.com/dimasma0305/gzcli/internal/gzcli/fileutil"
	"github.com/dimasma0305/gzcli/internal/log"
)

var challengeFileRegex = regexp.MustCompile(`^challenge\.(yaml|yml)$`)

// FileProcessor handles file processing operations for the watcher
type FileProcessor struct {
	eventName string
	eventPath string
}

// NewFileProcessor creates a new file processor
func NewFileProcessor(eventName, eventPath string) *FileProcessor {
	return &FileProcessor{
		eventName: eventName,
		eventPath: eventPath,
	}
}

// ProcessChallengeFile processes a challenge YAML file
func (fp *FileProcessor) ProcessChallengeFile(ctx context.Context, filePath string) (*config.ChallengeYaml, error) {
	// Check if it's a challenge file
	if !challengeFileRegex.MatchString(filepath.Base(filePath)) {
		return nil, fmt.Errorf("not a challenge file: %s", filePath)
	}

	// Load challenge configuration
	challengeConf, err := fileutil.LoadChallengeYaml(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load challenge YAML: %w", err)
	}

	// Validate challenge configuration
	if err := fp.validateChallengeConfig(challengeConf); err != nil {
		return nil, fmt.Errorf("invalid challenge configuration: %w", err)
	}

	log.InfoH3("[%s] Processing challenge file: %s", fp.eventName, filePath)
	return challengeConf, nil
}

// validateChallengeConfig validates a challenge configuration
func (fp *FileProcessor) validateChallengeConfig(conf *config.ChallengeYaml) error {
	if conf.Name == "" {
		return fmt.Errorf("challenge name is required")
	}
	if conf.Type == "" {
		return fmt.Errorf("challenge type is required")
	}
	return nil
}

// IsChallengeFile checks if a file is a challenge configuration file
func (fp *FileProcessor) IsChallengeFile(filePath string) bool {
	return challengeFileRegex.MatchString(filepath.Base(filePath))
}

// GetChallengeDirectory returns the directory containing a challenge file
func (fp *FileProcessor) GetChallengeDirectory(filePath string) string {
	return filepath.Dir(filePath)
}

// FileExists checks if a file exists
func (fp *FileProcessor) FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// GetFileModTime returns the modification time of a file
func (fp *FileProcessor) GetFileModTime(filePath string) (time.Time, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}