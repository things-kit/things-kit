// Package redis provides a lifecycle-managed Redis client for Things-Kit applications.
// It implements the cache.Cache interface while also providing direct access to the
// underlying Redis client for advanced use cases.
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"github.com/things-kit/module/cache"
	"go.uber.org/fx"
)

// Module provides the Redis client module to the application.
// It provides both the cache.Cache interface and the *redis.Client for power users.
var Module = fx.Module("redis",
	fx.Provide(
		NewConfig,
		NewRedisClient,
		NewRedisCache,
		// Provide as cache.Cache interface
		fx.Annotate(
			func(c *RedisCache) cache.Cache { return c },
			fx.As(new(cache.Cache)),
		),
	),
)

// Config holds the Redis configuration.
type Config struct {
	URL string `mapstructure:"url"` // Redis URL (e.g., redis://localhost:6379/0)
}

// NewConfig creates a new Redis configuration from Viper.
func NewConfig(v *viper.Viper) *Config {
	cfg := &Config{
		URL: "redis://localhost:6379/0", // Default URL
	}

	// Load configuration from viper
	if v != nil {
		_ = v.UnmarshalKey("redis", cfg)
	}

	return cfg
}

// NewRedisClient creates a new Redis client with lifecycle management.
func NewRedisClient(lc fx.Lifecycle, cfg *Config) (*redis.Client, error) {
	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Test connection on startup
			if err := client.Ping(ctx).Err(); err != nil {
				return fmt.Errorf("failed to connect to Redis: %w", err)
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return client.Close()
		},
	})

	return client, nil
}

// RedisCache implements the cache.Cache interface using Redis.
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache implementation.
func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

// Get retrieves the value for the given key.
func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Set stores a value with the given key and expiration duration.
func (c *RedisCache) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

// Delete removes the key from the cache.
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Exists checks if a key exists in the cache.
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetBytes retrieves the raw byte value for the given key.
func (c *RedisCache) GetBytes(ctx context.Context, key string) ([]byte, error) {
	return c.client.Get(ctx, key).Bytes()
}

// SetBytes stores a raw byte value with the given key and expiration.
func (c *RedisCache) SetBytes(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

// Expire sets a timeout on a key.
func (c *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return c.client.Expire(ctx, key, expiration).Result()
}

// TTL returns the remaining time to live of a key.
func (c *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key).Result()
}

// Ping tests connectivity to the cache backend.
func (c *RedisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Close closes the connection to the cache backend.
func (c *RedisCache) Close() error {
	return c.client.Close()
}
