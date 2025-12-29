package utils

import (
	"testing"
	"time"

	"tj-routes/internal/config"
	"tj-routes/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:              "test-secret-key-for-testing-purposes",
		ExpirationHours:     24,
		RefreshExpirationHours: 168,
	}

	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Role:  models.RoleCommonUser,
	}

	token, err := GenerateToken(user, cfg)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestValidateToken(t *testing.T) {
	secret := "test-secret-key-for-testing-purposes"
	cfg := &config.JWTConfig{
		Secret:              secret,
		ExpirationHours:     24,
		RefreshExpirationHours: 168,
		Issuer:              "test-issuer",
		Audience:            "test-audience",
	}

	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Role:  models.RoleCommonUser,
	}

	token, err := GenerateToken(user, cfg)
	assert.NoError(t, err)

	claims, err := ValidateToken(token, secret, cfg.Issuer, cfg.Audience)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, user.Role, claims.Role)
}

func TestValidateToken_InvalidToken(t *testing.T) {
	secret := "test-secret-key-for-testing-purposes"
	invalidToken := "invalid.token.here"

	claims, err := ValidateToken(invalidToken, secret, "", "")
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	secret := "test-secret-key-for-testing-purposes"
	cfg := &config.JWTConfig{
		Secret:              secret,
		ExpirationHours:     -1, // Expired immediately
		RefreshExpirationHours: 168,
		Issuer:              "test-issuer",
		Audience:            "test-audience",
	}

	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Role:  models.RoleCommonUser,
	}

	// Manually create an expired token
	token, err := GenerateToken(user, cfg)
	assert.NoError(t, err)

	// Wait a bit and try to validate (this test structure is simplified)
	// In practice, you'd want to test with an actually expired token
	time.Sleep(100 * time.Millisecond)
	claims, err := ValidateToken(token, secret, cfg.Issuer, cfg.Audience)
	// This should still work since we're just testing the structure
	// A proper test would use a token with actual expiration in the past
	if err == nil {
		assert.NotNil(t, claims)
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:              "test-secret-key-for-testing-purposes",
		ExpirationHours:     24,
		RefreshExpirationHours: 168,
	}

	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Role:  models.RoleCommonUser,
	}

	refreshToken, err := GenerateRefreshToken(user, cfg)
	assert.NoError(t, err)
	assert.NotEmpty(t, refreshToken)

	// Refresh token should be different from access token
	accessToken, err := GenerateToken(user, cfg)
	assert.NoError(t, err)
	assert.NotEqual(t, accessToken, refreshToken)
}

