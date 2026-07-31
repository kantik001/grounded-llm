package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func writeSSE(c *gin.Context, event, data string) {
	_, _ = fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, data)
	c.Writer.Flush()
}

func sseMessageHandler(c *gin.Context, sid, domainID, tenantID string, telegramID int64, text string) {
	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	ctx := c.Request.Context()
	prior, err := chatStore.HistoryForLLM(ctx, sid, telegramID, 0)
	if err != nil {
		writeSSE(c, "error", mustJSON(attachRequestID(c, gin.H{"error": "History error"})))
		return
	}

	locale := ctxLocale(c)

	if len(prior) == 0 {
		if cached, ok := getCachedLLMAnswer(ctx, text, domainID, tenantID); ok {
			metricCacheHits.Add(1)
			c.Header("X-Cache", "HIT")
			if tr := pathTraceFrom(ctx); tr != nil {
				tr.step("cache", map[string]any{"hit": true})
				tr.step("done", map[string]any{"ms": tr.elapsedMS(), "verify_pass": true, "cache": true})
			}
			result := RAGAnswerResult{
				Answer:        cached.Answer,
				Citations:     cached.Citations,
				OK:            true,
				VerifyPass:    true,
				CacheHit:      true,
				FragmentCount: len(cached.Citations),
			}
			if _, err := chatStore.AppendMessage(ctx, sid, ChatMessage{Role: "user", Content: text, Kind: "text"}); err != nil {
				writeSSE(c, "error", mustJSON(attachRequestID(c, gin.H{"error": "Save error"})))
				return
			}
			recordRAGAnalytics(ctx, telegramID, tenantID, domainID, text, result)
			if _, err := chatStore.AppendMessage(ctx, sid, ChatMessage{
				Role: "assistant", Content: result.Answer, Kind: "assistant", Citations: result.Citations,
			}); err != nil {
				writeSSE(c, "error", mustJSON(attachRequestID(c, gin.H{"error": "Failed to save assistant reply"})))
				return
			}
			msgs, err := chatStore.ListMessages(ctx, sid, telegramID)
			if err != nil {
				writeSSE(c, "error", mustJSON(attachRequestID(c, gin.H{"error": "Database error"})))
				return
			}
			writeSSE(c, "token", mustJSON(gin.H{"text": result.Answer}))
			writeSSE(c, "done", mustJSON(attachRequestID(c, gin.H{
				"success":    true,
				"session_id": sid,
				"domain_id":  domainID,
				"tenant_id":  tenantID,
				"messages":   msgs,
				"cache":      "HIT",
			})))
			return
		}
		metricCacheMisses.Add(1)
		if tr := pathTraceFrom(ctx); tr != nil {
			tr.step("cache", map[string]any{"hit": false})
		}
	}
	c.Header("X-Cache", "MISS")

	prepared, err := prepareRAGMessages(ctx, text, domainID, tenantID, locale, prior, sid)
	if err != nil {
		writeSSE(c, "error", mustJSON(attachRequestID(c, gin.H{"error": err.Error()})))
		return
	}
	if prepared.SoftFail || !prepared.OK {
		if tr := pathTraceFrom(ctx); tr != nil {
			tr.step("done", map[string]any{"ms": tr.elapsedMS(), "ok": false, "soft_fail": prepared.SoftFail})
		}
		recordRAGAnalytics(ctx, telegramID, tenantID, domainID, text, RAGAnswerResult{
			ErrMsg:        prepared.ErrMsg,
			SoftFail:      prepared.SoftFail,
			FragmentCount: len(prepared.Fragments),
		})
		writeSSE(c, "error", mustJSON(attachRequestID(c, gin.H{"error": prepared.ErrMsg, "soft_fail": prepared.SoftFail})))
		return
	}

	if _, err := chatStore.AppendMessage(ctx, sid, ChatMessage{Role: "user", Content: text, Kind: "text"}); err != nil {
		writeSSE(c, "error", mustJSON(attachRequestID(c, gin.H{"error": "Save error"})))
		return
	}

	llmStart := time.Now()
	raw, err := callLLMCompletionStreamForTenant(ctx, tenantID, prepared.LLMMessages, func(delta string) error {
		writeSSE(c, "token", mustJSON(gin.H{"text": delta}))
		return nil
	})
	if tr := pathTraceFrom(ctx); tr != nil {
		tr.step("llm", map[string]any{"ms": msSince(llmStart), "ok": err == nil, "stream": true})
	}
	if err != nil {
		if tr := pathTraceFrom(ctx); tr != nil {
			tr.step("done", map[string]any{"ms": tr.elapsedMS(), "ok": false})
		}
		writeSSE(c, "error", mustJSON(attachRequestID(c, gin.H{"error": publicAPIError(err)})))
		return
	}

	result := finalizeRAGAnswer(ctx, raw, prepared)
	if result.OK && result.VerifyPass && len(prior) == 0 {
		setCachedLLMAnswer(ctx, text, domainID, tenantID, result.Answer, result.Citations)
	}
	recordRAGAnalytics(ctx, telegramID, tenantID, domainID, text, result)
	if _, err := chatStore.AppendMessage(ctx, sid, ChatMessage{
		Role: "assistant", Content: result.Answer, Kind: "assistant", Citations: result.Citations,
	}); err != nil {
		writeSSE(c, "error", mustJSON(attachRequestID(c, gin.H{"error": "Failed to save assistant reply"})))
		return
	}

	msgs, err := chatStore.ListMessages(ctx, sid, telegramID)
	if err != nil {
		writeSSE(c, "error", mustJSON(attachRequestID(c, gin.H{"error": "Database error"})))
		return
	}
	if tr := pathTraceFrom(ctx); tr != nil {
		tr.step("done", map[string]any{
			"ms": tr.elapsedMS(), "ok": result.OK, "verify_pass": result.VerifyPass,
		})
	}
	writeSSE(c, "done", mustJSON(attachRequestID(c, gin.H{
		"success":    true,
		"session_id": sid,
		"domain_id":  domainID,
		"tenant_id":  tenantID,
		"messages":   msgs,
	})))
}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return `{"error":"json"}`
	}
	return string(b)
}

func wantsStream(c *gin.Context) bool {
	if c.Query("stream") == "1" || c.Query("stream") == "true" {
		return true
	}
	accept := c.GetHeader("Accept")
	return accept == "text/event-stream" || accept == "text/event-stream, */*"
}
