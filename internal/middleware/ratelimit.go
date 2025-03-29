package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// RateLimiterConfig holds rate limiter configuration
type RateLimiterConfig struct {
	Enabled           bool
	RequestsPerMinute int
}

// TokenBucket implements the token bucket algorithm for rate limiting
type TokenBucket struct {
	tokens         float64
	capacity       float64
	refillRate     float64 // tokens per nanosecond
	lastRefillTime time.Time
	clientBuckets  map[string]*TokenBucket
	mu             sync.Mutex
}

// NewTokenBucket creates a new token bucket rate limiter
func NewTokenBucket(requestsPerMinute int) *TokenBucket {
	capacity := float64(requestsPerMinute)
	refillRate := capacity / float64(time.Minute)

	return &TokenBucket{
		tokens:         capacity,
		capacity:       capacity,
		refillRate:     refillRate,
		lastRefillTime: time.Now(),
		clientBuckets:  make(map[string]*TokenBucket),
	}
}

// refill adds tokens to the bucket based on elapsed time
func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefillTime)
	tb.lastRefillTime = now

	tokensToAdd := elapsed.Seconds() * 60 * tb.refillRate
	tb.tokens = min(tb.capacity, tb.tokens+tokensToAdd)
}

// allow checks if a request is allowed based on available tokens
func (tb *TokenBucket) allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}

	return false
}

// getClientBucket gets or creates a token bucket for a specific client
func (tb *TokenBucket) getClientBucket(clientIP string) *TokenBucket {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	bucket, exists := tb.clientBuckets[clientIP]
	if !exists {
		bucket = NewTokenBucket(int(tb.capacity))
		tb.clientBuckets[clientIP] = bucket
	}

	return bucket
}

// RateLimiterMiddleware returns a gin middleware for rate limiting
func RateLimiterMiddleware(config RateLimiterConfig) gin.HandlerFunc {
	if !config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	limiter := NewTokenBucket(config.RequestsPerMinute)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		bucket := limiter.getClientBucket(clientIP)

		if !bucket.allow() {
			log.Debug().
				Str("client_ip", clientIP).
				Int("rate_limit", config.RequestsPerMinute).
				Msg("Rate limit exceeded")

			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}

		c.Next()
	}
}

// Helper function for min of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
