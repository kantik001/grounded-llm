package main

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
)

func writeSSE(c *gin.Context, event, data string) {
	fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, data)
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
		writeSSE(c, "error", `{"error":"Ошибка истории"}`)
		return
	}

	prepared, err := prepareRAGMessages(text, domainID, tenantID, prior, sid)
	if err != nil {
		writeSSE(c, "error", mustJSON(gin.H{"error": err.Error()}))
		return
	}
	if prepared.SoftFail || !prepared.OK {
		writeSSE(c, "error", mustJSON(gin.H{"error": prepared.ErrMsg, "soft_fail": prepared.SoftFail}))
		return
	}

	if _, err := chatStore.AppendMessage(ctx, sid, ChatMessage{Role: "user", Content: text, Kind: "text"}); err != nil {
		writeSSE(c, "error", `{"error":"Ошибка сохранения"}`)
		return
	}

	raw, err := callLLMCompletionStream(ctx, prepared.LLMMessages, func(delta string) error {
		writeSSE(c, "token", mustJSON(gin.H{"text": delta}))
		return nil
	})
	if err != nil {
		writeSSE(c, "error", mustJSON(gin.H{"error": publicAPIError(err)}))
		return
	}

	result := finalizeRAGAnswer(raw, prepared)
	if _, err := chatStore.AppendMessage(ctx, sid, ChatMessage{
		Role: "assistant", Content: result.Answer, Kind: "assistant", Citations: result.Citations,
	}); err != nil {
		writeSSE(c, "error", `{"error":"Ошибка сохранения ответа"}`)
		return
	}

	msgs, err := chatStore.ListMessages(ctx, sid, telegramID)
	if err != nil {
		writeSSE(c, "error", `{"error":"Ошибка базы данных"}`)
		return
	}
	writeSSE(c, "done", mustJSON(gin.H{
		"success":    true,
		"session_id": sid,
		"domain_id":  domainID,
		"tenant_id":  tenantID,
		"messages":   msgs,
	}))
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
