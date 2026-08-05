package rag

import (
	"strings"

	"grounded_llm_server/internal/store"
)

// MockEnabled reports whether RAG_MOCK is on.
func MockEnabled() bool {
	c := cfg()
	return c != nil && c.RAGMock
}

// MockContextResponse returns a deterministic retrieval payload for CI.
func MockContextResponse(question, domainID string) *ContextResponse {
	q := strings.ToLower(strings.TrimSpace(question))
	if mockRAGOutOfScope(q) {
		return &ContextResponse{Success: false, Error: "no documents found for this question"}
	}
	if strings.Contains(q, "vpn") && domainID == "default" {
		return &ContextResponse{Success: false, Error: "no information in this knowledge base"}
	}
	contextText := "Employees receive 28 paid vacation days per year."
	filename := "vacation_policy_en.txt"
	if strings.Contains(q, "vpn") || domainID == "it_support" {
		contextText = "Connect to VPN using the corporate client. Support SLA is 4 hours."
		filename = "vpn_access.txt"
	}
	return &ContextResponse{
		Success:  true,
		Context:  contextText,
		Category: "mock",
		Fragments: []store.RAGFragment{{
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
