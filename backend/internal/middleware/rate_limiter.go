package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	requests map[string]*clientRequests
	mu       sync.RWMutex
	limit    int           // max requests
	window   time.Duration // time window
}

type clientRequests struct {
	count     int
	resetTime time.Time
}

// NewRateLimiter creates a new rate limiter
// limit: maximum number of requests allowed
// window: time window for the limit (e.g., 1 minute)
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*clientRequests),
		limit:    limit,
		window:   window,
	}

	// Start cleanup goroutine to remove old entries
	go rl.cleanup()

	return rl
}

// cleanup removes expired entries from the rate limiter
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, req := range rl.requests {
			if now.After(req.resetTime) {
				delete(rl.requests, key)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow checks if a request from the given identifier is allowed
func (rl *RateLimiter) Allow(identifier string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	req, exists := rl.requests[identifier]

	if !exists || now.After(req.resetTime) {
		// First request or window has reset
		rl.requests[identifier] = &clientRequests{
			count:     1,
			resetTime: now.Add(rl.window),
		}
		return true
	}

	if req.count >= rl.limit {
		// Rate limit exceeded
		return false
	}

	// Increment count
	req.count++
	return true
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(limit, window)

	return func(c *gin.Context) {
		// Use IP address as identifier
		identifier := c.ClientIP()

		if !limiter.Allow(identifier) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success":    false,
				"error_code": "RATE_LIMIT_EXCEEDED",
				"message":    "Terlalu banyak permintaan. Silakan coba lagi nanti.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuthRateLimitMiddleware creates a rate limiting middleware specifically for authentication endpoints
// More restrictive than general rate limiting
func AuthRateLimitMiddleware() gin.HandlerFunc {
	// Allow 5 login attempts per minute per IP
	return RateLimitMiddleware(5, time.Minute)
}

// APIRateLimitMiddleware creates a rate limiting middleware for general API endpoints
func APIRateLimitMiddleware() gin.HandlerFunc {
	// Allow 100 requests per minute per IP
	return RateLimitMiddleware(100, time.Minute)
}

// PerUserRateLimiter implements rate limiting per user ID
type PerUserRateLimiter struct {
	requests map[uint]*clientRequests
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

// NewPerUserRateLimiter creates a new per-user rate limiter
func NewPerUserRateLimiter(limit int, window time.Duration) *PerUserRateLimiter {
	rl := &PerUserRateLimiter{
		requests: make(map[uint]*clientRequests),
		limit:    limit,
		window:   window,
	}

	go rl.cleanup()

	return rl
}

// cleanup removes expired entries
func (rl *PerUserRateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, req := range rl.requests {
			if now.After(req.resetTime) {
				delete(rl.requests, key)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow checks if a request from the given user is allowed
func (rl *PerUserRateLimiter) Allow(userID uint) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	req, exists := rl.requests[userID]

	if !exists || now.After(req.resetTime) {
		rl.requests[userID] = &clientRequests{
			count:     1,
			resetTime: now.Add(rl.window),
		}
		return true
	}

	if req.count >= rl.limit {
		return false
	}

	req.count++
	return true
}

// PerUserRateLimitMiddleware creates a rate limiting middleware per authenticated user
func PerUserRateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	limiter := NewPerUserRateLimiter(limit, window)

	return func(c *gin.Context) {
		// Get user ID from context (set by auth middleware)
		userIDInterface, exists := c.Get("user_id")
		if !exists {
			// If no user ID, fall back to IP-based rate limiting
			c.Next()
			return
		}

		userID := userIDInterface.(uint)

		if !limiter.Allow(userID) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success":    false,
				"error_code": "RATE_LIMIT_EXCEEDED",
				"message":    "Terlalu banyak permintaan. Silakan coba lagi nanti.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
