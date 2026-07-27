package main

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
)

var redisClient *redis.Client

type cachedLLMAnswer struct {
	Answer    string        `json:"answer"`
	Citations []RAGFragment `json:"citations,omitempty"`
}

func initRedis() {
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

func responseCacheTTL() time.Duration {
	sec, _ := strconv.Atoi(strings.TrimSpace(os.Getenv("LLM_RESPONSE_CACHE_TTL_SEC")))
	if sec <= 0 {
		sec = 86400
	}
	return time.Duration(sec) * time.Second
}

// responseCacheKey: response:{md5(query+domain[+tenant])}:{model}
func responseCacheKey(query, domainID, tenantID, model string) string {
	payload := strings.TrimSpace(query) + "|" + domainID
	if tenantID != "" {
		payload += "|" + tenantID
	}
	h := md5.Sum([]byte(payload))
	return "response:" + hex.EncodeToString(h[:]) + ":" + model
}

func getCachedLLMAnswer(ctx context.Context, query, domainID, tenantID string) (*cachedLLMAnswer, bool) {
	if redisClient == nil || config == nil {
		return nil, false
	}
	key := responseCacheKey(query, domainID, tenantID, config.LLMModel)
	raw, err := redisClient.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false
	}
	var out cachedLLMAnswer
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, false
	}
	return &out, true
}

func setCachedLLMAnswer(ctx context.Context, query, domainID, tenantID string, answer string, citations []RAGFragment) {
	if redisClient == nil || config == nil || strings.TrimSpace(answer) == "" {
		return
	}
	key := responseCacheKey(query, domainID, tenantID, config.LLMModel)
	payload, err := json.Marshal(cachedLLMAnswer{Answer: answer, Citations: citations})
	if err != nil {
		return
	}
	if err := redisClient.Set(ctx, key, payload, responseCacheTTL()).Err(); err != nil {
		log.Printf("LLM response cache set failed: %v", err)
	}
}
