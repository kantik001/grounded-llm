// Package metrics holds process-wide counters shared by llm, rag, and HTTP scrape.
package metrics

import (
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	HTTPRequests atomic.Uint64
	RAGRequests  atomic.Uint64
	LLMRequests  atomic.Uint64
	CacheHits    atomic.Uint64
	CacheMisses  atomic.Uint64
)

// TokenCounter is per-tenant/model LLM usage for Prometheus scrape.
type TokenCounter struct {
	Input      int64
	Output     int64
	LatencySum float64
	LatencyN   int64
	TtftSum    float64
	TtftN      int64
}

var (
	llmTokenMu sync.Mutex
	llmTokens  = map[string]*TokenCounter{}
)

func RecordLLMUsage(tenantID, model string, promptTokens, completionTokens int, latency, ttft time.Duration) {
	if promptTokens < 0 {
		promptTokens = 0
	}
	if completionTokens < 0 {
		completionTokens = 0
	}
	if tenantID == "" {
		tenantID = "default"
	}
	if model == "" {
		model = "unknown"
	}
	key := tenantID + "\x00" + model
	llmTokenMu.Lock()
	defer llmTokenMu.Unlock()
	c := llmTokens[key]
	if c == nil {
		c = &TokenCounter{}
		llmTokens[key] = c
	}
	c.Input += int64(promptTokens)
	c.Output += int64(completionTokens)
	if latency > 0 {
		c.LatencySum += latency.Seconds()
		c.LatencyN++
	}
	if ttft > 0 {
		c.TtftSum += ttft.Seconds()
		c.TtftN++
	}
}

func EstimateTokens(text string) int {
	n := len(strings.TrimSpace(text))
	if n == 0 {
		return 0
	}
	return (n + 3) / 4
}

func EstimateMessagesTokens(contents []string) int {
	total := 0
	for _, content := range contents {
		total += EstimateTokens(content) + 4
	}
	return total
}

// SnapshotLLMTokens copies token counters for Prometheus scrape.
func SnapshotLLMTokens() map[string]TokenCounter {
	llmTokenMu.Lock()
	defer llmTokenMu.Unlock()
	out := make(map[string]TokenCounter, len(llmTokens))
	for k, v := range llmTokens {
		out[k] = *v
	}
	return out
}

func FormatUint64(v uint64) string {
	return strconv.FormatUint(v, 10)
}
