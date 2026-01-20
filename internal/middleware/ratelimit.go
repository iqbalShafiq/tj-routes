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
		ips:         make(map[string]*IPEntry),
		mu:          &sync.RWMutex{},
		r:           rate.Limit(rps),
		b:           burst,
		entryTTL:    15 * time.Minute,
		maxEntries:  10000,
		cleanupTick: 5 * time.Minute,
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

	for ip, entry := range i.ips {
		if now.Sub(entry.lastAccess) > i.entryTTL {
			expired = append(expired, ip)
		}
	}

	for _, ip := range expired {
		delete(i.ips, ip)
	}

	if len(i.ips) > i.maxEntries {
		type ipTime struct {
			ip         string
			lastAccess time.Time
		}
		entries := make([]ipTime, 0, len(i.ips))
		for ip, entry := range i.ips {
			entries = append(entries, ipTime{ip, entry.lastAccess})
		}

		for j := 0; j < len(entries)-i.maxEntries; j++ {
			for k := j + 1; k < len(entries); k++ {
				if entries[j].lastAccess.After(entries[k].lastAccess) {
					entries[j], entries[k] = entries[k], entries[j]
				}
			}
		}

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
		"total_entries":   len(i.ips),
		"active_entries":  activeCount,
		"expired_entries": expiredCount,
		"max_entries":     i.maxEntries,
		"entry_ttl":       i.entryTTL.String(),
	}
}

var (
	defaultLimiter          = NewIPRateLimiter(100, 200)
	chatMessageLimiter      *IPRateLimiter
	chatConversationLimiter *IPRateLimiter
	chatGroupLimiter        *IPRateLimiter
	chatLimitersInitialized bool
)

// InitChatRateLimiters initializes chat-specific rate limiters with the given limits
func InitChatRateLimiters(messagesPerMinute, conversationsPerMinute, groupsPerMinute int) {
	chatMessageLimiter = NewIPRateLimiter(messagesPerMinute/60, messagesPerMinute/60)
	chatConversationLimiter = NewIPRateLimiter(conversationsPerMinute/60, conversationsPerMinute/60)
	chatGroupLimiter = NewIPRateLimiter(groupsPerMinute/60, groupsPerMinute/60)
	chatLimitersInitialized = true

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if chatMessageLimiter != nil {
				chatMessageLimiter.Cleanup()
			}
			if chatConversationLimiter != nil {
				chatConversationLimiter.Cleanup()
			}
			if chatGroupLimiter != nil {
				chatGroupLimiter.Cleanup()
			}
		}
	}()
}

// ChatRateLimitMiddleware is a rate limiting middleware for chat endpoints
func ChatRateLimitMiddleware(limitType string) gin.HandlerFunc {
	var limiter *IPRateLimiter

	if !chatLimitersInitialized {
		limiter = defaultLimiter
	} else {
		switch limitType {
		case "messages":
			limiter = chatMessageLimiter
		case "conversations":
			limiter = chatConversationLimiter
		case "groups":
			limiter = chatGroupLimiter
		default:
			limiter = defaultLimiter
		}
	}

	if limiter == nil || limiter.r == 0 {
		limiter = defaultLimiter
	}

	return func(c *gin.Context) {
		ip := c.ClientIP()
		rateLimiter := limiter.GetLimiter(ip)

		if !rateLimiter.Allow() {
			retryAfter := time.Minute
			c.Header("Retry-After", retryAfter.String())
			c.Header("X-RateLimit-Limit", "30")
			c.Header("X-RateLimit-Remaining", "0")

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Chat rate limit exceeded. Please try again later.",
				"retry_after": retryAfter.String(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(rps int, burst int) gin.HandlerFunc {
	var limiter *IPRateLimiter
	if rps > 0 && burst > 0 {
		limiter = NewIPRateLimiter(rps, burst)
	} else {
		limiter = defaultLimiter
	}

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
			retryAfter := time.Second
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

		c.Header("X-RateLimit-Limit", "200")
		c.Header("X-RateLimit-Remaining", "1")

		c.Next()
	}
}
