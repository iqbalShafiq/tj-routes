package utils

import (
	"context"
	"fmt"
	"time"

	"tj-routes/internal/cache"
	"tj-routes/internal/config"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// InitRedis initializes a Redis client and returns a cache instance
func InitRedis(cfg *config.Config, logger *zap.Logger) (cache.Cache, error) {
	if !cfg.Cache.Enabled {
		logger.Info("Caching is disabled, using no-op cache")
		return cache.NewNoOpCache(), nil
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Warn("Failed to connect to Redis, using no-op cache", zap.Error(err))
		return cache.NewNoOpCache(), nil
	}

	logger.Info("Redis connected successfully",
		zap.String("host", cfg.Redis.Host),
		zap.String("port", cfg.Redis.Port),
		zap.Int("db", cfg.Redis.DB))

	return cache.NewRedisCache(rdb, logger), nil
}

