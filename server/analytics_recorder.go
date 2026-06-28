package main

import (
	"context"
	"log"
	"strings"
)

func recordRAGAnalytics(ctx context.Context, telegramID int64, tenantID, domainID, question string, result RAGAnswerResult) {
	if chatStore == nil || !shouldRecordRAGAnalytics(result) {
		return
	}
	payload := map[string]any{
		"tenant_id":      tenantID,
		"domain_id":      domainID,
		"verify_pass":    result.VerifyPass,
		"soft_fail":      result.SoftFail,
		"fragment_count": result.FragmentCount,
	}
	if preview := questionPreviewForAnalytics(question); preview != "" {
		payload["question_preview"] = preview
	}
	if err := chatStore.LogAnalyticsEvent(ctx, telegramID, "rag_answer", payload); err != nil {
		log.Printf("LogAnalyticsEvent rag_answer: %v", err)
	}
}

func questionPreviewForAnalytics(q string) string {
	q = strings.TrimSpace(strings.Join(strings.Fields(q), " "))
	const maxLen = 80
	if q == "" {
		return ""
	}
	if len(q) <= maxLen {
		return q
	}
	return q[:maxLen] + "…"
}

func shouldRecordRAGAnalytics(r RAGAnswerResult) bool {
	if r.SoftFail {
		return true
	}
	return r.OK && r.ErrMsg == ""
}
