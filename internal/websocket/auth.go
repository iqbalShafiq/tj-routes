package websocket

import (
	"errors"
	"net/http"
	"strings"

	"tj-routes/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

func ValidateWebSocketToken(tokenString string, cfg *config.Config) (uint, error) {
	if tokenString == "" {
		return 0, errors.New("authorization token required")
	}

	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	tokenString = strings.TrimPrefix(tokenString, "Token ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(cfg.JWT.Secret), nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(float64)
		if !ok {
			return 0, errors.New("invalid token claims")
		}
		return uint(userID), nil
	}

	return 0, errors.New("invalid token")
}

func ExtractTokenFromRequest(r *http.Request) string {
	query := r.URL.Query()
	if token := query.Get("token"); token != "" {
		return token
	}

	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}
