package httpapi

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiter is a sliding-window limiter keyed by caller identity.
// When a Redis client is set (WithRedis), counters are shared across replicas;
// otherwise an in-process map is used.
type RateLimiter struct {
	mu       sync.Mutex
	limit    int
	window   time.Duration
	counters map[string][]time.Time
	rdb      redis.Cmdable
}

// NewRateLimiter creates a limiter (limit<=0 disables enforcement).
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limit:    limit,
		window:   window,
		counters: make(map[string][]time.Time),
	}
}

// WithRedis enables distributed limiting via Redis ZSET sliding window.
// Safe to call with nil (no-op). Returns the receiver for chaining.
func (rl *RateLimiter) WithRedis(c redis.Cmdable) *RateLimiter {
	if rl == nil || c == nil {
		return rl
	}
	// Defend against typed-nil concrete values boxed in the interface.
	if rc, ok := c.(*redis.Client); ok && rc == nil {
		return rl
	}
	rl.rdb = c
	log.Printf("Rate limiter: Redis backend enabled")
	return rl
}

func (rl *RateLimiter) allow(key string) bool {
	if rl.rdb != nil {
		ok, err := rl.allowRedis(key)
		if err == nil {
			return ok
		}
		log.Printf("rate limit redis fallback to memory: %v", err)
	}
	return rl.allowMemory(key)
}

func (rl *RateLimiter) allowMemory(key string) bool {
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

func (rl *RateLimiter) allowRedis(key string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	now := time.Now()
	member := now.UnixNano()
	zkey := "ratelimit:" + key
	cutoff := strconv.FormatInt(now.Add(-rl.window).UnixNano(), 10)
	pipe := rl.rdb.TxPipeline()
	pipe.ZRemRangeByScore(ctx, zkey, "0", cutoff)
	pipe.ZAdd(ctx, zkey, redis.Z{Score: float64(member), Member: member})
	count := pipe.ZCard(ctx, zkey)
	pipe.Expire(ctx, zkey, rl.window+time.Second)
	if _, err := pipe.Exec(ctx); err != nil {
		return false, err
	}
	n, err := count.Result()
	if err != nil {
		return false, err
	}
	if rl.limit > 0 && int(n) > rl.limit {
		_ = rl.rdb.ZRem(ctx, zkey, member).Err()
		return false, nil
	}
	return true, nil
}

// Middleware enforces the rate limit using keyFn.
// Unauthenticated callers are bucketed per client IP so public endpoints
// (signup, checkout) are still protected.
func (rl *RateLimiter) Middleware(keyFn func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rl == nil || rl.limit <= 0 {
			c.Next()
			return
		}
		key := "anon"
		if keyFn != nil {
			key = keyFn(c)
		}
		if key == "" || key == "anon" {
			key = "ip:" + c.ClientIP()
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
