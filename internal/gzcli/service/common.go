package service

import (
	"fmt"
	"time"
)

// RetryHandler handles retry logic for operations
type RetryHandler struct {
	MaxRetries int
	Delay      time.Duration
}

// NewRetryHandler creates a new retry handler
func NewRetryHandler(maxRetries int, delay time.Duration) *RetryHandler {
	return &RetryHandler{
		MaxRetries: maxRetries,
		Delay:      delay,
	}
}

// Execute executes a function with retry logic
func (r *RetryHandler) Execute(fn func() error) error {
	var lastErr error
	
	for i := 0; i <= r.MaxRetries; i++ {
		if err := fn(); err != nil {
			lastErr = err
			if i < r.MaxRetries {
				time.Sleep(r.Delay)
				continue
			}
			return fmt.Errorf("operation failed after %d retries: %w", r.MaxRetries, lastErr)
		}
		return nil
	}
	
	return lastErr
}

// ErrorHandler handles error formatting and context
type ErrorHandler struct{}

// NewErrorHandler creates a new error handler
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

// Wrap wraps an error with context
func (e *ErrorHandler) Wrap(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// Validator handles validation logic
type Validator struct{}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateNotEmpty validates that a string is not empty
func (v *Validator) ValidateNotEmpty(value, fieldName string) error {
	if value == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}
	return nil
}

// ValidatePositive validates that a number is positive
func (v *Validator) ValidatePositive(value int, fieldName string) error {
	if value <= 0 {
		return fmt.Errorf("%s must be positive, got %d", fieldName, value)
	}
	return nil
}