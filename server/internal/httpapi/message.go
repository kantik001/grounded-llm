package httpapi

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/rag"
	"grounded_llm_server/internal/store"
)

type jsonMessageRequest struct {
	SessionID string `json:"session_id"`
	Text      string `json:"text"`
	DomainID  string `json:"domain_id"`
	TenantID  string `json:"tenant_id"`
}

// Message handles POST /message (JSON or multipart; optional SSE via stream=1).
func Message(c *gin.Context) {
	s := requireServices()
	ct := c.GetHeader("Content-Type")
	var sessionID string
	var text string
	var domainIDRaw string
	var imageData []byte
	var err error

	if strings.HasPrefix(ct, "multipart/form-data") {
		if err := c.Request.ParseMultipartForm(MaxUploadImageBytes + 512*1024); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid multipart request"})
			return
		}
		sessionID = strings.TrimSpace(c.PostForm("session_id"))
		text = strings.TrimSpace(c.PostForm("text"))
		domainIDRaw = s.DomainIDFromForm(c)
		imageData, err = ReadImageFromForm(c, "image")
		if err != nil {
			s.JSONError(c, http.StatusBadRequest, err)
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
			if err := s.RequestTenantOverride(c, tid); err != nil {
				s.JSONError(c, http.StatusForbidden, err)
				return
			}
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

	tgUser, err := s.ActorUser(c)
	if err != nil {
		s.JSONError(c, http.StatusUnauthorized, err)
		return
	}

	tenantID, err := s.ResolveTenant(c)
	if err != nil {
		s.JSONError(c, http.StatusBadRequest, err)
		return
	}

	requestDomainID, err := s.NormalizeDomainID(domainIDRaw)
	if err != nil {
		s.JSONError(c, http.StatusBadRequest, err)
		return
	}

	if err := s.CheckMessageQuota(c.Request.Context(), tenantID); err != nil {
		s.QuotaExceeded(c, err)
		return
	}

	ctx := c.Request.Context()
	st := s.Store()
	if st == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Store not configured"})
		return
	}
	sid, sessionDomain, err := st.GetOrCreateSession(ctx, sessionID, tgUser, tenantID, requestDomainID)
	if err != nil {
		log.Printf("GetOrCreateSession: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Session error"})
		return
	}

	s.LogRequest(c, "message_sent", map[string]any{"domain_id": sessionDomain, "session_id": sid})

	tr := BeginPathTrace(CtxRequestID(c), tenantID)
	defer tr.End()
	tr.Step("message.accept", map[string]any{
		"domain_id": sessionDomain, "session_id": sid, "stream": WantsStream(c),
	})
	c.Request = c.Request.WithContext(ContextWithPathTrace(ctx, tr))

	if WantsStream(c) {
		sseMessage(c, sid, sessionDomain, tenantID, tgUser.ID, text)
		return
	}
	handleTextMessage(c, sid, sessionDomain, tenantID, s.Locale(c), tgUser.ID, text)
}

func respondWithMessages(c *gin.Context, sid, domainID, tenantID string, telegramID int64, extra gin.H, status int) {
	s := requireServices()
	msgs, err := s.Store().ListMessages(c.Request.Context(), sid, telegramID, tenantID)
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
	c.JSON(status, AttachRequestID(c, body))
}

func handleTextMessage(c *gin.Context, sid, domainID, tenantID, locale string, telegramID int64, text string) {
	s := requireServices()
	ctx := c.Request.Context()
	prior, err := s.Store().HistoryForLLM(ctx, sid, telegramID, tenantID, 0)
	if err != nil {
		log.Printf("HistoryForLLM: %v", err)
		c.JSON(http.StatusInternalServerError, AttachRequestID(c, gin.H{"success": false, "error": "History error"}))
		return
	}
	ragResult := rag.Answer(ctx, text, tenantID, domainID, locale, prior, sid)
	if ragResult.CacheHit {
		c.Header("X-Cache", "HIT")
	} else {
		c.Header("X-Cache", "MISS")
	}

	if _, err := s.Store().AppendMessage(ctx, sid, store.ChatMessage{Role: "user", Content: text, Kind: "text"}); err != nil {
		log.Printf("AppendMessage user: %v", err)
		c.JSON(http.StatusInternalServerError, AttachRequestID(c, gin.H{"success": false, "error": "Save error"}))
		return
	}

	s.RecordRAGAnalytics(ctx, telegramID, tenantID, domainID, text, ragResult)

	if ragResult.SoftFail {
		_, _ = s.Store().AppendMessage(ctx, sid, store.ChatMessage{Role: "assistant", Content: ragResult.ErrMsg, Kind: "assistant"})
		s.LogRequest(c, "rag_answer", map[string]any{"domain_id": domainID, "soft_fail": true})
		respondWithMessages(c, sid, domainID, tenantID, telegramID, gin.H{"error": ragResult.ErrMsg}, http.StatusOK)
		return
	}
	if !ragResult.OK {
		_, _ = s.Store().AppendMessage(ctx, sid, store.ChatMessage{Role: "assistant", Content: "Error: " + ragResult.ErrMsg, Kind: "assistant"})
		status := http.StatusInternalServerError
		if strings.Contains(ragResult.ErrMsg, "LLM_API_KEY") {
			status = http.StatusServiceUnavailable
		}
		respondWithMessages(c, sid, domainID, tenantID, telegramID, gin.H{"success": false, "error": ragResult.ErrMsg}, status)
		return
	}

	if _, err := s.Store().AppendMessage(ctx, sid, store.ChatMessage{
		Role: "assistant", Content: ragResult.Answer, Kind: "assistant", Citations: ragResult.Citations,
	}); err != nil {
		log.Printf("AppendMessage assistant: %v", err)
		c.JSON(http.StatusInternalServerError, AttachRequestID(c, gin.H{"success": false, "error": "Save error"}))
		return
	}
	s.LogRequest(c, "rag_answer", map[string]any{"domain_id": domainID, "soft_fail": false})
	respondWithMessages(c, sid, domainID, tenantID, telegramID, nil, http.StatusOK)
}
