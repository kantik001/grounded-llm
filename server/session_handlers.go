package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type newSessionRequest struct {
	DomainID string `json:"domain_id"`
	CropID   string `json:"crop_id"` // legacy
}

func handleNewSession(c *gin.Context) {
	var req newSessionRequest
	_ = c.ShouldBindJSON(&req)

	domainID := strings.TrimSpace(req.DomainID)
	if domainID == "" {
		domainID = strings.TrimSpace(req.CropID)
	}
	if domainID == "" {
		domainID = defaultDomainID()
	} else {
		var err error
		domainID, err = normalizeDomainID(domainID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}
	}

	tgUser, err := ctxTelegramUser(c)
	if err != nil {
		jsonError(c, http.StatusUnauthorized, err)
		return
	}
	ctx := c.Request.Context()
	userID, err := chatStore.UpsertUser(ctx, tgUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Ошибка пользователя"})
		return
	}
	sid, err := chatStore.CreateSession(ctx, userID, domainID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Ошибка создания сессии"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"session_id": sid,
		"domain_id":  domainID,
	})
}

func handleHistory(c *gin.Context) {
	id := strings.TrimSpace(c.Query("session_id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Нужен session_id"})
		return
	}
	tgUser, err := ctxTelegramUser(c)
	if err != nil {
		jsonError(c, http.StatusUnauthorized, err)
		return
	}
	msgs, err := chatStore.ListMessages(c.Request.Context(), id, tgUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Ошибка базы данных"})
		return
	}
	domainID, _ := chatStore.SessionDomainID(c.Request.Context(), id, tgUser.ID)
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"session_id": id,
		"domain_id":  domainID,
		"messages":   msgs,
	})
}
