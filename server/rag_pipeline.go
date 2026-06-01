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
}

func prepareRAGMessages(q, domainID, tenantID string, history []Message, sessionID string) (ragPrepared, error) {
	var fail ragPrepared
	metricRAGRequests.Add(1)
	q = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(q, "\r", " "), "\n", " "))
	if q == "" {
		fail.ErrMsg = "Пустой вопрос"
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

	ragOut, err := fetchRAGContext(q, tenantID, domainID)
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
		fail.ErrMsg = "Для текстового чата задайте LLM_API_KEY (OpenRouter / OpenAI-совместимый API)."
		return fail, nil
	}

	prompts := promptsForDomain(domainID)
	userPrompt := buildRAGUserPrompt(q, ragOut.Context, ragOut.FewShot, prompts.RAGTaskIntro, ragConstraintsText())
	var msgs []Message
	msgs = append(msgs, Message{Role: "system", Content: prompts.RAGSystem})
	msgs = append(msgs, history...)
	msgs = append(msgs, Message{Role: "user", Content: userPrompt})

	return ragPrepared{
		OK:          true,
		LLMMessages: msgs,
		Fragments:   ragOut.Fragments,
		DomainID:    domainID,
	}, nil
}

func finalizeRAGAnswer(raw string, p ragPrepared) RAGAnswerResult {
	answer := cleanRAGAnswer(raw)
	answer = appendRAGDisclaimer(answer)
	passed, reason := verifyRAGAnswer(answer, p.Fragments)
	logRAGOutcome(p.DomainID, "", len(p.Fragments), passed, reason, "", !passed)
	citations := publicCitations(p.Fragments)
	if !passed {
		return RAGAnswerResult{
			Answer:    fmt.Sprintf("⚠️ Система не смогла подтвердить ответ источниками. %s\n\n%s", reason, verifyFailHint()),
			Citations: citations,
			OK:        true,
		}
	}
	return RAGAnswerResult{Answer: answer, Citations: citations, OK: true}
}

func answerWithRAG(q, tenantID, domainID string, history []Message, sessionID string) RAGAnswerResult {
	prepared, err := prepareRAGMessages(q, domainID, tenantID, history, sessionID)
	if err != nil {
		return RAGAnswerResult{ErrMsg: publicAPIError(err)}
	}
	if !prepared.OK {
		return RAGAnswerResult{ErrMsg: prepared.ErrMsg, SoftFail: prepared.SoftFail}
	}
	raw, err := callLLMCompletion(prepared.LLMMessages)
	if err != nil {
		log.Printf("LLM chat error: %v", err)
		return RAGAnswerResult{ErrMsg: publicAPIError(err)}
	}
	return finalizeRAGAnswer(raw, prepared)
}
