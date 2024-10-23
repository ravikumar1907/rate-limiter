package cassandra

import (
	"log"
	"time"

	"github.com/gocql/gocql"
	"github.com/ravikumar1907/rate-limiter/internal/limiter"
)

type CassandraClient struct {
	session *gocql.Session
}

// NewCassandraClient initializes a new Cassandra client
func NewCassandraClient() *CassandraClient {
	cluster := gocql.NewCluster("localhost")
	cluster.Keyspace = "rate_limiter"
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatalf("Failed to connect to Cassandra: %v", err)
	}

	return &CassandraClient{session: session}
}

// SaveRateLimitConfig saves the rate limit configuration to Cassandra
func (cc *CassandraClient) SaveRateLimitConfig(id string, limit int, resetAfter int64) error {
	duration := time.Duration(resetAfter)
	if err := cc.session.Query(`INSERT INTO rate_limits (id, limit, reset_after) VALUES (?, ?, ?)`,
		id, limit, duration).Exec(); err != nil {
		return err
	}
	return nil
}

// internal/cassandra/cassandra_client.go

// UpdateRateLimitConfig updates the rate limit configuration in Cassandra
func (cc *CassandraClient) UpdateRateLimitConfig(id string, newLimit int, newResetAfter int) error {
	query := `UPDATE rate_limits SET limit = ?, reset_after = ? WHERE id = ?`
	if err := cc.session.Query(query, newLimit, newResetAfter, id).Exec(); err != nil {
		return err
	}
	return nil
}

// LoadRateLimitConfig loads the rate limit configuration from Cassandra
func (cc *CassandraClient) LoadRateLimitConfig(id string) (int, int64, error) {
	var limit int
	var resetAfter int64

	if err := cc.session.Query(`SELECT limit, reset_after FROM rate_limits WHERE id = ?`, id).Scan(&limit, &resetAfter); err != nil {
		return 0, 0, err
	}
	return limit, resetAfter, nil
}

// DeleteRateLimitConfig deletes a rate limiter configuration from Cassandra
func (cc *CassandraClient) DeleteRateLimitConfig(id string) error {
	query := "DELETE FROM rate_limits WHERE id = ?"
	return cc.session.Query(query, id).Exec()
}

func (cc *CassandraClient) GetAllRateLimitConfigs() ([]*limiter.RateLimiter, error) {
	var configs []*limiter.RateLimiter
	query := "SELECT id, limit, current_requests, reset_after FROM rate_limits"
	iter := cc.session.Query(query).Iter()

	var id string
	var limit, currentRequests int
	var resetAfter time.Duration

	for iter.Scan(&id, &limit, &currentRequests, &resetAfter) {
		config := &limiter.RateLimiter{
			ID:         id,
			Limit:      limit,
			ResetAfter: resetAfter,
		}
		configs = append(configs, config)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return configs, nil
}
