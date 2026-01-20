package utils

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"tj-routes/internal/cache"
)

const (
	// PasswordResetTokenTTL is the duration a password reset token is valid (1 hour)
	PasswordResetTokenTTL = 1 * time.Hour
	// PasswordResetRateLimitTTL is the duration for rate limit window (1 hour)
	PasswordResetRateLimitTTL = 1 * time.Hour
	// PasswordResetMaxRequests is the maximum number of reset requests per email per hour
	PasswordResetMaxRequests = 3
)

var (
	ErrRateLimitExceeded  = errors.New("too many password reset requests, please try again later")
	ErrInvalidResetToken  = errors.New("invalid or expired password reset token")
)

// PasswordResetManager handles password reset token operations
type PasswordResetManager struct {
	cache cache.Cache
}

// NewPasswordResetManager creates a new password reset manager
func NewPasswordResetManager(cacheInstance cache.Cache) *PasswordResetManager {
	return &PasswordResetManager{
		cache: cacheInstance,
	}
}

// GenerateResetToken generates a cryptographically secure 32-byte token (64 hex characters)
func (m *PasswordResetManager) GenerateResetToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate reset token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// StoreResetToken stores a password reset token in cache with the user ID
func (m *PasswordResetManager) StoreResetToken(ctx context.Context, token string, userID uint) error {
	if m.cache == nil {
		return errors.New("cache not available")
	}

	key := cache.PasswordResetTokenKey(token)
	return m.cache.Set(ctx, key, userID, PasswordResetTokenTTL)
}

// ValidateResetToken validates a password reset token and returns the associated user ID
func (m *PasswordResetManager) ValidateResetToken(ctx context.Context, token string) (uint, error) {
	if m.cache == nil {
		return 0, errors.New("cache not available")
	}

	key := cache.PasswordResetTokenKey(token)
	var userID uint
	if err := m.cache.Get(ctx, key, &userID); err != nil {
		return 0, ErrInvalidResetToken
	}

	return userID, nil
}

// InvalidateResetToken removes a password reset token from cache (single-use)
func (m *PasswordResetManager) InvalidateResetToken(ctx context.Context, token string) error {
	if m.cache == nil {
		return errors.New("cache not available")
	}

	key := cache.PasswordResetTokenKey(token)
	return m.cache.Delete(ctx, key)
}

// CheckRateLimit checks if the email has exceeded the rate limit for password reset requests
func (m *PasswordResetManager) CheckRateLimit(ctx context.Context, email string) error {
	if m.cache == nil {
		// If cache is not available, skip rate limiting
		return nil
	}

	key := cache.PasswordResetRateLimitKey(email)
	var count int
	if err := m.cache.Get(ctx, key, &count); err != nil {
		// Key doesn't exist, no requests yet
		return nil
	}

	if count >= PasswordResetMaxRequests {
		return ErrRateLimitExceeded
	}

	return nil
}

// IncrementRateLimit increments the rate limit counter for an email
func (m *PasswordResetManager) IncrementRateLimit(ctx context.Context, email string) error {
	if m.cache == nil {
		return nil
	}

	key := cache.PasswordResetRateLimitKey(email)
	var count int
	if err := m.cache.Get(ctx, key, &count); err != nil {
		// Key doesn't exist, start fresh
		count = 0
	}

	count++
	return m.cache.Set(ctx, key, count, PasswordResetRateLimitTTL)
}
