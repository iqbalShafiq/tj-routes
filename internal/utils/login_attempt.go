package utils

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"tj-routes/internal/cache"
	"tj-routes/internal/config"
)

// LoginAttemptTracker tracks failed login attempts for account lockout
type LoginAttemptTracker struct {
	cache cache.Cache
	cfg   config.SecurityConfig
}

// NewLoginAttemptTracker creates a new login attempt tracker
func NewLoginAttemptTracker(cacheInstance cache.Cache, cfg config.SecurityConfig) *LoginAttemptTracker {
	// Set defaults if not configured
	if cfg.MaxLoginAttempts == 0 {
		cfg.MaxLoginAttempts = 5
	}
	if cfg.AccountLockoutMinutes == 0 {
		cfg.AccountLockoutMinutes = 15
	}

	return &LoginAttemptTracker{
		cache: cacheInstance,
		cfg:   cfg,
	}
}

// key generates the cache key for login attempts
func (t *LoginAttemptTracker) key(email string) string {
	return fmt.Sprintf("login:attempts:%s", email)
}

// key generates the cache key for lockout
func (t *LoginAttemptTracker) lockoutKey(email string) string {
	return fmt.Sprintf("login:lockout:%s", email)
}

// RecordFailedAttempt records a failed login attempt and returns the current count
func (t *LoginAttemptTracker) RecordFailedAttempt(ctx context.Context, email string) (int, error) {
	key := t.key(email)

	// Get current attempt count
	var count int
	err := t.cache.Get(ctx, key, &count)
	if err != nil {
		// Key doesn't exist, start at 1
		count = 1
	} else {
		count++
	}

	// Increment attempt count with TTL
	// TTL resets on each attempt (sliding window)
	ttl := 15 * time.Minute
	if err := t.cache.Set(ctx, key, count, ttl); err != nil {
		return count, fmt.Errorf("failed to record login attempt: %w", err)
	}

	return count, nil
}

// CheckLockedOut checks if an account is locked out
func (t *LoginAttemptTracker) CheckLockedOut(ctx context.Context, email string) (bool, error) {
	key := t.lockoutKey(email)
	err := t.cache.Get(ctx, key, &struct{}{})
	if err != nil {
		// Key doesn't exist, not locked out
		return false, nil
	}
	return true, nil
}

// LockOut locks an account for the configured duration
func (t *LoginAttemptTracker) LockOut(ctx context.Context, email string) error {
	key := t.lockoutKey(email)
	return t.cache.Set(ctx, key, struct{}{}, t.cfg.LockoutDuration())
}

// GetRemainingLockoutTime returns the remaining lockout time in seconds, or 0 if not locked out
func (t *LoginAttemptTracker) GetRemainingLockoutTime(ctx context.Context, email string) (int, error) {
	key := t.lockoutKey(email)
	// Check if the key exists and get its TTL
	// This is a workaround since our cache interface doesn't expose TTL
	var exists bool
	err := t.cache.Get(ctx, key, &exists)
	if err != nil {
		// Key doesn't exist, not locked out
		return 0, nil
	}
	return t.cfg.AccountLockoutMinutes * 60, nil
}

// ClearFailedAttempts clears failed login attempts for an email
func (t *LoginAttemptTracker) ClearFailedAttempts(ctx context.Context, email string) error {
	key := t.key(email)
	return t.cache.Delete(ctx, key)
}

// GetFailedAttempts returns the current count of failed attempts
func (t *LoginAttemptTracker) GetFailedAttempts(ctx context.Context, email string) (int, error) {
	key := t.key(email)
	var count int
	err := t.cache.Get(ctx, key, &count)
	if err != nil {
		return 0, nil
	}
	return count, nil
}

// IsLoginAllowed checks if login is allowed for the given email
// Returns (allowed, error)
func (t *LoginAttemptTracker) IsLoginAllowed(ctx context.Context, email string) (bool, int, error) {
	// Check if locked out
	lockedOut, err := t.CheckLockedOut(ctx, email)
	if err != nil {
		return false, 0, fmt.Errorf("failed to check lockout status: %w", err)
	}
	if lockedOut {
		remaining, err := t.GetRemainingLockoutTime(ctx, email)
		if err != nil {
			remaining = t.cfg.AccountLockoutMinutes * 60
		}
		return false, remaining, nil
	}

	// Get current failed attempts
	attempts, err := t.GetFailedAttempts(ctx, email)
	if err != nil {
		return false, 0, fmt.Errorf("failed to get failed attempts: %w", err)
	}

	remainingAttempts := t.cfg.MaxLoginAttempts - attempts
	return true, remainingAttempts, nil
}

// RecordLoginSuccess clears failed attempts on successful login
func (t *LoginAttemptTracker) RecordLoginSuccess(ctx context.Context, email string) error {
	return t.ClearFailedAttempts(ctx, email)
}

// RecordLoginFailure records a failed attempt and returns true if now locked out
func (t *LoginAttemptTracker) RecordLoginFailure(ctx context.Context, email string) (lockedOut bool, remainingAttempts int, err error) {
	count, err := t.RecordFailedAttempt(ctx, email)
	if err != nil {
		return false, 0, err
	}

	remainingAttempts = t.cfg.MaxLoginAttempts - count

	if count >= t.cfg.MaxLoginAttempts {
		// Lock the account
		if err := t.LockOut(ctx, email); err != nil {
			return false, remainingAttempts, fmt.Errorf("failed to lock account: %w", err)
		}
		return true, remainingAttempts, nil
	}

	return false, remainingAttempts, nil
}

// LockoutInfo contains information about account lockout status
type LockoutInfo struct {
	IsLockedOut       bool `json:"is_locked_out"`
	RemainingAttempts int  `json:"remaining_attempts,omitempty"`
	RemainingSeconds  int  `json:"remaining_seconds,omitempty"`
	MaxAttempts       int  `json:"max_attempts"`
}

// GetLockoutInfo returns lockout information for an email
func (t *LoginAttemptTracker) GetLockoutInfo(ctx context.Context, email string) (*LockoutInfo, error) {
	info := &LockoutInfo{
		MaxAttempts: t.cfg.MaxLoginAttempts,
	}

	lockedOut, err := t.CheckLockedOut(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check lockout status: %w", err)
	}
	info.IsLockedOut = lockedOut

	if lockedOut {
		remaining, err := t.GetRemainingLockoutTime(ctx, email)
		if err != nil {
			remaining = t.cfg.AccountLockoutMinutes * 60
		}
		info.RemainingSeconds = remaining
	} else {
		attempts, err := t.GetFailedAttempts(ctx, email)
		if err != nil {
			return nil, fmt.Errorf("failed to get failed attempts: %w", err)
		}
		info.RemainingAttempts = t.cfg.MaxLoginAttempts - attempts
	}

	return info, nil
}

// Convert attempt count from interface{} to int
func parseAttemptCount(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("unknown type for attempt count: %T", value)
	}
}
