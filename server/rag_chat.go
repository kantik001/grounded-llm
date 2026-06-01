package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// RAGFragment — фрагмент документа из Python (content — полный текст для verify).
type RAGFragment struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
	Page     int    `json:"page,omitempty"`
	Excerpt  string `json:"excerpt,omitempty"`
}

// pythonRAGContextResponse — ответ POST /rag/context.
type pythonRAGContextResponse struct {
	Success   bool          `json:"success"`
	Error     string        `json:"error,omitempty"`
	Context   string        `json:"context,omitempty"`
	FewShot   string        `json:"few_shot,omitempty"`
	Category  string        `json:"category,omitempty"`
	Fragments []RAGFragment `json:"fragments,omitempty"`
}

func fetchRAGContext(question, domainID string) (*pythonRAGContextResponse, error) {
	body := map[string]string{"question": question, "domain_id": domainID}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal RAG request: %w", err)
	}
	req, err := http.NewRequest("POST", config.PythonRAGURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create RAG request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("RAG request failed: %w", err)
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read RAG response: %w", err)
	}
	var out pythonRAGContextResponse
	if err := json.Unmarshal(responseBody, &out); err != nil {
		return nil, fmt.Errorf("parse RAG response: %w: %s", err, string(responseBody))
	}
	if resp.StatusCode != http.StatusOK {
		if out.Error != "" {
			return &out, fmt.Errorf("%s", out.Error)
		}
		return &out, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(responseBody))
	}
	return &out, nil
}

const ragUserPromptTpl = `<system>%s</system>
<context>%s</context>
<examples>%s</examples>
<task>Ответь на вопрос пользователя чётко, по делу, грамотно.</task>
<constraints>
%s
</constraints>
<output_format>
Ответ должен начинаться сразу с факта, без лишних вступлений. Будь подробным и грамотным.
</output_format>
Вопрос: %s
`

func buildRAGUserPrompt(question, context, fewShot, taskIntro, constraints string) string {
	return fmt.Sprintf(ragUserPromptTpl, taskIntro, context, fewShot, constraints, question)
}

func answerWithRAG(q, domainID string, history []Message, sessionID string) RAGAnswerResult {
	var fail RAGAnswerResult
	q = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(q, "\r", " "), "\n", " "))
	if q == "" {
		fail.ErrMsg = "Пустой вопрос"
		return fail
	}

	domainID, err := normalizeDomainID(domainID)
	if err != nil {
		fail.ErrMsg = publicAPIError(err)
		return fail
	}
	if err := requireRAGEnabled(domainID); err != nil {
		fail.ErrMsg = publicAPIError(err)
		return fail
	}

	ragOut, err := fetchRAGContext(q, domainID)
	if err != nil {
		log.Printf("RAG fetch error: %v", err)
		msg := publicAPIError(err)
		if ragOut != nil && ragOut.Error != "" {
			msg = ragOut.Error
		}
		fail.ErrMsg = msg
		return fail
	}
	if !ragOut.Success {
		logRAGOutcome(domainID, q, len(ragOut.Fragments), false, ragOut.Error, sessionID, true)
		fail.ErrMsg = ragOut.Error
		fail.SoftFail = true
		return fail
	}
	if config.LLMAPIKey == "" {
		fail.ErrMsg = "Для текстового чата задайте LLM_API_KEY (OpenRouter / OpenAI-совместимый API)."
		return fail
	}

	prompts := promptsForDomain(domainID)
	userPrompt := buildRAGUserPrompt(q, ragOut.Context, ragOut.FewShot, prompts.RAGTaskIntro, ragConstraintsText())
	var msgs []Message
	msgs = append(msgs, Message{Role: "system", Content: prompts.RAGSystem})
	msgs = append(msgs, history...)
	msgs = append(msgs, Message{Role: "user", Content: userPrompt})

	raw, err := callLLMCompletion(msgs)
	if err != nil {
		log.Printf("LLM chat error: %v", err)
		fail.ErrMsg = publicAPIError(err)
		return fail
	}
	answer := cleanRAGAnswer(raw)
	answer = appendRAGDisclaimer(answer)
	passed, reason := verifyRAGAnswer(answer, ragOut.Fragments)
	logRAGOutcome(domainID, q, len(ragOut.Fragments), passed, reason, sessionID, !passed)
	citations := publicCitations(ragOut.Fragments)
	if !passed {
		return RAGAnswerResult{
			Answer:    fmt.Sprintf("⚠️ Система не смогла подтвердить ответ источниками. %s\n\n%s", reason, verifyFailHint()),
			Citations: citations,
			OK:        true,
			SoftFail:  false,
		}
	}
	return RAGAnswerResult{Answer: answer, Citations: citations, OK: true}
}
