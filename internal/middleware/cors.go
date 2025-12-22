package middleware

import (
	"strings"

	"tj-routes/internal/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware creates a CORS middleware based on configuration
func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	config := cors.DefaultConfig()

	// Parse allowed origins
	allowedOrigins := cfg.Server.AllowedOrigin
	if allowedOrigins == "*" {
		// Only allow * in development
		if cfg.Server.Environment == "production" {
			// In production, default to empty (no CORS) if not configured
			config.AllowOrigins = []string{}
		} else {
			config.AllowAllOrigins = true
		}
	} else {
		// Split comma-separated origins
		origins := strings.Split(allowedOrigins, ",")
		for i, origin := range origins {
			origins[i] = strings.TrimSpace(origin)
		}
		config.AllowOrigins = origins
	}

	config.AllowCredentials = true
	config.AllowHeaders = []string{
		"Content-Type",
		"Content-Length",
		"Accept-Encoding",
		"X-CSRF-Token",
		"Authorization",
		"accept",
		"origin",
		"Cache-Control",
		"X-Requested-With",
	}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}

	return cors.New(config)
}

