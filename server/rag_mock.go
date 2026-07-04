package main

import (
	"os"
	"strings"
)

func ragMockEnabled() bool {
	return config != nil && config.RAGMock
}

func mockRAGContextResponse(question, domainID string) *pythonRAGContextResponse {
	q := strings.ToLower(strings.TrimSpace(question))
	contextText := "Employees receive 28 paid vacation days per year."
	filename := "vacation_policy_en.txt"
	if strings.Contains(q, "vpn") || strings.Contains(domainID, "it") {
		contextText = "Connect to VPN using the corporate client. Support SLA is 4 hours."
		filename = "vpn_access.txt"
	}
	return &pythonRAGContextResponse{
		Success:  true,
		Context:  contextText,
		Category: "mock",
		Fragments: []RAGFragment{{
			Filename: filename,
			Content:  contextText,
			Excerpt:  contextText,
		}},
	}
}

func isTruthyEnv(key string) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	return v == "1" || v == "true" || v == "yes"
}
