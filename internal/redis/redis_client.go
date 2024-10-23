// internal/redis/redis_client.go

package redis

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	ctx         = context.Background()
	redisClient *redis.Client
)

func Initialize() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis server address
		Password: "",               // No password set
		DB:       0,                // use default DB
	})
}

type RateLimitData struct {
	Limit      int   `json:"limit"`
	ResetAfter int64 `json:"reset_after"`
}

// InitializeRateLimitConfig saves the rate limit configuration in Redis as JSON
func InitializeRateLimitConfig(id string, limit int, resetAfter int64) error {
	rateLimitData := RateLimitData{
		Limit:      limit,
		ResetAfter: resetAfter,
	}

	data, err := json.Marshal(rateLimitData)
	if err != nil {
		return err
	}

	_, err = redisClient.Set(ctx, id, data, time.Duration(resetAfter)*time.Second).Result()
	return err
}

// internal/redis/redis_client.go

// UpdateRateLimitConfig updates the rate limit configuration in Redis
func UpdateRateLimitConfig(id string, newLimit int, newResetAfter int64) error {
	// Create the updated RateLimitData struct
	rateLimitData := RateLimitData{
		Limit:      newLimit,
		ResetAfter: newResetAfter,
	}

	// Marshal the struct into a JSON string
	data, err := json.Marshal(rateLimitData)
	if err != nil {
		return err
	}

	// Update the value in Redis and set a new TTL
	_, err = redisClient.Set(ctx, id, data, time.Duration(newResetAfter)*time.Second).Result()
	if err != nil {
		return err
	}

	return nil
}

// IncrementCurrentRequests increments the current request count for a given ID
func IncrementCurrentRequests(id string, resetAfter time.Duration) (int, error) {
	currentRequests, err := redisClient.Incr(ctx, id).Result()
	if err != nil {
		return 0, err
	}

	if currentRequests == 1 {
		err = redisClient.Expire(ctx, id, resetAfter).Err()
		if err != nil {
			return 0, err
		}
	}

	return int(currentRequests), nil
}

// GetRateLimitConfig retrieves the rate limit configuration from Redis
func GetRateLimitConfig(id string) (int, int64, error) {
	data, err := redisClient.Get(ctx, id).Result()
	if err != nil {
		return 0, 0, err
	}

	var rateLimitData RateLimitData
	err = json.Unmarshal([]byte(data), &rateLimitData)
	if err != nil {
		return 0, 0, errors.New("failed to unmarshal rate limit data")
	}

	return rateLimitData.Limit, rateLimitData.ResetAfter, nil
}

// internal/redis/redis_client.go

// DeleteRateLimitConfig deletes the rate limit configuration and current request count for the given client ID from Redis
func DeleteRateLimitConfig(id string) error {
	// Use Redis pipeline to execute multiple delete commands in one go
	pipeline := redisClient.TxPipeline()

	// Delete the rate limit config keys (limit and reset_after) and current requests key
	pipeline.Del(ctx, id+":limit")
	pipeline.Del(ctx, id+":reset_after")
	pipeline.Del(ctx, id+":current_requests")

	// Execute the pipeline
	_, err := pipeline.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

// GetCurrentRequests retrieves the current request count for the given client ID from Redis
func GetCurrentRequests(id string) (int, error) {
	// Attempt to get the current request count from Redis
	currentRequests, err := redisClient.Get(ctx, id+":current_requests").Result()
	if err == redis.Nil {
		// If key does not exist, return 0 as the initial request count
		return 0, nil
	} else if err != nil {
		// If there's any other error, return it
		return 0, err
	}

	// Convert the string result from Redis into an integer
	count, err := strconv.Atoi(currentRequests)
	if err != nil {
		return 0, err
	}

	return count, nil
}
