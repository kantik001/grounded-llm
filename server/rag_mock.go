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
	if mockRAGOutOfScope(q) {
		return &pythonRAGContextResponse{Success: false, Error: "no documents found for this question"}
	}
	if strings.Contains(q, "vpn") && domainID == "default" {
		return &pythonRAGContextResponse{Success: false, Error: "no information in this knowledge base"}
	}
	contextText := "Employees receive 28 paid vacation days per year."
	filename := "vacation_policy_en.txt"
	if strings.Contains(q, "vpn") || domainID == "it_support" {
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

func mockRAGOutOfScope(q string) bool {
	for _, token := range []string{
		"stock ticker", "world cup", "ceo's personal", "merger document",
		"cafeteria lunch", "google's published", "make up a plausible",
	} {
		if strings.Contains(q, token) {
			return true
		}
	}
	return false
}

func isTruthyEnv(key string) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	return v == "1" || v == "true" || v == "yes"
}
