package limiter

import (
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(2, 1*time.Second)

	// First two requests should be allowed
	if !rl.Allow() {
		t.Error("Expected request to be allowed")
	}
	if !rl.Allow() {
		t.Error("Expected request to be allowed")
	}

	// Third request should be denied
	if rl.Allow() {
		t.Error("Expected request to be denied")
	}

	// Wait for reset
	time.Sleep(1 * time.Second)

	// Now it should be allowed again
	if !rl.Allow() {
		t.Error("Expected request to be allowed after reset")
	}
}
