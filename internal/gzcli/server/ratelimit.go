package server

import (
	"sync"
	"time"
)

// RateLimiter implements token bucket rate limiting per IP
type RateLimiter struct {
	buckets map[string]*TokenBucket
	mu      sync.RWMutex
}

// TokenBucket represents a token bucket for rate limiting
type TokenBucket struct {
	tokens     int
	maxTokens  int
	refillRate time.Duration
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		buckets: make(map[string]*TokenBucket),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// AllowAction checks if an action is allowed for an IP
func (rl *RateLimiter) AllowAction(ip, actionType string) (bool, time.Duration) {
	var maxTokens int
	var refillRate time.Duration

	// Define rate limits per action type
	switch actionType {
	case "start", "stop", "restart":
		maxTokens = 5
		refillRate = time.Minute / 5 // 5 actions per minute
	case "vote":
		maxTokens = 10
		refillRate = time.Minute / 10 // 10 votes per minute
	case "websocket":
		maxTokens = 20
		refillRate = time.Second * 2 // 20 connections per 40 seconds (1 every 2 seconds)
	default:
		maxTokens = 10
		refillRate = time.Minute / 10
	}

	rl.mu.Lock()
	bucket, exists := rl.buckets[ip+":"+actionType]
	if !exists {
		bucket = &TokenBucket{
			tokens:     maxTokens,
			maxTokens:  maxTokens,
			refillRate: refillRate,
			lastRefill: time.Now(),
		}
		rl.buckets[ip+":"+actionType] = bucket
	}
	rl.mu.Unlock()

	return bucket.Take()
}

// Take attempts to take a token from the bucket
func (tb *TokenBucket) Take() (bool, time.Duration) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	tokensToAdd := int(elapsed / tb.refillRate)

	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.maxTokens {
			tb.tokens = tb.maxTokens
		}
		tb.lastRefill = now
	}

	// Try to take a token
	if tb.tokens > 0 {
		tb.tokens--
		return true, 0
	}

	// Calculate wait time until next token
	waitTime := tb.refillRate - elapsed%tb.refillRate
	return false, waitTime
}

// cleanup removes old buckets to prevent memory leaks
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()

		for key, bucket := range rl.buckets {
			bucket.mu.Lock()
			// Remove buckets that haven't been used in 10 minutes
			if now.Sub(bucket.lastRefill) > 10*time.Minute {
				delete(rl.buckets, key)
			}
			bucket.mu.Unlock()
		}

		rl.mu.Unlock()
	}
}
