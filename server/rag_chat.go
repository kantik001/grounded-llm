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

	"github.com/gin-gonic/gin"
)

// RAGFragment — фрагмент документа из Python.
type RAGFragment struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
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

type ChatRequest struct {
	Question string `json:"question"`
	DomainID string `json:"domain_id"`
}

func answerWithRAG(q, domainID string, history []Message, sessionID string) (answer string, success bool, errMsg string, ragSoftFail bool) {
	q = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(q, "\r", " "), "\n", " "))
	if q == "" {
		return "", false, "Пустой вопрос", false
	}

	domainID, err := normalizeDomainID(domainID)
	if err != nil {
		return "", false, publicAPIError(err), false
	}
	if err := requireRAGEnabled(domainID); err != nil {
		return "", false, publicAPIError(err), false
	}

	ragOut, err := fetchRAGContext(q, domainID)
	if err != nil {
		log.Printf("RAG fetch error: %v", err)
		msg := publicAPIError(err)
		if ragOut != nil && ragOut.Error != "" {
			msg = ragOut.Error
		}
		return "", false, msg, false
	}
	if !ragOut.Success {
		logRAGOutcome(domainID, q, len(ragOut.Fragments), false, ragOut.Error, sessionID, true)
		return "", false, ragOut.Error, true
	}
	if config.LLMAPIKey == "" {
		return "", false, "Для текстового чата задайте LLM_API_KEY (OpenRouter / OpenAI-совместимый API).", false
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
		return "", false, publicAPIError(err), false
	}
	answer = cleanRAGAnswer(raw)
	answer = appendRAGDisclaimer(answer)
	passed, reason := verifyRAGAnswer(answer, ragOut.Fragments)
	logRAGOutcome(domainID, q, len(ragOut.Fragments), passed, reason, sessionID, !passed)
	if !passed {
		return fmt.Sprintf("⚠️ Система не смогла подтвердить ответ источниками. %s\n\n%s", reason, verifyFailHint()), true, "", false
	}
	return answer, true, "", false
}

func handleChat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Некорректный JSON (нужно поле question)",
		})
		return
	}
	q := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(req.Question, "\r", " "), "\n", " "))
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Пустой вопрос"})
		return
	}
	domainID := strings.TrimSpace(req.DomainID)

	answer, ok, errMsg, ragSoft := answerWithRAG(q, domainID, nil, "")
	if ragSoft {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": errMsg})
		return
	}
	if errMsg != "" && !ok {
		if strings.Contains(errMsg, "LLM_API_KEY") {
			c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": errMsg})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": errMsg})
		return
	}
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": errMsg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "answer": answer})
}
