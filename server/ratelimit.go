package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	mu       sync.Mutex
	limit    int
	window   time.Duration
	counters map[string][]time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		limit:    limit,
		window:   window,
		counters: make(map[string][]time.Time),
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)
	times := rl.counters[key]
	var kept []time.Time
	for _, ts := range times {
		if ts.After(cutoff) {
			kept = append(kept, ts)
		}
	}
	if rl.limit > 0 && len(kept) >= rl.limit {
		rl.counters[key] = kept
		return false
	}
	kept = append(kept, now)
	rl.counters[key] = kept
	return true
}

func rateLimitMiddleware(rl *rateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rl == nil || rl.limit <= 0 {
			c.Next()
			return
		}
		key := rateLimitKey(c)
		if key == "anon" {
			c.Next()
			return
		}
		if !rl.allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "Too many requests. Wait a minute and try again.",
			})
			return
		}
		c.Next()
	}
}
