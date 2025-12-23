package cache

import (
	"context"
	"time"
)

// NoOpCache is a no-op cache implementation that does nothing
// Used when caching is disabled or Redis is unavailable
type NoOpCache struct{}

// NewNoOpCache creates a new no-op cache instance
func NewNoOpCache() *NoOpCache {
	return &NoOpCache{}
}

// Get always returns ErrCacheMiss
func (n *NoOpCache) Get(ctx context.Context, key string, dest interface{}) error {
	return ErrCacheMiss
}

// Set does nothing
func (n *NoOpCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return nil
}

// Delete does nothing
func (n *NoOpCache) Delete(ctx context.Context, key string) error {
	return nil
}

// InvalidatePattern does nothing
func (n *NoOpCache) InvalidatePattern(ctx context.Context, pattern string) error {
	return nil
}

// IsAvailable always returns false
func (n *NoOpCache) IsAvailable(ctx context.Context) bool {
	return false
}

