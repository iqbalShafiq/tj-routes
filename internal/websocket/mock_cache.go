package websocket

import (
	"context"
	"time"

	"tj-routes/internal/cache"
)

type MockCache struct{}

func (m *MockCache) Get(ctx context.Context, key string, dest interface{}) error {
	return cache.ErrCacheMiss
}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return nil
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	return nil
}

func (m *MockCache) InvalidatePattern(ctx context.Context, pattern string) error {
	return nil
}

func (m *MockCache) IsAvailable(ctx context.Context) bool {
	return true
}
