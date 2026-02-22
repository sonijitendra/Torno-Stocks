package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"tinystock/backend/internal/response"
)

// RateLimiter holds per-IP rate limiters
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter creates a rate limiter (requests per second, burst)
func NewRateLimiter(rps int, burst int) *RateLimiter {
	if burst < rps {
		burst = rps
	}
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(float64(rps) / 60), // per minute
		burst:    burst,
	}
}

// getLimiter returns limiter for IP
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.RLock()
	l, ok := rl.limiters[ip]
	rl.mu.RUnlock()
	if ok {
		return l
	}
	rl.mu.Lock()
	defer rl.mu.Unlock()
	l = rate.NewLimiter(rl.rate, rl.burst)
	rl.limiters[ip] = l
	return l
}

// Middleware returns a rate limiting middleware
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := rl.getLimiter(ip)
		if !limiter.Allow() {
			response.ErrorResponse(c, http.StatusTooManyRequests, "RATE_LIMIT", "Too many requests")
			c.Abort()
			return
		}
		c.Next()
	}
}
