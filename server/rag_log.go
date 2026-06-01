package main

import (
	"log"
	"strings"
)

// logRAGOutcome пишет структурированную строку для разбора качества (без тела LLM).
func logRAGOutcome(domainID, question string, fragmentCount int, verifyPass bool, verifyReason, sessionID string, softFail bool) {
	q := strings.TrimSpace(question)
	if len(q) > 120 {
		q = q[:120] + "…"
	}
	log.Printf(
		"[RAG] domain_id=%s session_id=%s fragments=%d verify_pass=%v soft_fail=%v reason=%q question=%q",
		domainID,
		sessionID,
		fragmentCount,
		verifyPass,
		softFail,
		verifyReason,
		q,
	)
}
