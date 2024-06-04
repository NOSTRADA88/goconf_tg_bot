// Package redis provides a Redis client for caching.
package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

// Nil is a constant representing a Redis nil reply.
const Nil = redis.Nil

// CacheClient is an interface that defines methods for a caching client.
type CacheClient interface {
	Set(ctx context.Context, key int64, value interface{}, duration time.Duration) error // Set adds a value to the cache with a specified duration.
	Get(ctx context.Context, key int64) (string, error)                                  // Get retrieves a value from the cache by key.
}

// Client is a struct that wraps the Redis client and implements the CacheClient interface.
type Client struct {
	Rdb *redis.Client // Rdb is the underlying Redis client.
}

// New creates a new Client and returns a pointer to it.
// It takes a host and port as arguments and connects to the Redis server at that address.
func New(host string, port int) *Client {
	rdb := redis.NewClient(&redis.Options{Addr: fmt.Sprintf("%s:%v", host, port), Password: "", DB: 0})
	return &Client{Rdb: rdb}
}

// Get retrieves a value from the Redis cache by key.
// It returns the value as a string and an error if there is one.
func (c *Client) Get(ctx context.Context, key int64) (string, error) {
	return c.Rdb.Get(ctx, strconv.Itoa(int(key))).Result()
}

// Set adds a value to the Redis cache with a specified duration.
// The key is an int64, the value is an interface{}, and the duration is a time.Duration.
// It returns an error if there is one.
func (c *Client) Set(ctx context.Context, key int64, value interface{}, duration time.Duration) error {
	return c.Rdb.Set(ctx, strconv.Itoa(int(key)), value, duration).Err()
}
