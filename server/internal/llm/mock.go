package llm

import (
	"strings"

	"grounded_llm_server/internal/metrics"
	"grounded_llm_server/internal/store"
)

// MockEnabled reports whether LLM_MOCK is on.
func MockEnabled() bool {
	c := cfg()
	return c != nil && c.LLMMock
}

// MockComplete returns a deterministic answer for CI and local smoke tests.
func MockComplete(messages []store.Message) (string, error) {
	metrics.LLMRequests.Add(1)
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
