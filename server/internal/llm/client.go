package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"grounded_llm_server/internal/metrics"
	"grounded_llm_server/internal/store"
)

// Request — тело запроса к OpenAI-совместимому chat/completions.
type Request struct {
	Model    string          `json:"model"`
	Messages []store.Message `json:"messages"`
}

// Response — ответ chat/completions.
type Response struct {
	Choices []Choice `json:"choices"`
	Usage   *Usage   `json:"usage,omitempty"`
}

// Usage — usage из OpenAI-совместимого ответа.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Choice — один вариант ответа модели.
type Choice struct {
	Message store.Message `json:"message"`
}

// Complete отправляет запрос в LLM API (OpenAI-совместимый).
func Complete(messages []store.Message) (string, error) {
	return CompleteForTenant("default", messages)
}

// CompleteForTenant sends a non-streaming chat completion for the given tenant.
func CompleteForTenant(tenantID string, messages []store.Message) (string, error) {
	c := cfg()
	if MockEnabled() {
		return MockComplete(messages)
	}
	if c == nil || c.LLMAPIKey == "" {
		return "", fmt.Errorf("LLM API key not configured")
	}
	metrics.LLMRequests.Add(1)
	start := time.Now()
	llmReq := &Request{
		Model:    c.LLMModel,
		Messages: messages,
	}
	requestBody, err := json.Marshal(llmReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal LLM request: %v", err)
	}
	req, err := http.NewRequest("POST", c.LLMBaseURL+"/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create LLM request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.LLMAPIKey))
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send LLM request: %v", err)
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read LLM response: %v", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("LLM HTTP %d: %s", resp.StatusCode, truncateForErr(string(responseBody), 400))
	}
	var llmResp Response
	if err := json.Unmarshal(responseBody, &llmResp); err != nil {
		return "", fmt.Errorf("failed to parse LLM response: %v", err)
	}
	if len(llmResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in LLM response")
	}
	content := llmResp.Choices[0].Message.Content
	promptTok, completionTok := 0, 0
	if llmResp.Usage != nil {
		promptTok = llmResp.Usage.PromptTokens
		completionTok = llmResp.Usage.CompletionTokens
	}
	if promptTok == 0 {
		promptTok = estimateMessagesTokens(messages)
	}
	if completionTok == 0 {
		completionTok = metrics.EstimateTokens(content)
	}
	lat := time.Since(start)
	metrics.RecordLLMUsage(tenantID, c.LLMModel, promptTok, completionTok, lat, lat)
	return content, nil
}

func estimateMessagesTokens(messages []store.Message) int {
	contents := make([]string, len(messages))
	for i, m := range messages {
		contents[i] = m.Content
	}
	return metrics.EstimateMessagesTokens(contents)
}

func truncateForErr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
