package utils

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"tj-routes/internal/cache"
)

// OAuthStateStore provides secure OAuth state token storage
type OAuthStateStore struct {
	cache cache.Cache
	ttl   time.Duration
}

// NewOAuthStateStore creates a new OAuth state store
func NewOAuthStateStore(cacheInstance cache.Cache) *OAuthStateStore {
	return &OAuthStateStore{
		cache: cacheInstance,
		ttl:   10 * time.Minute, // OAuth state expires after 10 minutes
	}
}

// stateKey returns the Redis key for an OAuth state
func (s *OAuthStateStore) stateKey(state string) string {
	return fmt.Sprintf("oauth:state:%s", state)
}

// GenerateState generates a cryptographically secure random state token
func (s *OAuthStateStore) GenerateState(ctx context.Context, redirectURL string) (string, error) {
	// Generate 32 bytes of random data
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", fmt.Errorf("failed to generate state token: %w", err)
	}

	state := base64.URLEncoding.EncodeToString(stateBytes)

	// Store state with redirect URL
	key := s.stateKey(state)
	if err := s.cache.Set(ctx, key, redirectURL, s.ttl); err != nil {
		return "", fmt.Errorf("failed to store state token: %w", err)
	}

	return state, nil
}

// ValidateState validates an OAuth state token and returns the associated redirect URL
// Returns (redirectURL, true) if valid, ("", false) if invalid or expired
func (s *OAuthStateStore) ValidateState(ctx context.Context, state string) (string, bool, error) {
	if state == "" {
		return "", false, nil
	}

	key := s.stateKey(state)
	var redirectURL string
	err := s.cache.Get(ctx, key, &redirectURL)

	if err != nil {
		// State doesn't exist or has expired
		return "", false, nil
	}

	// Delete the state after validation (one-time use)
	if err := s.cache.Delete(ctx, key); err != nil {
		// Log but don't fail - the state is still valid
		fmt.Printf("warning: failed to delete OAuth state: %v\n", err)
	}

	return redirectURL, true, nil
}

// StateInfo contains information about an OAuth state for debugging
type StateInfo struct {
	State      string
	RedirectURL string
	ExpiresIn  int // seconds
}

// GetStateInfo returns information about an OAuth state (for debugging)
func (s *OAuthStateStore) GetStateInfo(ctx context.Context, state string) (*StateInfo, error) {
	key := s.stateKey(state)
	var redirectURL string
	err := s.cache.Get(ctx, key, &redirectURL)

	if err != nil {
		return nil, fmt.Errorf("state not found or expired")
	}

	return &StateInfo{
		State:       state,
		RedirectURL: redirectURL,
		ExpiresIn:   int(s.ttl.Seconds()),
	}, nil
}

// InvalidateState invalidates a specific state token
func (s *OAuthStateStore) InvalidateState(ctx context.Context, state string) error {
	key := s.stateKey(state)
	return s.cache.Delete(ctx, key)
}

// InvalidateAllStates invalidates all OAuth state tokens (for security incidents)
func (s *OAuthStateStore) InvalidateAllStates(ctx context.Context) error {
	return s.cache.InvalidatePattern(ctx, "oauth:state:*")
}
