package common

import (
	"context"
	"fmt"
	"time"

	"github.com/dimasma0305/gzcli/internal/gzcli/errors"
	"github.com/dimasma0305/gzcli/internal/log"
)

// RetryConfig holds configuration for retry operations
type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   time.Second,
		MaxDelay:    30 * time.Second,
		Multiplier:  2.0,
	}
}

// RetryHandler handles retry logic for operations
type RetryHandler struct {
	config RetryConfig
}

// NewRetryHandler creates a new RetryHandler
func NewRetryHandler(config RetryConfig) *RetryHandler {
	return &RetryHandler{
		config: config,
	}
}

// Execute executes a function with retry logic
func (r *RetryHandler) Execute(ctx context.Context, operation func() error) error {
	var lastErr error
	
	for attempt := 1; attempt <= r.config.MaxAttempts; attempt++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "operation cancelled")
		default:
		}

		// Execute the operation
		err := operation()
		if err == nil {
			if attempt > 1 {
				log.Info("Operation succeeded on attempt %d", attempt)
			}
			return nil
		}

		lastErr = err
		
		// Don't retry on the last attempt
		if attempt == r.config.MaxAttempts {
			break
		}

		// Calculate delay with exponential backoff
		delay := r.calculateDelay(attempt)
		log.Warn("Operation failed on attempt %d, retrying in %v: %v", attempt, delay, err)

		// Wait for the delay or context cancellation
		select {
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "operation cancelled during retry")
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return errors.Wrapf(lastErr, "operation failed after %d attempts", r.config.MaxAttempts)
}

// calculateDelay calculates the delay for the given attempt
func (r *RetryHandler) calculateDelay(attempt int) time.Duration {
	delay := time.Duration(float64(r.config.BaseDelay) * pow(r.config.Multiplier, float64(attempt-1)))
	
	if delay > r.config.MaxDelay {
		delay = r.config.MaxDelay
	}
	
	return delay
}

// pow calculates x^y
func pow(x, y float64) float64 {
	result := 1.0
	for i := 0; i < int(y); i++ {
		result *= x
	}
	return result
}