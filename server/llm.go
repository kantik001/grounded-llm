package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LLMRequest — тело запроса к OpenAI-совместимому chat/completions.
type LLMRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// Message — одно сообщение в диалоге для LLM.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LLMResponse — ответ chat/completions.
type LLMResponse struct {
	Choices []Choice  `json:"choices"`
	Usage   *LLMUsage `json:"usage,omitempty"`
}

// LLMUsage — usage из OpenAI-совместимого ответа.
type LLMUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Choice — один вариант ответа модели.
type Choice struct {
	Message Message `json:"message"`
}

// callLLMCompletion отправляет запрос в LLM API (OpenAI-совместимый).
func callLLMCompletion(messages []Message) (string, error) {
	return callLLMCompletionForTenant("default", messages)
}

func callLLMCompletionForTenant(tenantID string, messages []Message) (string, error) {
	if llmMockEnabled() {
		return mockLLMCompletion(messages)
	}
	if config.LLMAPIKey == "" {
		return "", fmt.Errorf("LLM API key not configured")
	}
	metricLLMRequests.Add(1)
	start := time.Now()
	llmReq := &LLMRequest{
		Model:    config.LLMModel,
		Messages: messages,
	}
	requestBody, err := json.Marshal(llmReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal LLM request: %v", err)
	}
	req, err := http.NewRequest("POST", config.LLMBaseURL+"/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create LLM request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.LLMAPIKey))
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
	var llmResp LLMResponse
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
		completionTok = estimateTokens(content)
	}
	// Non-stream: treat full latency as TTFT proxy for dashboards.
	lat := time.Since(start)
	recordLLMUsage(tenantID, config.LLMModel, promptTok, completionTok, lat, lat)
	return content, nil
}

func truncateForErr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
