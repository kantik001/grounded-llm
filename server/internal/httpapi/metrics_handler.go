package httpapi

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/config"
	"grounded_llm_server/internal/llm"
	"grounded_llm_server/internal/metrics"
)

// MetricsMiddleware increments HTTP request counters and logs the request.
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		metrics.HTTPRequests.Add(1)
		if svc != nil {
			svc.LogRequest(c, "http", map[string]any{
				"path":     c.FullPath(),
				"status":   c.Writer.Status(),
				"duration": time.Since(start).Milliseconds(),
			})
		}
	}
}

// metricsAuthorized gates /metrics: if METRICS_TOKEN is set, a matching
// Bearer token is required. In production an unset token closes the endpoint
// (metrics expose per-tenant/model counters).
func metricsAuthorized(c *gin.Context) bool {
	token := strings.TrimSpace(os.Getenv("METRICS_TOKEN"))
	if token == "" {
		return !config.IsProductionEnv()
	}
	authz := strings.TrimSpace(c.GetHeader("Authorization"))
	presented, ok := strings.CutPrefix(authz, "Bearer ")
	if !ok {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(strings.TrimSpace(presented)), []byte(token)) == 1
}

// Metrics scrapes Prometheus text exposition.
func Metrics(c *gin.Context) {
	if !metricsAuthorized(c) {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "metrics access requires METRICS_TOKEN bearer auth"})
		return
	}
	c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	var b strings.Builder
	fmt.Fprintf(&b, "# HELP grounded_http_requests_total Total HTTP requests handled\n")
	fmt.Fprintf(&b, "# TYPE grounded_http_requests_total counter\n")
	fmt.Fprintf(&b, "grounded_http_requests_total %s\n", metrics.FormatUint64(metrics.HTTPRequests.Load()))
	fmt.Fprintf(&b, "# HELP grounded_rag_requests_total Total RAG pipeline invocations\n")
	fmt.Fprintf(&b, "# TYPE grounded_rag_requests_total counter\n")
	fmt.Fprintf(&b, "grounded_rag_requests_total %s\n", metrics.FormatUint64(metrics.RAGRequests.Load()))
	fmt.Fprintf(&b, "# HELP grounded_llm_requests_total Total LLM completion calls\n")
	fmt.Fprintf(&b, "# TYPE grounded_llm_requests_total counter\n")
	fmt.Fprintf(&b, "grounded_llm_requests_total %s\n", metrics.FormatUint64(metrics.LLMRequests.Load()))
	fmt.Fprintf(&b, "# HELP grounded_llm_response_cache_hits_total Semantic LLM response cache hits\n")
	fmt.Fprintf(&b, "# TYPE grounded_llm_response_cache_hits_total counter\n")
	fmt.Fprintf(&b, "grounded_llm_response_cache_hits_total %s\n", metrics.FormatUint64(metrics.CacheHits.Load()))
	fmt.Fprintf(&b, "# HELP grounded_llm_response_cache_misses_total Semantic LLM response cache misses\n")
	fmt.Fprintf(&b, "# TYPE grounded_llm_response_cache_misses_total counter\n")
	fmt.Fprintf(&b, "grounded_llm_response_cache_misses_total %s\n", metrics.FormatUint64(metrics.CacheMisses.Load()))

	fmt.Fprintf(&b, "# HELP llm_tokens_input_total Estimated or reported LLM prompt tokens\n")
	fmt.Fprintf(&b, "# TYPE llm_tokens_input_total counter\n")
	fmt.Fprintf(&b, "# HELP llm_tokens_output_total Estimated or reported LLM completion tokens\n")
	fmt.Fprintf(&b, "# TYPE llm_tokens_output_total counter\n")
	fmt.Fprintf(&b, "# HELP llm_latency_seconds LLM completion latency (sum/count)\n")
	fmt.Fprintf(&b, "# TYPE llm_latency_seconds summary\n")
	fmt.Fprintf(&b, "# HELP llm_ttft_seconds Time to first token on streaming completions\n")
	fmt.Fprintf(&b, "# TYPE llm_ttft_seconds summary\n")

	for key, ctr := range metrics.SnapshotLLMTokens() {
		parts := strings.SplitN(key, "\x00", 2)
		tenant, model := "default", "unknown"
		if len(parts) == 2 {
			tenant, model = parts[0], parts[1]
		}
		labels := fmt.Sprintf(`tenant=%q,model=%q`, tenant, model)
		fmt.Fprintf(&b, "llm_tokens_input_total{%s} %d\n", labels, ctr.Input)
		fmt.Fprintf(&b, "llm_tokens_output_total{%s} %d\n", labels, ctr.Output)
		fmt.Fprintf(&b, "llm_latency_seconds_sum{%s} %s\n", labels, strconv.FormatFloat(ctr.LatencySum, 'f', 6, 64))
		fmt.Fprintf(&b, "llm_latency_seconds_count{%s} %d\n", labels, ctr.LatencyN)
		fmt.Fprintf(&b, "llm_ttft_seconds_sum{%s} %s\n", labels, strconv.FormatFloat(ctr.TtftSum, 'f', 6, 64))
		fmt.Fprintf(&b, "llm_ttft_seconds_count{%s} %d\n", labels, ctr.TtftN)
	}

	if redis := llm.Redis(); redis != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if n, err := redis.Get(ctx, "rag_embedding_cache_hit_total").Int64(); err == nil {
			fmt.Fprintf(&b, "# HELP rag_embedding_cache_hit_total Embedding vector cache hits (Redis)\n")
			fmt.Fprintf(&b, "# TYPE rag_embedding_cache_hit_total counter\n")
			fmt.Fprintf(&b, "rag_embedding_cache_hit_total %d\n", n)
		}
	}

	c.String(http.StatusOK, b.String())
}
