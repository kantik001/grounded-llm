package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	metricHTTPRequests atomic.Uint64
	metricRAGRequests  atomic.Uint64
	metricLLMRequests  atomic.Uint64
	metricCacheHits    atomic.Uint64
	metricCacheMisses  atomic.Uint64

	llmTokenMu sync.Mutex
	llmTokens  = map[string]*llmTokenCounters{}
)

type llmTokenCounters struct {
	input      uint64
	output     uint64
	latencySum float64
	latencyN   uint64
	ttftSum    float64
	ttftN      uint64
}

func metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		metricHTTPRequests.Add(1)
		logRequest(c, "http", map[string]any{
			"path":     c.FullPath(),
			"status":   c.Writer.Status(),
			"duration": time.Since(start).Milliseconds(),
		})
	}
}

func llmMetricKey(tenant, model string) string {
	if tenant == "" {
		tenant = "default"
	}
	if model == "" {
		model = "unknown"
	}
	return tenant + "\x00" + model
}

func recordLLMUsage(tenant, model string, promptTokens, completionTokens int, latency, ttft time.Duration) {
	if promptTokens < 0 {
		promptTokens = 0
	}
	if completionTokens < 0 {
		completionTokens = 0
	}
	key := llmMetricKey(tenant, model)
	llmTokenMu.Lock()
	defer llmTokenMu.Unlock()
	c := llmTokens[key]
	if c == nil {
		c = &llmTokenCounters{}
		llmTokens[key] = c
	}
	c.input += uint64(promptTokens)
	c.output += uint64(completionTokens)
	if latency > 0 {
		c.latencySum += latency.Seconds()
		c.latencyN++
	}
	if ttft > 0 {
		c.ttftSum += ttft.Seconds()
		c.ttftN++
	}
}

func estimateTokens(text string) int {
	n := len(strings.TrimSpace(text))
	if n == 0 {
		return 0
	}
	return (n + 3) / 4
}

func estimateMessagesTokens(messages []Message) int {
	total := 0
	for _, m := range messages {
		total += estimateTokens(m.Content) + 4
	}
	return total
}

func handleMetrics(c *gin.Context) {
	c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	var b strings.Builder
	fmt.Fprintf(&b, "# HELP grounded_http_requests_total Total HTTP requests handled\n")
	fmt.Fprintf(&b, "# TYPE grounded_http_requests_total counter\n")
	fmt.Fprintf(&b, "grounded_http_requests_total %s\n", u64(metricHTTPRequests.Load()))
	fmt.Fprintf(&b, "# HELP grounded_rag_requests_total Total RAG pipeline invocations\n")
	fmt.Fprintf(&b, "# TYPE grounded_rag_requests_total counter\n")
	fmt.Fprintf(&b, "grounded_rag_requests_total %s\n", u64(metricRAGRequests.Load()))
	fmt.Fprintf(&b, "# HELP grounded_llm_requests_total Total LLM completion calls\n")
	fmt.Fprintf(&b, "# TYPE grounded_llm_requests_total counter\n")
	fmt.Fprintf(&b, "grounded_llm_requests_total %s\n", u64(metricLLMRequests.Load()))
	fmt.Fprintf(&b, "# HELP grounded_llm_response_cache_hits_total Semantic LLM response cache hits\n")
	fmt.Fprintf(&b, "# TYPE grounded_llm_response_cache_hits_total counter\n")
	fmt.Fprintf(&b, "grounded_llm_response_cache_hits_total %s\n", u64(metricCacheHits.Load()))
	fmt.Fprintf(&b, "# HELP grounded_llm_response_cache_misses_total Semantic LLM response cache misses\n")
	fmt.Fprintf(&b, "# TYPE grounded_llm_response_cache_misses_total counter\n")
	fmt.Fprintf(&b, "grounded_llm_response_cache_misses_total %s\n", u64(metricCacheMisses.Load()))

	fmt.Fprintf(&b, "# HELP llm_tokens_input_total Estimated or reported LLM prompt tokens\n")
	fmt.Fprintf(&b, "# TYPE llm_tokens_input_total counter\n")
	fmt.Fprintf(&b, "# HELP llm_tokens_output_total Estimated or reported LLM completion tokens\n")
	fmt.Fprintf(&b, "# TYPE llm_tokens_output_total counter\n")
	fmt.Fprintf(&b, "# HELP llm_latency_seconds LLM completion latency (sum/count)\n")
	fmt.Fprintf(&b, "# TYPE llm_latency_seconds summary\n")
	fmt.Fprintf(&b, "# HELP llm_ttft_seconds Time to first token on streaming completions\n")
	fmt.Fprintf(&b, "# TYPE llm_ttft_seconds summary\n")

	llmTokenMu.Lock()
	for key, ctr := range llmTokens {
		parts := strings.SplitN(key, "\x00", 2)
		tenant, model := "default", "unknown"
		if len(parts) == 2 {
			tenant, model = parts[0], parts[1]
		}
		labels := fmt.Sprintf(`tenant=%q,model=%q`, tenant, model)
		fmt.Fprintf(&b, "llm_tokens_input_total{%s} %d\n", labels, ctr.input)
		fmt.Fprintf(&b, "llm_tokens_output_total{%s} %d\n", labels, ctr.output)
		fmt.Fprintf(&b, "llm_latency_seconds_sum{%s} %s\n", labels, strconv.FormatFloat(ctr.latencySum, 'f', 6, 64))
		fmt.Fprintf(&b, "llm_latency_seconds_count{%s} %d\n", labels, ctr.latencyN)
		fmt.Fprintf(&b, "llm_ttft_seconds_sum{%s} %s\n", labels, strconv.FormatFloat(ctr.ttftSum, 'f', 6, 64))
		fmt.Fprintf(&b, "llm_ttft_seconds_count{%s} %d\n", labels, ctr.ttftN)
	}
	llmTokenMu.Unlock()

	if redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if n, err := redisClient.Get(ctx, "rag_embedding_cache_hit_total").Int64(); err == nil {
			fmt.Fprintf(&b, "# HELP rag_embedding_cache_hit_total Embedding vector cache hits (Redis)\n")
			fmt.Fprintf(&b, "# TYPE rag_embedding_cache_hit_total counter\n")
			fmt.Fprintf(&b, "rag_embedding_cache_hit_total %d\n", n)
		}
	}

	c.String(http.StatusOK, b.String())
}

func u64(v uint64) string {
	return strconv.FormatUint(v, 10)
}
