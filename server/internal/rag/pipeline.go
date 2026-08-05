package rag

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"grounded_llm_server/internal/guardrails"
	"grounded_llm_server/internal/llm"
	"grounded_llm_server/internal/metrics"
	"grounded_llm_server/internal/store"
)

// Prepared is the LLM-ready RAG turn after retrieval.
type Prepared struct {
	OK          bool
	SoftFail    bool
	ErrMsg      string
	LLMMessages []store.Message
	Fragments   []store.RAGFragment
	DomainID    string
	Locale      string
}

// PrepareMessages retrieves context and builds LLM messages.
func PrepareMessages(ctx context.Context, q, domainID, tenantID, locale string, history []store.Message, sessionID string) (Prepared, error) {
	var fail Prepared
	metrics.RAGRequests.Add(1)
	q = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(q, "\r", " "), "\n", " "))
	if q == "" {
		fail.ErrMsg = "Empty question"
		return fail, nil
	}
	if host == nil {
		fail.ErrMsg = "RAG host not bound"
		return fail, nil
	}

	domainID, err := host.NormalizeDomainID(domainID)
	if err != nil {
		fail.ErrMsg = host.PublicAPIError(err)
		return fail, nil
	}
	if err := host.RequireRAGEnabled(domainID); err != nil {
		fail.ErrMsg = host.PublicAPIError(err)
		return fail, nil
	}

	ragOut, err := FetchContext(ctx, q, tenantID, domainID, locale)
	if err != nil {
		log.Printf("RAG fetch error: %v", err)
		msg := host.PublicAPIError(err)
		if ragOut != nil && ragOut.Error != "" {
			msg = ragOut.Error
		}
		fail.ErrMsg = msg
		return fail, nil
	}
	if !ragOut.Success {
		LogOutcome(ctx, domainID, q, len(ragOut.Fragments), false, ragOut.Error, sessionID, true)
		fail.ErrMsg = ragOut.Error
		fail.SoftFail = true
		return fail, nil
	}
	c := cfg()
	if c != nil && c.LLMAPIKey == "" && !c.LLMMock {
		fail.ErrMsg = "Set LLM_API_KEY for text chat (OpenRouter / OpenAI-compatible API)."
		return fail, nil
	}

	system, taskIntro := host.Prompts(domainID, locale)
	userPrompt := BuildUserPrompt(q, ragOut.Context, ragOut.FewShot, taskIntro, host.RAGConstraints(locale))
	var msgs []store.Message
	msgs = append(msgs, store.Message{Role: "system", Content: system})
	msgs = append(msgs, history...)
	msgs = append(msgs, store.Message{Role: "user", Content: userPrompt})

	return Prepared{
		OK:          true,
		LLMMessages: msgs,
		Fragments:   ragOut.Fragments,
		DomainID:    domainID,
		Locale:      locale,
	}, nil
}

// FinalizeAnswer runs verify and builds the public result.
func FinalizeAnswer(ctx context.Context, raw string, p Prepared) AnswerResult {
	answer := CleanAnswer(raw)
	answer = AppendDisclaimer(answer, p.Locale)
	start := time.Now()
	passed, reason := VerifyAnswer(answer, p.Fragments, p.Locale)
	mode := "local"
	if c := cfg(); c != nil {
		mode = string(guardrails.NormalizeMode(c.GuardrailsMode))
	}
	traceStep(ctx, "verify", map[string]any{
		"ms": msSince(start), "pass": passed, "mode": mode,
	})
	LogOutcome(ctx, p.DomainID, "", len(p.Fragments), passed, reason, "", !passed)
	citations := PublicCitations(p.Fragments)
	fragmentCount := len(p.Fragments)
	if !passed {
		hint := "Contact your knowledge base administrator."
		if host != nil {
			hint = host.VerifyFailHint(p.Locale)
		}
		return AnswerResult{
			Answer:        fmt.Sprintf("⚠️ Could not verify the answer against sources. %s\n\n%s", reason, hint),
			Citations:     citations,
			OK:            true,
			VerifyPass:    false,
			FragmentCount: fragmentCount,
		}
	}
	return AnswerResult{
		Answer:        answer,
		Citations:     citations,
		OK:            true,
		VerifyPass:    true,
		FragmentCount: fragmentCount,
	}
}

// Answer runs the full non-streaming RAG pipeline.
func Answer(ctx context.Context, q, tenantID, domainID, locale string, history []store.Message, sessionID string) AnswerResult {
	if ctx == nil {
		ctx = context.Background()
	}
	if len(history) == 0 {
		if cached, ok := llm.GetCachedAnswer(ctx, q, domainID, tenantID); ok {
			metrics.CacheHits.Add(1)
			traceStep(ctx, "cache", map[string]any{"hit": true})
			traceStep(ctx, "done", map[string]any{
				"ms": traceElapsedMS(ctx), "verify_pass": true, "cache": true,
			})
			return AnswerResult{
				Answer:        cached.Answer,
				Citations:     cached.Citations,
				OK:            true,
				VerifyPass:    true,
				CacheHit:      true,
				FragmentCount: len(cached.Citations),
			}
		}
		metrics.CacheMisses.Add(1)
		traceStep(ctx, "cache", map[string]any{"hit": false})
	}

	prepared, err := PrepareMessages(ctx, q, domainID, tenantID, locale, history, sessionID)
	if err != nil {
		msg := "Server error"
		if host != nil {
			msg = host.PublicAPIError(err)
		}
		return AnswerResult{ErrMsg: msg}
	}
	if !prepared.OK {
		traceStep(ctx, "done", map[string]any{
			"ms": traceElapsedMS(ctx), "ok": false, "soft_fail": prepared.SoftFail,
		})
		return AnswerResult{
			ErrMsg:        prepared.ErrMsg,
			SoftFail:      prepared.SoftFail,
			FragmentCount: len(prepared.Fragments),
		}
	}
	llmStart := time.Now()
	raw, err := llm.CompleteForTenant(tenantID, prepared.LLMMessages)
	traceStep(ctx, "llm", map[string]any{"ms": msSince(llmStart), "ok": err == nil, "stream": false})
	if err != nil {
		log.Printf("LLM chat error: %v", err)
		traceStep(ctx, "done", map[string]any{"ms": traceElapsedMS(ctx), "ok": false})
		msg := "Server error"
		if host != nil {
			msg = host.PublicAPIError(err)
		}
		return AnswerResult{ErrMsg: msg}
	}
	result := FinalizeAnswer(ctx, raw, prepared)
	if result.OK && result.VerifyPass && len(history) == 0 {
		llm.SetCachedAnswer(ctx, q, domainID, tenantID, result.Answer, result.Citations)
	}
	traceStep(ctx, "done", map[string]any{
		"ms": traceElapsedMS(ctx), "ok": result.OK, "verify_pass": result.VerifyPass,
	})
	return result
}
