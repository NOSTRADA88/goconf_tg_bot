package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

const Nil = redis.Nil

// CacheClient interface defines and describes Client methods
type CacheClient interface {
	Ping(ctx context.Context) error
	Set(ctx context.Context, key int64, value interface{}, duration time.Duration) error
	Get(ctx context.Context, key int64) (string, error)
}

// Client struct is a shell on *redis.Client that implements CacheClient interface methods
type Client struct {
	Rdb *redis.Client
}

// NewClient function creates and returns Client pointer
func NewClient(host string, port int) *Client {
	rdb := redis.NewClient(&redis.Options{Addr: fmt.Sprintf("%s:%v", host, port)})
	return &Client{Rdb: rdb}
}

// Get method gets value from redis by key
func (c *Client) Get(ctx context.Context, key int64) (string, error) {
	return c.Rdb.Get(ctx, strconv.Itoa(int(key))).Result()
}

// Set method sets value by key, value (any) and duration (0 == infinite living time).
func (c *Client) Set(ctx context.Context, key int64, value interface{}, duration time.Duration) error {
	return c.Rdb.Set(ctx, strconv.Itoa(int(key)), value, duration).Err()
}

// Ping method checks redis status: up or not.
func (c *Client) Ping(ctx context.Context) error {
	return c.Rdb.Ping(ctx).Err()
}
