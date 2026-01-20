package cache

import "fmt"

// PasswordResetTokenKey generates a cache key for password reset tokens
func PasswordResetTokenKey(token string) string {
	return fmt.Sprintf("password:reset:token:%s", token)
}

// PasswordResetRateLimitKey generates a cache key for rate limiting password reset requests per email
func PasswordResetRateLimitKey(email string) string {
	return fmt.Sprintf("password:reset:ratelimit:%s", email)
}
