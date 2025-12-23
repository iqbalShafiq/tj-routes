package cache

import (
	"context"
	"time"
)

// Cache defines the interface for cache operations
type Cache interface {
	// Get retrieves a value from cache by key
	Get(ctx context.Context, key string, dest interface{}) error

	// Set stores a value in cache with the given key and TTL
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Delete removes a value from cache by key
	Delete(ctx context.Context, key string) error

	// InvalidatePattern removes all keys matching the pattern
	InvalidatePattern(ctx context.Context, pattern string) error

	// IsAvailable checks if the cache is available
	IsAvailable(ctx context.Context) bool
}

