package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ravikumar1907/rate-limiter/internal/cassandra"
	"github.com/ravikumar1907/rate-limiter/internal/redis"
)

// RateLimitConfig represents the rate limit configuration
type RateLimitConfig struct {
	ID         string `json:"id"`
	Limit      int    `json:"limit"`
	ResetAfter int64  `json:"reset_after"` // Duration in seconds
}

// HandleCreateRateLimit creates a new rate limiter configuration
func HandleCreateRateLimit(w http.ResponseWriter, r *http.Request) {
	var config RateLimitConfig

	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Save to Cassandra
	cc := cassandra.NewCassandraClient()
	err := cc.SaveRateLimitConfig(config.ID, config.Limit, config.ResetAfter)
	if err != nil {
		http.Error(w, "Error saving to database", http.StatusInternalServerError)
		return
	}

	// Also, initialize the limit in Redis
	err = redis.InitializeRateLimitConfig(config.ID, config.Limit, config.ResetAfter)
	if err != nil {
		http.Error(w, "Error initializing Redis", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(config)
}

// HandleGetRateLimit retrieves the rate limiter configuration by ID
func HandleGetRateLimit(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	// Attempt to fetch from Redis
	limit, resetAfter, err := redis.GetRateLimitConfig(id)
	if err != nil {
		// If not found in Redis, fall back to Cassandra
		cc := cassandra.NewCassandraClient()
		limit, resetAfter, err = cc.LoadRateLimitConfig(id)
		if err != nil {
			http.Error(w, "Rate limiter not found", http.StatusNotFound)
			return
		}
	}

	config := RateLimitConfig{
		ID:         id,
		Limit:      limit,
		ResetAfter: resetAfter,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(config)
}

// HandleUpdateRateLimit updates an existing rate limiter configuration
func HandleUpdateRateLimit(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	cc := cassandra.NewCassandraClient()
	// Fetch existing settings from Redis
	_, _, err := redis.GetRateLimitConfig(id)
	if err != nil {
		// If not found in Redis, fall back to Cassandra
		_, _, err = cc.LoadRateLimitConfig(id)
		if err != nil {
			http.Error(w, "Rate limiter not found", http.StatusNotFound)
			return
		}
	}

	var config RateLimitConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Update in Cassandra
	err = cc.SaveRateLimitConfig(config.ID, config.Limit, config.ResetAfter)
	if err != nil {
		http.Error(w, "Error updating in database", http.StatusInternalServerError)
		return
	}

	// Also update in Redis
	err = redis.UpdateRateLimitConfig(config.ID, config.Limit, config.ResetAfter)
	if err != nil {
		http.Error(w, "Error updating in Redis", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(config)
}

// HandleDeleteRateLimit deletes a rate limiter configuration by ID
func HandleDeleteRateLimit(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	cc := cassandra.NewCassandraClient()
	// Remove from Cassandra
	err := cc.DeleteRateLimitConfig(id)
	if err != nil {
		http.Error(w, "Error deleting from database", http.StatusInternalServerError)
		return
	}

	// Remove from Redis
	err = redis.DeleteRateLimitConfig(id)
	if err != nil {
		http.Error(w, "Error deleting from Redis", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleCheckRateLimit checks if the current request should be admitted or rejected
func HandleCheckRateLimit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID string `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	cc := cassandra.NewCassandraClient()
	// Attempt to fetch rate limit config from Redis
	limit, resetAfter, err := redis.GetRateLimitConfig(req.ID)
	if err != nil {
		// If not found in Redis, fall back to Cassandra
		limit, resetAfter, err = cc.LoadRateLimitConfig(req.ID)
		if err != nil {
			http.Error(w, "Rate limiter not found", http.StatusNotFound)
			return
		}
	}

	currentRequests, err := redis.GetCurrentRequests(req.ID)
	if err != nil {
		http.Error(w, "Error retrieving current request count from Redis", http.StatusInternalServerError)
		return
	}

	if currentRequests >= limit {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	// Increment the request count
	_, err = redis.IncrementCurrentRequests(req.ID, time.Duration(resetAfter))
	if err != nil {
		http.Error(w, "Error incrementing request count", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Request admitted"})
}

// func LoadConfigFromDBToRedis() error {
//     // Assuming there's a function in the cassandra package to retrieve all rate limits
//     rateLimits, err := cassandra.GetAllRateLimitConfigs()
//     if err != nil {
//         return err
//     }

//     for _, rateLimit := range rateLimits {
//         err = InitializeRateLimitConfig(rateLimit.Limit, rateLimit.ResetAfter, rateLimit.ID)
//         if err != nil {
//             return err
//         }
//     }

//     return nil
// }
