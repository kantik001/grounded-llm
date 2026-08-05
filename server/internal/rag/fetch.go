package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"grounded_llm_server/internal/store"
)

// ContextResponse — ответ POST /rag/context.
type ContextResponse struct {
	Success   bool                `json:"success"`
	Error     string              `json:"error,omitempty"`
	Context   string              `json:"context,omitempty"`
	FewShot   string              `json:"few_shot,omitempty"`
	Category  string              `json:"category,omitempty"`
	Fragments []store.RAGFragment `json:"fragments,omitempty"`
}

// FetchContext loads retrieval context from Python RAG (or mock).
func FetchContext(ctx context.Context, question, tenantID, domainID, locale string) (*ContextResponse, error) {
	if MockEnabled() {
		out := MockContextResponse(question, domainID)
		traceStep(ctx, "retrieve", map[string]any{
			"ms": 0, "fragments": len(out.Fragments), "ok": out.Success, "mock": true,
		})
		return out, nil
	}
	c := cfg()
	if c == nil {
		return nil, fmt.Errorf("config not loaded")
	}
	start := time.Now()
	body := map[string]string{
		"question":  question,
		"domain_id": domainID,
		"tenant_id": tenantID,
		"locale":    locale,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal RAG request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", c.PythonRAGURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create RAG request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	setPythonServiceHeaders(req)
	if rid := traceRequestID(ctx); rid != "" {
		req.Header.Set("X-Request-ID", rid)
	}
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		traceStep(ctx, "retrieve", map[string]any{"ms": msSince(start), "ok": false, "error": "transport"})
		return nil, fmt.Errorf("RAG request failed: %w", err)
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read RAG response: %w", err)
	}
	var out ContextResponse
	if err := json.Unmarshal(responseBody, &out); err != nil {
		return nil, fmt.Errorf("parse RAG response: %w: %s", err, string(responseBody))
	}
	traceStep(ctx, "retrieve", map[string]any{
		"ms": msSince(start), "fragments": len(out.Fragments), "ok": out.Success && resp.StatusCode == http.StatusOK,
	})
	if resp.StatusCode != http.StatusOK {
		if out.Error != "" {
			return &out, fmt.Errorf("%s", out.Error)
		}
		return &out, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(responseBody))
	}
	if !out.Success && out.Error == "" {
		out.Error = "RAG returned success=false"
	}
	return &out, nil
}

const ragUserPromptTpl = `<system>%s</system>
<context>%s</context>
<examples>%s</examples>
<task>Answer the user's question clearly and accurately.</task>
<constraints>
%s
</constraints>
<output_format>
Start with the fact directly, without filler introductions. Be thorough and clear.
</output_format>
Question: %s
`

// BuildUserPrompt formats the RAG user message for the LLM.
func BuildUserPrompt(question, context, fewShot, taskIntro, constraints string) string {
	return fmt.Sprintf(ragUserPromptTpl, taskIntro, context, fewShot, constraints, question)
}

func msSince(start time.Time) int64 {
	return time.Since(start).Milliseconds()
}
