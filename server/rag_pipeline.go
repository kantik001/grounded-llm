package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

type ragPrepared struct {
	OK          bool
	SoftFail    bool
	ErrMsg      string
	LLMMessages []Message
	Fragments   []RAGFragment
	DomainID    string
	Locale      string
}

func prepareRAGMessages(ctx context.Context, q, domainID, tenantID, locale string, history []Message, sessionID string) (ragPrepared, error) {
	var fail ragPrepared
	metricRAGRequests.Add(1)
	q = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(q, "\r", " "), "\n", " "))
	if q == "" {
		fail.ErrMsg = "Empty question"
		return fail, nil
	}

	domainID, err := normalizeDomainID(domainID)
	if err != nil {
		fail.ErrMsg = publicAPIError(err)
		return fail, nil
	}
	if err := requireRAGEnabled(domainID); err != nil {
		fail.ErrMsg = publicAPIError(err)
		return fail, nil
	}

	ragOut, err := fetchRAGContext(ctx, q, tenantID, domainID, locale)
	if err != nil {
		log.Printf("RAG fetch error: %v", err)
		msg := publicAPIError(err)
		if ragOut != nil && ragOut.Error != "" {
			msg = ragOut.Error
		}
		fail.ErrMsg = msg
		return fail, nil
	}
	if !ragOut.Success {
		logRAGOutcome(ctx, domainID, q, len(ragOut.Fragments), false, ragOut.Error, sessionID, true)
		fail.ErrMsg = ragOut.Error
		fail.SoftFail = true
		return fail, nil
	}
	if config.LLMAPIKey == "" && !config.LLMMock {
		fail.ErrMsg = "Set LLM_API_KEY for text chat (OpenRouter / OpenAI-compatible API)."
		return fail, nil
	}

	prompts := promptsForDomainLocale(domainID, locale)
	userPrompt := buildRAGUserPrompt(q, ragOut.Context, ragOut.FewShot, prompts.RAGTaskIntro, ragConstraintsForLocale(locale))
	var msgs []Message
	msgs = append(msgs, Message{Role: "system", Content: prompts.RAGSystem})
	msgs = append(msgs, history...)
	msgs = append(msgs, Message{Role: "user", Content: userPrompt})

	return ragPrepared{
		OK:          true,
		LLMMessages: msgs,
		Fragments:   ragOut.Fragments,
		DomainID:    domainID,
		Locale:      locale,
	}, nil
}

func finalizeRAGAnswer(ctx context.Context, raw string, p ragPrepared) RAGAnswerResult {
	answer := cleanRAGAnswer(raw)
	answer = appendRAGDisclaimer(answer, p.Locale)
	start := time.Now()
	passed, reason := verifyRAGAnswer(answer, p.Fragments, p.Locale)
	mode := "local"
	if config != nil {
		mode = string(normalizeGuardrailsMode(config.GuardrailsMode))
	}
	if tr := pathTraceFrom(ctx); tr != nil {
		tr.step("verify", map[string]any{
			"ms": msSince(start), "pass": passed, "mode": mode,
		})
	}
	logRAGOutcome(ctx, p.DomainID, "", len(p.Fragments), passed, reason, "", !passed)
	citations := publicCitations(p.Fragments)
	fragmentCount := len(p.Fragments)
	if !passed {
		return RAGAnswerResult{
			Answer:        fmt.Sprintf("⚠️ Could not verify the answer against sources. %s\n\n%s", reason, verifyFailHintForLocale(p.Locale)),
			Citations:     citations,
			OK:            true,
			VerifyPass:    false,
			FragmentCount: fragmentCount,
		}
	}
	return RAGAnswerResult{
		Answer:        answer,
		Citations:     citations,
		OK:            true,
		VerifyPass:    true,
		FragmentCount: fragmentCount,
	}
}

func answerWithRAG(ctx context.Context, q, tenantID, domainID, locale string, history []Message, sessionID string) RAGAnswerResult {
	if ctx == nil {
		ctx = context.Background()
	}
	if len(history) == 0 {
		if cached, ok := getCachedLLMAnswer(ctx, q, domainID, tenantID); ok {
			metricCacheHits.Add(1)
			if tr := pathTraceFrom(ctx); tr != nil {
				tr.step("cache", map[string]any{"hit": true})
				tr.step("done", map[string]any{
					"ms": tr.elapsedMS(), "verify_pass": true, "cache": true,
				})
			}
			return RAGAnswerResult{
				Answer:        cached.Answer,
				Citations:     cached.Citations,
				OK:            true,
				VerifyPass:    true,
				CacheHit:      true,
				FragmentCount: len(cached.Citations),
			}
		}
		metricCacheMisses.Add(1)
		if tr := pathTraceFrom(ctx); tr != nil {
			tr.step("cache", map[string]any{"hit": false})
		}
	}

	prepared, err := prepareRAGMessages(ctx, q, domainID, tenantID, locale, history, sessionID)
	if err != nil {
		return RAGAnswerResult{ErrMsg: publicAPIError(err)}
	}
	if !prepared.OK {
		if tr := pathTraceFrom(ctx); tr != nil {
			tr.step("done", map[string]any{
				"ms": tr.elapsedMS(), "ok": false, "soft_fail": prepared.SoftFail,
			})
		}
		return RAGAnswerResult{
			ErrMsg:        prepared.ErrMsg,
			SoftFail:      prepared.SoftFail,
			FragmentCount: len(prepared.Fragments),
		}
	}
	llmStart := time.Now()
	raw, err := callLLMCompletionForTenant(tenantID, prepared.LLMMessages)
	if tr := pathTraceFrom(ctx); tr != nil {
		tr.step("llm", map[string]any{"ms": msSince(llmStart), "ok": err == nil, "stream": false})
	}
	if err != nil {
		log.Printf("LLM chat error: %v", err)
		if tr := pathTraceFrom(ctx); tr != nil {
			tr.step("done", map[string]any{"ms": tr.elapsedMS(), "ok": false})
		}
		return RAGAnswerResult{ErrMsg: publicAPIError(err)}
	}
	result := finalizeRAGAnswer(ctx, raw, prepared)
	if result.OK && result.VerifyPass && len(history) == 0 {
		setCachedLLMAnswer(ctx, q, domainID, tenantID, result.Answer, result.Citations)
	}
	if tr := pathTraceFrom(ctx); tr != nil {
		tr.step("done", map[string]any{
			"ms": tr.elapsedMS(), "ok": result.OK, "verify_pass": result.VerifyPass,
		})
	}
	return result
}
