// Package cache defines framework-level cache abstractions.
// This package provides interfaces that cache implementations must satisfy,
// allowing users to swap cache backends (Redis, Memcached, Valkey, in-memory, etc.)
// while maintaining compatibility with the framework.
//
// For a production-ready implementation, see the redis package.
package cache

import (
	"context"
	"time"
)

// Cache represents a distributed cache that can store and retrieve key-value pairs.
// Implementations should handle serialization, expiration, and error handling.
type Cache interface {
	// Get retrieves the value for the given key.
	// Returns an error if the key doesn't exist or if there's a connection issue.
	Get(ctx context.Context, key string) (string, error)

	// Set stores a value with the given key and expiration duration.
	// If expiration is 0, the key will not expire.
	Set(ctx context.Context, key string, value string, expiration time.Duration) error

	// Delete removes the key from the cache.
	// Returns an error if there's a connection issue.
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists in the cache.
	Exists(ctx context.Context, key string) (bool, error)

	// GetBytes retrieves the raw byte value for the given key.
	// Useful for storing binary data.
	GetBytes(ctx context.Context, key string) ([]byte, error)

	// SetBytes stores a raw byte value with the given key and expiration.
	SetBytes(ctx context.Context, key string, value []byte, expiration time.Duration) error

	// Expire sets a timeout on a key.
	// Returns true if the timeout was set, false if key doesn't exist.
	Expire(ctx context.Context, key string, expiration time.Duration) (bool, error)

	// TTL returns the remaining time to live of a key.
	// Returns -1 if the key exists but has no expiration.
	// Returns -2 if the key does not exist.
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Ping tests connectivity to the cache backend.
	// Returns an error if the connection is unavailable.
	Ping(ctx context.Context) error

	// Close closes the connection to the cache backend.
	// Should be called during application shutdown.
	Close() error
}

// BatchCache extends Cache with batch operations for improved performance.
// Implementations can optionally implement this interface for bulk operations.
type BatchCache interface {
	Cache

	// MGet retrieves multiple values at once.
	// Returns a map of key-value pairs for keys that exist.
	MGet(ctx context.Context, keys ...string) (map[string]string, error)

	// MSet sets multiple key-value pairs at once.
	// All keys will have the same expiration.
	MSet(ctx context.Context, pairs map[string]string, expiration time.Duration) error

	// MDelete removes multiple keys at once.
	MDelete(ctx context.Context, keys ...string) error
}
