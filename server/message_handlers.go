package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type jsonMessageRequest struct {
	SessionID string `json:"session_id"`
	Text      string `json:"text"`
	DomainID  string `json:"domain_id"`
	TenantID  string `json:"tenant_id"`
}

func handleMessage(c *gin.Context) {
	ct := c.GetHeader("Content-Type")
	var sessionID string
	var text string
	var domainIDRaw string
	var imageData []byte
	var err error

	if strings.HasPrefix(ct, "multipart/form-data") {
		if err := c.Request.ParseMultipartForm(int64(maxUploadImageBytes + 512*1024)); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid multipart request"})
			return
		}
		sessionID = strings.TrimSpace(c.PostForm("session_id"))
		text = strings.TrimSpace(c.PostForm("text"))
		domainIDRaw = domainIDFromForm(c)
		imageData, err = readImageFromFormFile(c, "image")
		if err != nil {
			jsonError(c, http.StatusBadRequest, err)
			return
		}
	} else {
		var req jsonMessageRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Expected JSON: session_id, text"})
			return
		}
		sessionID = strings.TrimSpace(req.SessionID)
		text = strings.TrimSpace(req.Text)
		domainIDRaw = strings.TrimSpace(req.DomainID)
		if tid := strings.TrimSpace(req.TenantID); tid != "" {
			c.Set(ctxKeyTenantID, normalizeTenantID(tid))
		}
	}

	if text == "" && len(imageData) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Text is required"})
		return
	}

	if len(imageData) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Image upload is disabled in platform core. Attach a vision domain pack.",
		})
		return
	}

	tgUser, err := ctxActorUser(c)
	if err != nil {
		jsonError(c, http.StatusUnauthorized, err)
		return
	}

	tenantID, err := resolveTenantID(c, config)
	if err != nil {
		jsonError(c, http.StatusBadRequest, err)
		return
	}

	requestDomainID, err := normalizeDomainID(domainIDRaw)
	if err != nil {
		jsonError(c, http.StatusBadRequest, err)
		return
	}

	ctx := c.Request.Context()
	sid, sessionDomain, err := chatStore.GetOrCreateSession(ctx, sessionID, tgUser, tenantID, requestDomainID)
	if err != nil {
		log.Printf("GetOrCreateSession: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Session error"})
		return
	}

	logRequest(c, "message_sent", map[string]any{"domain_id": sessionDomain, "session_id": sid})

	if wantsStream(c) {
		sseMessageHandler(c, sid, sessionDomain, tenantID, tgUser.ID, text)
		return
	}
	handleTextMessage(c, sid, sessionDomain, tenantID, ctxLocale(c), tgUser.ID, text)
}

func respondWithMessages(c *gin.Context, sid, domainID, tenantID string, telegramID int64, extra gin.H, status int) {
	msgs, err := chatStore.ListMessages(c.Request.Context(), sid, telegramID)
	if err != nil {
		log.Printf("ListMessages after reply: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Database error"})
		return
	}
	body := gin.H{
		"success":    true,
		"session_id": sid,
		"domain_id":  domainID,
		"tenant_id":  tenantID,
		"messages":   msgs,
	}
	for k, v := range extra {
		body[k] = v
	}
	c.JSON(status, body)
}

func handleTextMessage(c *gin.Context, sid, domainID, tenantID, locale string, telegramID int64, text string) {
	ctx := c.Request.Context()
	prior, err := chatStore.HistoryForLLM(ctx, sid, telegramID, 0)
	if err != nil {
		log.Printf("HistoryForLLM: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "History error"})
		return
	}
	ragResult := answerWithRAG(text, tenantID, domainID, locale, prior, sid)

	if _, err := chatStore.AppendMessage(ctx, sid, ChatMessage{Role: "user", Content: text, Kind: "text"}); err != nil {
		log.Printf("AppendMessage user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Save error"})
		return
	}

	if ragResult.SoftFail {
		_, _ = chatStore.AppendMessage(ctx, sid, ChatMessage{Role: "assistant", Content: ragResult.ErrMsg, Kind: "assistant"})
		logRequest(c, "rag_answer", map[string]any{"domain_id": domainID, "soft_fail": true})
		respondWithMessages(c, sid, domainID, tenantID, telegramID, gin.H{"error": ragResult.ErrMsg}, http.StatusOK)
		return
	}
	if !ragResult.OK {
		_, _ = chatStore.AppendMessage(ctx, sid, ChatMessage{Role: "assistant", Content: "Error: " + ragResult.ErrMsg, Kind: "assistant"})
		status := http.StatusInternalServerError
		if strings.Contains(ragResult.ErrMsg, "LLM_API_KEY") {
			status = http.StatusServiceUnavailable
		}
		respondWithMessages(c, sid, domainID, tenantID, telegramID, gin.H{"success": false, "error": ragResult.ErrMsg}, status)
		return
	}

	if _, err := chatStore.AppendMessage(ctx, sid, ChatMessage{
		Role: "assistant", Content: ragResult.Answer, Kind: "assistant", Citations: ragResult.Citations,
	}); err != nil {
		log.Printf("AppendMessage assistant: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Save error"})
		return
	}
	logRequest(c, "rag_answer", map[string]any{"domain_id": domainID, "soft_fail": false})
	respondWithMessages(c, sid, domainID, tenantID, telegramID, nil, http.StatusOK)
}
