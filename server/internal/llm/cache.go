package llm

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"grounded_llm_server/internal/store"
)

var redisClient *redis.Client

// CachedAnswer is a semantic response-cache payload.
type CachedAnswer struct {
	Answer    string              `json:"answer"`
	Citations []store.RAGFragment `json:"citations,omitempty"`
}

// InitRedis connects REDIS_URL for the LLM response cache (optional).
func InitRedis() {
	url := strings.TrimSpace(os.Getenv("REDIS_URL"))
	if url == "" {
		return
	}
	opt, err := redis.ParseURL(url)
	if err != nil {
		log.Printf("REDIS_URL parse error: %v", err)
		return
	}
	c := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := c.Ping(ctx).Err(); err != nil {
		log.Printf("Redis unavailable (LLM response cache disabled): %v", err)
		_ = c.Close()
		return
	}
	redisClient = c
	log.Printf("Redis connected (LLM semantic response cache enabled)")
}

// Redis returns the shared Redis client (may be nil).
func Redis() *redis.Client {
	return redisClient
}

func responseCacheTTL() time.Duration {
	sec, _ := strconv.Atoi(strings.TrimSpace(os.Getenv("LLM_RESPONSE_CACHE_TTL_SEC")))
	if sec <= 0 {
		sec = 86400
	}
	return time.Duration(sec) * time.Second
}

// ResponseCacheKey: response:{md5(query+domain[+tenant])}:{model}
func ResponseCacheKey(query, domainID, tenantID, model string) string {
	payload := strings.TrimSpace(query) + "|" + domainID
	if tenantID != "" {
		payload += "|" + tenantID
	}
	h := md5.Sum([]byte(payload))
	return "response:" + hex.EncodeToString(h[:]) + ":" + model
}

// GetCachedAnswer looks up a semantic LLM response.
func GetCachedAnswer(ctx context.Context, query, domainID, tenantID string) (*CachedAnswer, bool) {
	c := cfg()
	if redisClient == nil || c == nil {
		return nil, false
	}
	key := ResponseCacheKey(query, domainID, tenantID, c.LLMModel)
	raw, err := redisClient.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false
	}
	var out CachedAnswer
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, false
	}
	return &out, true
}

// SetCachedAnswer stores a verified answer in the semantic cache.
func SetCachedAnswer(ctx context.Context, query, domainID, tenantID string, answer string, citations []store.RAGFragment) {
	c := cfg()
	if redisClient == nil || c == nil || strings.TrimSpace(answer) == "" {
		return
	}
	key := ResponseCacheKey(query, domainID, tenantID, c.LLMModel)
	payload, err := json.Marshal(CachedAnswer{Answer: answer, Citations: citations})
	if err != nil {
		return
	}
	if err := redisClient.Set(ctx, key, payload, responseCacheTTL()).Err(); err != nil {
		log.Printf("LLM response cache set failed: %v", err)
	}
}
