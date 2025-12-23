package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisCache implements the Cache interface using Redis
type RedisCache struct {
	client *redis.Client
	logger *zap.Logger
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(client *redis.Client, logger *zap.Logger) *RedisCache {
	return &RedisCache{
		client: client,
		logger: logger,
	}
}

// Get retrieves a value from cache by key
func (r *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		r.logger.Warn("Cache get error", zap.String("key", key), zap.Error(err))
		return err
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		r.logger.Warn("Cache unmarshal error", zap.String("key", key), zap.Error(err))
		return err
	}

	return nil
}

// Set stores a value in cache with the given key and TTL
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		r.logger.Warn("Cache marshal error", zap.String("key", key), zap.Error(err))
		return err
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		r.logger.Warn("Cache set error", zap.String("key", key), zap.Error(err))
		return err
	}

	return nil
}

// Delete removes a value from cache by key
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		r.logger.Warn("Cache delete error", zap.String("key", key), zap.Error(err))
		return err
	}
	return nil
}

// InvalidatePattern removes all keys matching the pattern
func (r *RedisCache) InvalidatePattern(ctx context.Context, pattern string) error {
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	keys := make([]string, 0)

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		r.logger.Warn("Cache pattern scan error", zap.String("pattern", pattern), zap.Error(err))
		return err
	}

	if len(keys) > 0 {
		if err := r.client.Del(ctx, keys...).Err(); err != nil {
			r.logger.Warn("Cache pattern delete error", zap.String("pattern", pattern), zap.Error(err))
			return err
		}
		r.logger.Info("Cache pattern invalidated", zap.String("pattern", pattern), zap.Int("keys_deleted", len(keys)))
	}

	return nil
}

// IsAvailable checks if the cache is available
func (r *RedisCache) IsAvailable(ctx context.Context) bool {
	if r.client == nil {
		return false
	}
	err := r.client.Ping(ctx).Err()
	return err == nil
}

// ErrCacheMiss is returned when a key is not found in cache
var ErrCacheMiss = fmt.Errorf("cache miss")

