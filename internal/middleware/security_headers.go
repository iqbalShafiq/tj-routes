package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders adds HTTP security headers to all responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// XSS Protection (legacy but still useful for older browsers)
		c.Header("X-XSS-Protection", "1; mode=block")

		// Strict Transport Security (HSTS)
		// Only set in production to avoid issues with development
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Referrer Policy - control how much referrer information is sent
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions Policy - control browser features
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Content Security Policy
		c.Header("Content-Security-Policy", "default-src 'self'")

		c.Next()
	}
}
