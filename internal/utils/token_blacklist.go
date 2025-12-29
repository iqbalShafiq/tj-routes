package utils

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"tj-routes/internal/cache"
	"tj-routes/internal/config"
	"tj-routes/internal/models"
)

// TokenBlacklist provides JWT token blacklisting functionality
type TokenBlacklist struct {
	cache cache.Cache
	cfg   *config.JWTConfig
}

// NewTokenBlacklist creates a new token blacklist
func NewTokenBlacklist(cacheInstance cache.Cache, cfg *config.JWTConfig) *TokenBlacklist {
	return &TokenBlacklist{
		cache: cacheInstance,
		cfg:   cfg,
	}
}

// hashToken creates a SHA256 hash of the token for storage
func (t *TokenBlacklist) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// blacklistKey returns the Redis key for blacklisted tokens
func (t *TokenBlacklist) blacklistKey(tokenHash string) string {
	return fmt.Sprintf("jwt:blacklist:%s", tokenHash)
}

// fingerprintKey returns the Redis key for token fingerprints
func (t *TokenBlacklist) fingerprintKey(userID uint, fingerprint string) string {
	return fmt.Sprintf("jwt:fingerprint:%d:%s", userID, fingerprint)
}

// CreateTokenFingerprint creates a fingerprint for the token based on request characteristics
func CreateTokenFingerprint(ctx context.Context) string {
	// In a real implementation, you might want to include:
	// - Client IP
	// - User-Agent
	// - Other request characteristics

	// For simplicity, we'll use a time-based fingerprint that rotates
	// In production, you should include actual request characteristics
	return fmt.Sprintf("%d", time.Now().Unix()/3600) // Rotates every hour
}

// BlacklistToken adds a token to the blacklist
func (t *TokenBlacklist) BlacklistToken(ctx context.Context, token string) error {
	tokenHash := t.hashToken(token)
	key := t.blacklistKey(tokenHash)

	// Calculate remaining TTL for the token
	// For access tokens, we use a shorter TTL
	ttl := t.cfg.ExpirationDuration()
	if ttl == 0 {
		ttl = 24 * time.Hour // Default to 24 hours
	}

	return t.cache.Set(ctx, key, struct{}{}, ttl)
}

// IsBlacklisted checks if a token is blacklisted
func (t *TokenBlacklist) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	tokenHash := t.hashToken(token)
	key := t.blacklistKey(tokenHash)

	err := t.cache.Get(ctx, key, &struct{}{})
	if err != nil {
		// Key doesn't exist, not blacklisted
		return false, nil
	}
	return true, nil
}

// BlacklistRefreshToken blacklists a refresh token with the full refresh duration
func (t *TokenBlacklist) BlacklistRefreshToken(ctx context.Context, token string) error {
	tokenHash := t.hashToken(token)
	key := t.blacklistKey(tokenHash)

	ttl := t.cfg.RefreshExpirationDuration()
	if ttl == 0 {
		ttl = 168 * time.Hour // Default to 7 days
	}

	return t.cache.Set(ctx, key, struct{}{}, ttl)
}

// StoreRefreshTokenFingerprint stores a fingerprint for a refresh token
// This is used for rotation tracking
func (t *TokenBlacklist) StoreRefreshTokenFingerprint(ctx context.Context, userID uint, fingerprint string, refreshToken string) error {
	fpKey := t.fingerprintKey(userID, fingerprint)

	// Store the token hash associated with this fingerprint
	tokenHash := t.hashToken(refreshToken)

	// TTL should match refresh token duration
	ttl := t.cfg.RefreshExpirationDuration()
	if ttl == 0 {
		ttl = 168 * time.Hour
	}

	return t.cache.Set(ctx, fpKey, tokenHash, ttl)
}

// ValidateRefreshTokenFingerprint validates that a refresh token matches the stored fingerprint
func (t *TokenBlacklist) ValidateRefreshTokenFingerprint(ctx context.Context, userID uint, fingerprint string, refreshToken string) (bool, error) {
	fpKey := t.fingerprintKey(userID, fingerprint)

	var storedTokenHash string
	err := t.cache.Get(ctx, fpKey, &storedTokenHash)
	if err != nil {
		// No stored fingerprint, token might be new (rotation case)
		return false, nil
	}

	expectedHash := t.hashToken(refreshToken)
	return storedTokenHash == expectedHash, nil
}

// InvalidateRefreshTokens invalidates all refresh tokens for a user
func (t *TokenBlacklist) InvalidateRefreshTokens(ctx context.Context, userID uint) error {
	pattern := fmt.Sprintf("jwt:fingerprint:%d:*", userID)
	return t.cache.InvalidatePattern(ctx, pattern)
}

// TokenInfo contains information about a token for debugging/logging
type TokenInfo struct {
	UserID      uint
	Email       string
	ExpiresAt   time.Time
	Blacklisted bool
}

// GetTokenInfo retrieves token information (for debugging/monitoring)
func (t *TokenBlacklist) GetTokenInfo(ctx context.Context, token string, claims *Claims) (*TokenInfo, error) {
	info := &TokenInfo{
		UserID:    claims.UserID,
		Email:     claims.Email,
		ExpiresAt: claims.ExpiresAt.Time,
	}

	blacklisted, err := t.IsBlacklisted(ctx, token)
	if err != nil {
		return nil, err
	}
	info.Blacklisted = blacklisted

	return info, nil
}

// TokenBlacklistService provides higher-level token blacklist operations
type TokenBlacklistService struct {
	blacklist *TokenBlacklist
}

// NewTokenBlacklistService creates a new token blacklist service
func NewTokenBlacklistService(cacheInstance cache.Cache, cfg *config.JWTConfig) *TokenBlacklistService {
	return &TokenBlacklistService{
		blacklist: NewTokenBlacklist(cacheInstance, cfg),
	}
}

// InvalidateUserTokens invalidates all tokens for a user (logout all devices)
func (s *TokenBlacklistService) InvalidateUserTokens(ctx context.Context, userID uint) error {
	// Invalidate refresh token fingerprints
	if err := s.blacklist.InvalidateRefreshTokens(ctx, userID); err != nil {
		return fmt.Errorf("failed to invalidate refresh tokens: %w", err)
	}

	return nil
}

// BlacklistAccessToken blacklists an access token
func (s *TokenBlacklistService) BlacklistAccessToken(ctx context.Context, token string) error {
	return s.blacklist.BlacklistToken(ctx, token)
}

// BlacklistRefreshToken blacklists a refresh token
func (s *TokenBlacklistService) BlacklistRefreshToken(ctx context.Context, token string) error {
	return s.blacklist.BlacklistRefreshToken(ctx, token)
}

// IsTokenBlacklisted checks if a token is blacklisted
func (s *TokenBlacklistService) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	return s.blacklist.IsBlacklisted(ctx, token)
}

// RefreshTokenRotationResult contains the result of a token rotation
type RefreshTokenRotationResult struct {
	NewAccessToken  string
	NewRefreshToken string
	WasRotated      bool
}

// RotateRefreshToken performs refresh token rotation
// Returns new tokens and whether rotation occurred
func (s *TokenBlacklistService) RotateRefreshToken(ctx context.Context, user *models.User, oldRefreshToken string, fingerprint string) (*RefreshTokenRotationResult, error) {
	// Validate the fingerprint
	valid, err := s.blacklist.ValidateRefreshTokenFingerprint(ctx, user.ID, fingerprint, oldRefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to validate fingerprint: %w", err)
	}

	if !valid {
		// Fingerprint not found or doesn't match - this might be a new token
		// Still allow it but don't count as rotation
		return &RefreshTokenRotationResult{
			WasRotated: false,
		}, nil
	}

	// Blacklist the old refresh token
	if err := s.blacklist.BlacklistRefreshToken(ctx, oldRefreshToken); err != nil {
		return nil, fmt.Errorf("failed to blacklist old refresh token: %w", err)
	}

	// Generate new tokens
	cfg := &config.JWTConfig{
		Secret:                 s.blacklist.cfg.Secret,
		ExpirationHours:        s.blacklist.cfg.ExpirationHours,
		RefreshExpirationHours: s.blacklist.cfg.RefreshExpirationHours,
		Issuer:                 s.blacklist.cfg.Issuer,
		Audience:               s.blacklist.cfg.Audience,
	}

	newAccessToken, err := GenerateToken(user, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new access token: %w", err)
	}

	newRefreshToken, err := GenerateRefreshToken(user, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new refresh token: %w", err)
	}

	// Store new fingerprint
	newFingerprint := CreateTokenFingerprint(ctx)
	if err := s.blacklist.StoreRefreshTokenFingerprint(ctx, user.ID, newFingerprint, newRefreshToken); err != nil {
		return nil, fmt.Errorf("failed to store new fingerprint: %w", err)
	}

	return &RefreshTokenRotationResult{
		NewAccessToken:  newAccessToken,
		NewRefreshToken: newRefreshToken,
		WasRotated:      true,
	}, nil
}
