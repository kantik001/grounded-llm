package main

import (
	"fmt"
	"log"
	"strings"
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

func prepareRAGMessages(q, domainID, tenantID, locale string, history []Message, sessionID string) (ragPrepared, error) {
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

	ragOut, err := fetchRAGContext(q, tenantID, domainID, locale)
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
		logRAGOutcome(domainID, q, len(ragOut.Fragments), false, ragOut.Error, sessionID, true)
		fail.ErrMsg = ragOut.Error
		fail.SoftFail = true
		return fail, nil
	}
	if config.LLMAPIKey == "" {
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

func finalizeRAGAnswer(raw string, p ragPrepared) RAGAnswerResult {
	answer := cleanRAGAnswer(raw)
	answer = appendRAGDisclaimer(answer, p.Locale)
	passed, reason := verifyRAGAnswer(answer, p.Fragments, p.Locale)
	logRAGOutcome(p.DomainID, "", len(p.Fragments), passed, reason, "", !passed)
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

func answerWithRAG(q, tenantID, domainID, locale string, history []Message, sessionID string) RAGAnswerResult {
	prepared, err := prepareRAGMessages(q, domainID, tenantID, locale, history, sessionID)
	if err != nil {
		return RAGAnswerResult{ErrMsg: publicAPIError(err)}
	}
	if !prepared.OK {
		return RAGAnswerResult{
			ErrMsg:        prepared.ErrMsg,
			SoftFail:      prepared.SoftFail,
			FragmentCount: len(prepared.Fragments),
		}
	}
	raw, err := callLLMCompletion(prepared.LLMMessages)
	if err != nil {
		log.Printf("LLM chat error: %v", err)
		return RAGAnswerResult{ErrMsg: publicAPIError(err)}
	}
	return finalizeRAGAnswer(raw, prepared)
}
