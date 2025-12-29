package utils

import (
	"errors"
	"time"

	"tj-routes/internal/config"
	"tj-routes/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken is returned when the token is invalid
var ErrInvalidToken = errors.New("invalid token")

// ErrExpiredToken is returned when the token has expired
var ErrExpiredToken = errors.New("token has expired")

// ErrInvalidIssuer is returned when the token issuer is invalid
var ErrInvalidIssuer = errors.New("invalid token issuer")

// ErrInvalidAudience is returned when the token audience is invalid
var ErrInvalidAudience = errors.New("invalid token audience")

// ErrTokenNotYetValid is returned when the token is not yet valid
var ErrTokenNotYetValid = errors.New("token not yet valid")

type Claims struct {
	UserID uint            `json:"user_id"`
	Email  string          `json:"email"`
	Role   models.UserRole `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(user *models.User, cfg *config.JWTConfig) (string, error) {
	expirationTime := time.Now().Add(cfg.ExpirationDuration())

	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.Issuer,
			Audience:  []string{cfg.Audience},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

func GenerateRefreshToken(user *models.User, cfg *config.JWTConfig) (string, error) {
	expirationTime := time.Now().Add(cfg.RefreshExpirationDuration())

	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.Issuer,
			Audience:  []string{cfg.Audience},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

func ValidateToken(tokenString string, secret string, expectedIssuer string, expectedAudience string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, err
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// Validate issuer if expectedIssuer is provided
	if expectedIssuer != "" {
		issuer, err := claims.GetIssuer()
		if err != nil || issuer != expectedIssuer {
			return nil, ErrInvalidIssuer
		}
	}

	// Validate audience if expectedAudience is provided
	if expectedAudience != "" {
		audience, err := claims.GetAudience()
		if err != nil {
			return nil, ErrInvalidAudience
		}
		found := false
		for _, aud := range audience {
			if aud == expectedAudience {
				found = true
				break
			}
		}
		if !found {
			return nil, ErrInvalidAudience
		}
	}

	// Validate NotBefore (nbf) claim
	notBefore, err := claims.GetNotBefore()
	if err == nil && notBefore != nil {
		if time.Now().Before(notBefore.Time) {
			return nil, ErrTokenNotYetValid
		}
	}

	return claims, nil
}

// ValidateTokenWithoutClaimsValidation validates token signature without validating issuer/audience
// This is useful for backwards compatibility or when these claims are not set
func ValidateTokenWithoutClaimsValidation(tokenString string, secret string) (*Claims, error) {
	return ValidateToken(tokenString, secret, "", "")
}
