package httpapi

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/llm"
	"grounded_llm_server/internal/metrics"
	"grounded_llm_server/internal/rag"
	"grounded_llm_server/internal/store"
)

func writeSSE(c *gin.Context, event, data string) {
	_, _ = fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, data)
	c.Writer.Flush()
}

func sseMessage(c *gin.Context, sid, domainID, tenantID string, telegramID int64, text string) {
	s := requireServices()
	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	ctx := c.Request.Context()
	prior, err := s.Store().HistoryForLLM(ctx, sid, telegramID, tenantID, 0)
	if err != nil {
		writeSSE(c, "error", mustJSON(AttachRequestID(c, gin.H{"error": "History error"})))
		return
	}

	locale := s.Locale(c)

	if len(prior) == 0 {
		if cached, ok := llm.GetCachedAnswer(ctx, text, domainID, tenantID); ok {
			metrics.CacheHits.Add(1)
			c.Header("X-Cache", "HIT")
			if tr := PathTraceFrom(ctx); tr != nil {
				tr.Step("cache", map[string]any{"hit": true})
				tr.Step("done", map[string]any{"ms": tr.ElapsedMS(), "verify_pass": true, "cache": true})
			}
			result := rag.AnswerResult{
				Answer:        cached.Answer,
				Citations:     cached.Citations,
				OK:            true,
				VerifyPass:    true,
				CacheHit:      true,
				FragmentCount: len(cached.Citations),
			}
			if _, err := s.Store().AppendMessage(ctx, sid, store.ChatMessage{Role: "user", Content: text, Kind: "text"}); err != nil {
				writeSSE(c, "error", mustJSON(AttachRequestID(c, gin.H{"error": "Save error"})))
				return
			}
			s.RecordRAGAnalytics(ctx, telegramID, tenantID, domainID, text, result)
			if _, err := s.Store().AppendMessage(ctx, sid, store.ChatMessage{
				Role: "assistant", Content: result.Answer, Kind: "assistant", Citations: result.Citations,
			}); err != nil {
				writeSSE(c, "error", mustJSON(AttachRequestID(c, gin.H{"error": "Failed to save assistant reply"})))
				return
			}
			msgs, err := s.Store().ListMessages(ctx, sid, telegramID, tenantID)
			if err != nil {
				writeSSE(c, "error", mustJSON(AttachRequestID(c, gin.H{"error": "Database error"})))
				return
			}
			writeSSE(c, "token", mustJSON(gin.H{"text": result.Answer}))
			writeSSE(c, "done", mustJSON(AttachRequestID(c, gin.H{
				"success":    true,
				"session_id": sid,
				"domain_id":  domainID,
				"tenant_id":  tenantID,
				"messages":   msgs,
				"cache":      "HIT",
			})))
			return
		}
		metrics.CacheMisses.Add(1)
		if tr := PathTraceFrom(ctx); tr != nil {
			tr.Step("cache", map[string]any{"hit": false})
		}
	}
	c.Header("X-Cache", "MISS")

	prepared, err := rag.PrepareMessages(ctx, text, domainID, tenantID, locale, prior, sid)
	if err != nil {
		writeSSE(c, "error", mustJSON(AttachRequestID(c, gin.H{"error": s.PublicAPIError(err)})))
		return
	}
	if prepared.SoftFail || !prepared.OK {
		if tr := PathTraceFrom(ctx); tr != nil {
			tr.Step("done", map[string]any{"ms": tr.ElapsedMS(), "ok": false, "soft_fail": prepared.SoftFail})
		}
		s.RecordRAGAnalytics(ctx, telegramID, tenantID, domainID, text, rag.AnswerResult{
			ErrMsg:        prepared.ErrMsg,
			SoftFail:      prepared.SoftFail,
			FragmentCount: len(prepared.Fragments),
		})
		writeSSE(c, "error", mustJSON(AttachRequestID(c, gin.H{"error": prepared.ErrMsg, "soft_fail": prepared.SoftFail})))
		return
	}

	if _, err := s.Store().AppendMessage(ctx, sid, store.ChatMessage{Role: "user", Content: text, Kind: "text"}); err != nil {
		writeSSE(c, "error", mustJSON(AttachRequestID(c, gin.H{"error": "Save error"})))
		return
	}

	llmStart := time.Now()
	raw, err := llm.CompleteStreamForTenant(ctx, tenantID, prepared.LLMMessages, func(delta string) error {
		writeSSE(c, "token", mustJSON(gin.H{"text": delta}))
		return nil
	})
	if tr := PathTraceFrom(ctx); tr != nil {
		tr.Step("llm", map[string]any{"ms": MSSince(llmStart), "ok": err == nil, "stream": true})
	}
	if err != nil {
		if tr := PathTraceFrom(ctx); tr != nil {
			tr.Step("done", map[string]any{"ms": tr.ElapsedMS(), "ok": false})
		}
		writeSSE(c, "error", mustJSON(AttachRequestID(c, gin.H{"error": s.PublicAPIError(err)})))
		return
	}

	result := rag.FinalizeAnswer(ctx, raw, prepared)
	if result.OK && result.VerifyPass && len(prior) == 0 {
		llm.SetCachedAnswer(ctx, text, domainID, tenantID, result.Answer, result.Citations)
	}
	s.RecordRAGAnalytics(ctx, telegramID, tenantID, domainID, text, result)
	if _, err := s.Store().AppendMessage(ctx, sid, store.ChatMessage{
		Role: "assistant", Content: result.Answer, Kind: "assistant", Citations: result.Citations,
	}); err != nil {
		writeSSE(c, "error", mustJSON(AttachRequestID(c, gin.H{"error": "Failed to save assistant reply"})))
		return
	}

	msgs, err := s.Store().ListMessages(ctx, sid, telegramID, tenantID)
	if err != nil {
		writeSSE(c, "error", mustJSON(AttachRequestID(c, gin.H{"error": "Database error"})))
		return
	}
	if tr := PathTraceFrom(ctx); tr != nil {
		tr.Step("done", map[string]any{
			"ms": tr.ElapsedMS(), "ok": result.OK, "verify_pass": result.VerifyPass,
		})
	}
	writeSSE(c, "done", mustJSON(AttachRequestID(c, gin.H{
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

// WantsStream reports whether the client requested SSE.
func WantsStream(c *gin.Context) bool {
	if c.Query("stream") == "1" || c.Query("stream") == "true" {
		return true
	}
	accept := c.GetHeader("Accept")
	return accept == "text/event-stream" || accept == "text/event-stream, */*"
}
