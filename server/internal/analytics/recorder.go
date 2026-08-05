package analytics

import (
	"context"
	"log"
	"strings"

	"grounded_llm_server/internal/rag"
)

// RecordRAG logs a rag_answer analytics event when appropriate.
func RecordRAG(ctx context.Context, telegramID int64, tenantID, domainID, question string, result rag.AnswerResult) {
	if st() == nil || !shouldRecordRAG(result) {
		return
	}
	payload := map[string]any{
		"tenant_id":      tenantID,
		"domain_id":      domainID,
		"verify_pass":    result.VerifyPass,
		"soft_fail":      result.SoftFail,
		"fragment_count": result.FragmentCount,
	}
	if preview := questionPreview(question); preview != "" {
		payload["question_preview"] = preview
	}
	if err := st().LogAnalyticsEvent(ctx, telegramID, "rag_answer", payload); err != nil {
		log.Printf("LogAnalyticsEvent rag_answer: %v", err)
	}
}

func questionPreview(q string) string {
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

func shouldRecordRAG(r rag.AnswerResult) bool {
	if r.SoftFail {
		return true
	}
	return r.OK && r.ErrMsg == ""
}

// QuestionPreview normalizes and truncates a question for analytics storage.
func QuestionPreview(q string) string {
	return questionPreview(q)
}

// ShouldRecordRAG reports whether a RAG result should be logged to analytics.
func ShouldRecordRAG(r rag.AnswerResult) bool {
	return shouldRecordRAG(r)
}
