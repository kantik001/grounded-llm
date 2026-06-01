package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// rateLimiter — простой in-memory лимитер (для одного инстанса Go).
// На продакшене с несколькими репликами заменим на Redis (фаза 1B).
type rateLimiter struct {
	mu       sync.Mutex
	limit    int
	window   time.Duration
	counters map[int64][]time.Time
}

// Создаёт in-memory лимитер запросов на пользователя за окно времени.
func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		limit:    limit,
		window:   window,
		counters: make(map[int64][]time.Time),
	}
}

// Проверяет, не превышен ли лимит для telegram user id.
func (rl *rateLimiter) allow(userID int64) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)
	times := rl.counters[userID]
	var kept []time.Time
	for _, ts := range times {
		if ts.After(cutoff) {
			kept = append(kept, ts)
		}
	}
	if len(kept) >= rl.limit {
		rl.counters[userID] = kept
		return false
	}
	kept = append(kept, now)
	rl.counters[userID] = kept
	return true
}

// Gin-middleware: 429 при превышении лимита запросов в минуту.
func rateLimitMiddleware(rl *rateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rl == nil || rl.limit <= 0 {
			c.Next()
			return
		}
		rawID, exists := c.Get("telegram_user_id")
		if !exists {
			c.Next()
			return
		}
		userID, ok := rawID.(int64)
		if !ok || userID == 0 {
			c.Next()
			return
		}
		if !rl.allow(userID) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "Слишком много запросов. Подождите минуту и попробуйте снова.",
			})
			return
		}
		c.Next()
	}
}
