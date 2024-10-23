package middleware

import (
	"net/http"

	"github.com/ravikumar1907/rate-limiter/internal/limiter"
)

func RateLimitMiddleware(rl *limiter.RateLimiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !rl.Allow() {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		http.HandlerFunc(http.NotFound).ServeHTTP(w, r) // Placeholder for actual handler
	}
}
