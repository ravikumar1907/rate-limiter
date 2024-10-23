package limiter

import (
	"time"
)

// RateLimiter manages the rate limiting for a user
type RateLimiter struct {
	ID         string        // Unique identifier for the user
	Limit      int           // Maximum allowed requests
	ResetAfter time.Duration // Duration after which requests reset
}

// NewRateLimiter creates a new RateLimiter instance
func NewRateLimiter(id string, limit int, resetAfter time.Duration) *RateLimiter {
	return &RateLimiter{
		ID:         id,
		Limit:      limit,
		ResetAfter: resetAfter,
	}
}
