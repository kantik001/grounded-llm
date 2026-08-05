package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"grounded_llm_server/internal/metrics"
	"grounded_llm_server/internal/store"
)

type streamRequest struct {
	Model    string          `json:"model"`
	Messages []store.Message `json:"messages"`
	Stream   bool            `json:"stream"`
}

type streamDelta struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
	Usage *Usage `json:"usage,omitempty"`
}

// CompleteStreamForTenant streams chat completions and invokes onDelta for each text chunk.
func CompleteStreamForTenant(ctx context.Context, tenantID string, messages []store.Message, onDelta func(string) error) (string, error) {
	c := cfg()
	if MockEnabled() {
		full, err := MockComplete(messages)
		if err != nil {
			return "", err
		}
		if onDelta != nil && full != "" {
			if err := onDelta(full); err != nil {
				return full, err
			}
		}
		return full, nil
	}
	if c == nil || c.LLMAPIKey == "" {
		return "", fmt.Errorf("LLM API key not configured")
	}
	metrics.LLMRequests.Add(1)
	start := time.Now()
	body, err := json.Marshal(&streamRequest{
		Model:    c.LLMModel,
		Messages: messages,
		Stream:   true,
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.LLMBaseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.LLMAPIKey)
	req.Header.Set("Accept", "text/event-stream")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("LLM stream HTTP %d: %s", resp.StatusCode, string(b))
	}

	var full strings.Builder
	var ttft time.Duration
	var usage *Usage
	scanner := bufio.NewScanner(resp.Body)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data: "))
		if payload == "" || payload == "[DONE]" {
			continue
		}
		var chunk streamDelta
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			continue
		}
		if chunk.Usage != nil {
			usage = chunk.Usage
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		delta := chunk.Choices[0].Delta.Content
		if delta == "" {
			continue
		}
		if ttft == 0 {
			ttft = time.Since(start)
		}
		full.WriteString(delta)
		if onDelta != nil {
			if err := onDelta(delta); err != nil {
				return full.String(), err
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return full.String(), err
	}
	out := full.String()
	if out == "" {
		return "", fmt.Errorf("empty streamed LLM response")
	}
	promptTok, completionTok := 0, 0
	if usage != nil {
		promptTok = usage.PromptTokens
		completionTok = usage.CompletionTokens
	}
	if promptTok == 0 {
		promptTok = estimateMessagesTokens(messages)
	}
	if completionTok == 0 {
		completionTok = metrics.EstimateTokens(out)
	}
	metrics.RecordLLMUsage(tenantID, c.LLMModel, promptTok, completionTok, time.Since(start), ttft)
	return out, nil
}
