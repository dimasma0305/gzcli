package errors

import (
	"errors"
	"fmt"
)

// Sentinel errors for common error conditions
var (
	// Authentication errors
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmptyURL          = errors.New("URL cannot be empty")
	ErrLoginFailed       = errors.New("login failed")

	// Configuration errors
	ErrConfigNotFound    = errors.New("configuration not found")
	ErrInvalidConfig     = errors.New("invalid configuration")
	ErrMissingRequired   = errors.New("missing required field")

	// Challenge errors
	ErrChallengeNotFound    = errors.New("challenge not found")
	ErrValidationFailed     = errors.New("validation failed")
	ErrInvalidChallengeType = errors.New("invalid challenge type")
	ErrMissingFlags         = errors.New("missing flags for static challenge")
	ErrMissingFlagTemplate  = errors.New("missing flag template for dynamic container")

	// API errors
	ErrAPIConnection = errors.New("API connection failed")
	ErrAPIResponse   = errors.New("invalid API response")
	ErrRateLimited   = errors.New("rate limited")

	// File system errors
	ErrFileNotFound = errors.New("file not found")
	ErrInvalidPath  = errors.New("invalid path")
	ErrPermissionDenied = errors.New("permission denied")

	// Watcher errors
	ErrWatcherNotRunning = errors.New("watcher not running")
	ErrWatcherFailed     = errors.New("watcher failed")
)

// Wrap wraps an error with additional context
func Wrap(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// Wrapf wraps an error with formatted context
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}

// Is checks if the error is of a specific type
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As checks if the error can be unwrapped to the target type
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}