package httpapi

import (
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestWithRedisIgnoresTypedNilClient(t *testing.T) {
	rl := NewRateLimiter(10, 0)
	var typedNil *redis.Client
	// Boxed typed-nil must not enable Redis backend (would panic on allow).
	rl.WithRedis(typedNil)
	if rl.rdb != nil {
		t.Fatal("typed-nil redis.Cmdable must not enable Redis backend")
	}
	if !rl.allow("k") {
		t.Fatal("memory limiter should still allow")
	}
}
