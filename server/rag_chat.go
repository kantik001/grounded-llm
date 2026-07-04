package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// RAGFragment — фрагмент документа из Python (content — полный текст для verify).
type RAGFragment struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
	Page     int    `json:"page,omitempty"`
	Excerpt  string `json:"excerpt,omitempty"`
}

// pythonRAGContextResponse — ответ POST /rag/context.
type pythonRAGContextResponse struct {
	Success   bool          `json:"success"`
	Error     string        `json:"error,omitempty"`
	Context   string        `json:"context,omitempty"`
	FewShot   string        `json:"few_shot,omitempty"`
	Category  string        `json:"category,omitempty"`
	Fragments []RAGFragment `json:"fragments,omitempty"`
}

func fetchRAGContext(question, tenantID, domainID, locale string) (*pythonRAGContextResponse, error) {
	if ragMockEnabled() {
		return mockRAGContextResponse(question, domainID), nil
	}
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
	req, err := http.NewRequest("POST", config.PythonRAGURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create RAG request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	setPythonServiceHeaders(req)
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("RAG request failed: %w", err)
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read RAG response: %w", err)
	}
	var out pythonRAGContextResponse
	if err := json.Unmarshal(responseBody, &out); err != nil {
		return nil, fmt.Errorf("parse RAG response: %w: %s", err, string(responseBody))
	}
	if resp.StatusCode != http.StatusOK {
		if out.Error != "" {
			return &out, fmt.Errorf("%s", out.Error)
		}
		return &out, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(responseBody))
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

func buildRAGUserPrompt(question, context, fewShot, taskIntro, constraints string) string {
	return fmt.Sprintf(ragUserPromptTpl, taskIntro, context, fewShot, constraints, question)
}
