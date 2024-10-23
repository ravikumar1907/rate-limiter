package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ravikumar1907/rate-limiter/api"
	"github.com/ravikumar1907/rate-limiter/internal/cassandra"
	"github.com/ravikumar1907/rate-limiter/internal/redis"
)

// LoadConfigFromDBToRedis loads the rate limit configuration from Cassandra to Redis
func LoadConfigFromDBToRedis() error {
	// Assuming there's a function in the cassandra package to retrieve all rate limits
	cc := cassandra.NewCassandraClient()
	rateLimits, err := cc.GetAllRateLimitConfigs()
	if err != nil {
		return err
	}

	for _, rateLimit := range rateLimits {
		err = redis.InitializeRateLimitConfig(rateLimit.ID, rateLimit.Limit, int64(rateLimit.ResetAfter))
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {

	LoadConfigFromDBToRedis()

	// Set up HTTP router
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/rate-limit", api.HandleCreateRateLimit).Methods("POST")
	r.HandleFunc("/api/v1/rate-limit", api.HandleGetRateLimit).Methods("GET")
	r.HandleFunc("/api/v1/rate-limit", api.HandleUpdateRateLimit).Methods("PUT")
	r.HandleFunc("/api/v1/rate-limit", api.HandleDeleteRateLimit).Methods("DELETE")
	r.HandleFunc("/api/v1/check-rate-limit", api.HandleCheckRateLimit).Methods("POST")

	// Start the HTTP server
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Could not start server: %s", err)
	}
}
