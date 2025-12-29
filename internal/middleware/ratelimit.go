package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// IPEntry tracks IP address with last access time for TTL-based cleanup
type IPEntry struct {
	limiter    *rate.Limiter
	lastAccess time.Time
}

// IPRateLimiter stores rate limiters per IP with TTL-based cleanup
type IPRateLimiter struct {
	ips         map[string]*IPEntry
	mu          *sync.RWMutex
	r           rate.Limit
	b           int
	entryTTL    time.Duration
	maxEntries  int
	cleanupTick time.Duration
}

// NewIPRateLimiter creates a new IP rate limiter with TTL-based cleanup
func NewIPRateLimiter(rps int, burst int) *IPRateLimiter {
	return &IPRateLimiter{
		ips:        make(map[string]*IPEntry),
		mu:         &sync.RWMutex{},
		r:          rate.Limit(rps),
		b:          burst,
		entryTTL:   15 * time.Minute, // Entries expire after 15 minutes of inactivity
		maxEntries: 10000,            // Maximum number of entries before forced cleanup
		cleanupTick: 5 * time.Minute, // Cleanup every 5 minutes
	}
}

// GetLimiter returns the rate limiter for the provided IP address
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	entry, exists := i.ips[ip]
	if !exists {
		entry = &IPEntry{
			limiter:    rate.NewLimiter(i.r, i.b),
			lastAccess: time.Now(),
		}
		i.ips[ip] = entry
	} else {
		entry.lastAccess = time.Now()
	}

	return entry.limiter
}

// Cleanup removes old IP entries that have expired (call periodically)
func (i *IPRateLimiter) Cleanup() {
	i.mu.Lock()
	defer i.mu.Unlock()

	now := time.Now()
	expired := make([]string, 0)

	// Find expired entries
	for ip, entry := range i.ips {
		if now.Sub(entry.lastAccess) > i.entryTTL {
			expired = append(expired, ip)
		}
	}

	// Remove expired entries
	for _, ip := range expired {
		delete(i.ips, ip)
	}

	// If still over max entries, remove oldest entries
	if len(i.ips) > i.maxEntries {
		// Find oldest entries
		type ipTime struct {
			ip        string
			lastAccess time.Time
		}
		entries := make([]ipTime, 0, len(i.ips))
		for ip, entry := range i.ips {
			entries = append(entries, ipTime{ip, entry.lastAccess})
		}

		// Sort by last access time (oldest first)
		for j := 0; j < len(entries)-i.maxEntries; j++ {
			for k := j + 1; k < len(entries); k++ {
				if entries[j].lastAccess.After(entries[k].lastAccess) {
					entries[j], entries[k] = entries[k], entries[j]
				}
			}
		}

		// Remove oldest entries
		toRemove := len(entries) - i.maxEntries
		for j := 0; j < toRemove; j++ {
			delete(i.ips, entries[j].ip)
		}
	}
}

// GetStats returns rate limiter statistics
func (i *IPRateLimiter) GetStats() map[string]interface{} {
	i.mu.RLock()
	defer i.mu.RUnlock()

	now := time.Now()
	activeCount := 0
	expiredCount := 0

	for _, entry := range i.ips {
		if now.Sub(entry.lastAccess) > i.entryTTL {
			expiredCount++
		} else {
			activeCount++
		}
	}

	return map[string]interface{}{
		"total_entries":  len(i.ips),
		"active_entries": activeCount,
		"expired_entries": expiredCount,
		"max_entries":    i.maxEntries,
		"entry_ttl":      i.entryTTL.String(),
	}
}

var (
	// Default rate limiter: 100 requests per second, burst of 200
	defaultLimiter = NewIPRateLimiter(100, 200)
)

// RateLimitMiddleware creates a rate limiting middleware
// rps: requests per second, burst: burst size
func RateLimitMiddleware(rps int, burst int) gin.HandlerFunc {
	var limiter *IPRateLimiter
	if rps > 0 && burst > 0 {
		limiter = NewIPRateLimiter(rps, burst)
	} else {
		limiter = defaultLimiter
	}

	// Cleanup old entries periodically
	go func() {
		ticker := time.NewTicker(limiter.cleanupTick)
		defer ticker.Stop()
		for range ticker.C {
			limiter.Cleanup()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := limiter.GetLimiter(ip)

		if !limiter.Allow() {
			retryAfter := time.Second // Suggest retry after 1 second
			c.Header("Retry-After", retryAfter.String())
			c.Header("X-RateLimit-Limit", "200")
			c.Header("X-RateLimit-Remaining", "0")

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded. Please try again later.",
				"retry_after": retryAfter.String(),
			})
			c.Abort()
			return
		}

		// Add rate limit headers to response
		c.Header("X-RateLimit-Limit", "200")
		c.Header("X-RateLimit-Remaining", "1") // Simplified, actual remaining is internal

		c.Next()
	}
}

