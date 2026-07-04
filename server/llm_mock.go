package main

import (
	"strings"
)

func llmMockEnabled() bool {
	return config != nil && config.LLMMock
}

// mockLLMCompletion returns a deterministic answer for CI and local smoke tests.
func mockLLMCompletion(messages []Message) (string, error) {
	metricLLMRequests.Add(1)
	prompt := ""
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			prompt = strings.ToLower(messages[i].Content)
			break
		}
	}
	if strings.Contains(prompt, "vpn") {
		return "Use the corporate VPN client. Support SLA is 4 hours.", nil
	}
	return "Employees receive 28 paid vacation days per year.", nil
}
