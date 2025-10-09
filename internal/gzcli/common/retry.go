package common

import (
	"context"
	"fmt"
	"time"
)

// RetryHandler handles retry logic with exponential backoff
type RetryHandler struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

// NewRetryHandler creates a new retry handler
func NewRetryHandler(maxRetries int, baseDelay, maxDelay time.Duration) *RetryHandler {
	return &RetryHandler{
		MaxRetries: maxRetries,
		BaseDelay:  baseDelay,
		MaxDelay:   maxDelay,
	}
}

// Execute executes a function with retry logic
func (r *RetryHandler) Execute(ctx context.Context, fn func() error) error {
	var lastErr error
	
	for attempt := 0; attempt <= r.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := r.calculateDelay(attempt)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
		
		if err := fn(); err != nil {
			lastErr = err
			continue
		}
		
		return nil
	}
	
	return fmt.Errorf("operation failed after %d attempts: %w", r.MaxRetries+1, lastErr)
}

// calculateDelay calculates the delay for the given attempt
func (r *RetryHandler) calculateDelay(attempt int) time.Duration {
	delay := r.BaseDelay * time.Duration(1<<uint(attempt-1))
	if delay > r.MaxDelay {
		delay = r.MaxDelay
	}
	return delay
}