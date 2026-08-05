package rag

import (
	"context"
	"log"
	"strings"
)

// LogOutcome writes a structured quality line (no LLM body).
func LogOutcome(ctx context.Context, domainID, question string, fragmentCount int, verifyPass bool, verifyReason, sessionID string, softFail bool) {
	q := strings.TrimSpace(question)
	if len(q) > 120 {
		q = q[:120] + "…"
	}
	req := traceRequestID(ctx)
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
