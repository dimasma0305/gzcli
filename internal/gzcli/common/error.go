package common

import (
	"fmt"
	"strings"
)

// ErrorHandler provides common error handling utilities
type ErrorHandler struct{}

// NewErrorHandler creates a new error handler
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

// WrapError wraps an error with additional context
func (e *ErrorHandler) WrapError(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// WrapErrorf wraps an error with formatted context
func (e *ErrorHandler) WrapErrorf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}

// CollectErrors collects multiple errors into a single error
func (e *ErrorHandler) CollectErrors(errors []error) error {
	if len(errors) == 0 {
		return nil
	}
	
	var errorStrings []string
	for _, err := range errors {
		if err != nil {
			errorStrings = append(errorStrings, err.Error())
		}
	}
	
	if len(errorStrings) == 0 {
		return nil
	}
	
	return fmt.Errorf("multiple errors occurred: %s", strings.Join(errorStrings, "; "))
}

// IsRetryableError determines if an error is retryable
func (e *ErrorHandler) IsRetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	// Add logic to determine if an error is retryable
	// This could check for network errors, temporary failures, etc.
	errorStr := err.Error()
	
	// Common retryable error patterns
	retryablePatterns := []string{
		"timeout",
		"connection refused",
		"temporary failure",
		"rate limit",
		"server error",
		"network error",
	}
	
	for _, pattern := range retryablePatterns {
		if strings.Contains(strings.ToLower(errorStr), pattern) {
			return true
		}
	}
	
	return false
}