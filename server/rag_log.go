package main

import (
	"context"
	"log"
	"strings"
)

// logRAGOutcome пишет структурированную строку для разбора качества (без тела LLM).
func logRAGOutcome(ctx context.Context, domainID, question string, fragmentCount int, verifyPass bool, verifyReason, sessionID string, softFail bool) {
	q := strings.TrimSpace(question)
	if len(q) > 120 {
		q = q[:120] + "…"
	}
	req := ""
	if tr := pathTraceFrom(ctx); tr != nil {
		req = tr.reqID
	}
	if req != "" {
		log.Printf(
			"[RAG] req=%s domain_id=%s session_id=%s fragments=%d verify_pass=%v soft_fail=%v reason=%q question=%q",
			req,
			domainID,
			sessionID,
			fragmentCount,
			verifyPass,
			softFail,
			verifyReason,
			q,
		)
		return
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
