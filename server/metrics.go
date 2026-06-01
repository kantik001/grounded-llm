package main

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	metricHTTPRequests atomic.Uint64
	metricRAGRequests  atomic.Uint64
	metricLLMRequests  atomic.Uint64
)

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

func handleMetrics(c *gin.Context) {
	c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	body := "# HELP grounded_http_requests_total Total HTTP requests handled\n" +
		"# TYPE grounded_http_requests_total counter\n" +
		"grounded_http_requests_total " + u64(metricHTTPRequests.Load()) + "\n" +
		"# HELP grounded_rag_requests_total Total RAG pipeline invocations\n" +
		"# TYPE grounded_rag_requests_total counter\n" +
		"grounded_rag_requests_total " + u64(metricRAGRequests.Load()) + "\n" +
		"# HELP grounded_llm_requests_total Total LLM completion calls\n" +
		"# TYPE grounded_llm_requests_total counter\n" +
		"grounded_llm_requests_total " + u64(metricLLMRequests.Load()) + "\n"
	c.String(http.StatusOK, body)
}

func u64(v uint64) string {
	if v == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}
