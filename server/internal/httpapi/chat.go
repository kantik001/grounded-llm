package httpapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/domain"
	"grounded_llm_server/internal/locale"
	"grounded_llm_server/internal/store"
)

type newSessionRequest struct {
	DomainID string `json:"domain_id"`
}

// NewSession handles POST /session.
func NewSession(c *gin.Context) {
	s := requireServices()
	var req newSessionRequest
	_ = c.ShouldBindJSON(&req)

	domainID := strings.TrimSpace(req.DomainID)
	domainID, err := s.NormalizeDomainID(domainID)
	if err != nil {
		JSONError(c, http.StatusBadRequest, err)
		return
	}

	tgUser, err := s.ActorUser(c)
	if err != nil {
		s.JSONError(c, http.StatusUnauthorized, err)
		return
	}
	ctx := c.Request.Context()
	st := s.Store()
	if st == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Store not configured"})
		return
	}
	userID, err := st.UpsertUser(ctx, tgUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "User error"})
		return
	}
	sid, err := st.CreateSession(ctx, userID, s.TenantID(c), domainID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create session"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"session_id": sid,
		"domain_id":  domainID,
		"tenant_id":  s.TenantID(c),
	})
}

// History handles GET /history.
func History(c *gin.Context) {
	s := requireServices()
	id := strings.TrimSpace(c.Query("session_id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "session_id is required"})
		return
	}
	tgUser, err := s.ActorUser(c)
	if err != nil {
		s.JSONError(c, http.StatusUnauthorized, err)
		return
	}
	st := s.Store()
	tenantID := s.TenantID(c)
	msgs, err := st.ListMessages(c.Request.Context(), id, tgUser.ID, tenantID)
	if err != nil {
		if err == store.ErrSessionNotFound {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Database error"})
		return
	}
	domainID, _ := st.SessionDomainID(c.Request.Context(), id, tgUser.ID, tenantID)
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"session_id": id,
		"domain_id":  domainID,
		"messages":   msgs,
	})
}

// Media handles GET /media/:token.
func Media(c *gin.Context) {
	s := requireServices()
	token := strings.TrimSpace(c.Param("token"))
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid token"})
		return
	}
	tgUser, err := s.ActorUser(c)
	if err != nil {
		s.JSONError(c, http.StatusUnauthorized, err)
		return
	}
	st := s.Store()
	if st == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "Storage unavailable"})
		return
	}
	ok, err := st.UserCanAccessImage(c.Request.Context(), token, tgUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Database error"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "File not found"})
		return
	}
	data, err := st.ReadImage(token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "File not found"})
		return
	}
	c.Data(http.StatusOK, "application/octet-stream", data)
}

type feedbackRequest struct {
	SessionID string `json:"session_id"`
	MessageID int64  `json:"message_id"`
	Rating    int    `json:"rating"`
}

// Feedback handles POST /feedback.
func Feedback(c *gin.Context) {
	s := requireServices()
	tgUser, err := s.ActorUser(c)
	if err != nil {
		JSONError(c, http.StatusUnauthorized, err)
		return
	}
	var req feedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Required: session_id, message_id, rating (1 or -1)"})
		return
	}
	if req.MessageID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid message_id"})
		return
	}
	ctx := c.Request.Context()
	st := s.Store()
	if err := st.SaveMessageFeedback(ctx, tgUser.ID, s.TenantID(c), req.MessageID, req.Rating); err != nil {
		if err == store.ErrSessionNotFound {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Message not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to save rating"})
		return
	}
	_ = st.LogEvent(ctx, tgUser.ID, "message_feedback", map[string]any{
		"message_id": req.MessageID,
		"rating":     req.Rating,
		"session_id": req.SessionID,
	})
	c.JSON(http.StatusOK, gin.H{"success": true, "message_id": req.MessageID, "rating": req.Rating})
}

// Domains handles GET /domains.
func Domains(c *gin.Context) {
	s := requireServices()
	localeCode := s.Locale(c)
	list := make([]gin.H, 0)
	for _, entry := range domain.VisibleEntries(localeCode) {
		list = append(list, gin.H{
			"id":          entry.ID,
			"name":        entry.Name,
			"emoji":       entry.Emoji,
			"rag_enabled": entry.RAGEnabled,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"default_domain": domain.DefaultID(),
		"locale":         localeCode,
		"domains":        list,
	})
}

// Branding handles GET /branding.
func Branding(c *gin.Context) {
	s := requireServices()
	loc := s.Locale(c)
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"locale":   loc,
		"branding": locale.BrandingForLocale(loc),
	})
}

// Onboarding handles GET /onboarding.
func Onboarding(c *gin.Context) {
	s := requireServices()
	loc := s.Locale(c)
	domainID, err := s.NormalizeDomainID(domain.IDFromQuery(c))
	if err != nil {
		JSONError(c, http.StatusBadRequest, err)
		return
	}
	questions := locale.OnboardingForDomainLocale(domainID, loc)
	if questions == nil {
		questions = []string{}
	}
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"locale":    loc,
		"domain_id": domainID,
		"questions": questions,
	})
}
