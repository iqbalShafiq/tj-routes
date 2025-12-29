package middleware

import (
	"net/http"
	"strings"

	"tj-routes/internal/config"
	"tj-routes/internal/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := utils.ValidateToken(tokenString, cfg.JWT.Secret, cfg.JWT.Issuer, cfg.JWT.Audience)
		if err != nil {
			status := http.StatusUnauthorized
			message := "Invalid or expired token"

			switch err {
			case utils.ErrExpiredToken:
				message = "Token has expired"
			case utils.ErrInvalidIssuer:
				message = "Invalid token issuer"
			case utils.ErrInvalidAudience:
				message = "Invalid token audience"
			case utils.ErrTokenNotYetValid:
				message = "Token not yet valid"
			}

			c.JSON(status, gin.H{"error": message})
			c.Abort()
			return
		}

		// Store user info in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

func OptionalAuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString := parts[1]
				claims, err := utils.ValidateToken(tokenString, cfg.JWT.Secret, cfg.JWT.Issuer, cfg.JWT.Audience)
				if err == nil {
					c.Set("user_id", claims.UserID)
					c.Set("user_email", claims.Email)
					c.Set("user_role", claims.Role)
				}
			}
		}
		c.Next()
	}
}

