package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type feedbackRequest struct {
	SessionID string `json:"session_id"`
	MessageID int64  `json:"message_id"`
	Rating    int    `json:"rating"`
}

// handleFeedback — 👍/👎 на ответ ассистента.
func handleFeedback(c *gin.Context) {
	tgUser, err := ctxTelegramUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": err.Error()})
		return
	}
	var req feedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Нужны session_id, message_id, rating (1 или -1)"})
		return
	}
	if req.MessageID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Некорректный message_id"})
		return
	}
	ctx := c.Request.Context()
	if err := chatStore.SaveMessageFeedback(ctx, tgUser.ID, req.MessageID, req.Rating); err != nil {
		if err == errSessionNotFound {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Сообщение не найдено"})
			return
		}
		log.Printf("SaveMessageFeedback: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Не удалось сохранить оценку"})
		return
	}
	_ = chatStore.LogEvent(ctx, tgUser.ID, "message_feedback", map[string]any{
		"message_id": req.MessageID,
		"rating":     req.Rating,
		"session_id": req.SessionID,
	})
	c.JSON(http.StatusOK, gin.H{"success": true, "message_id": req.MessageID, "rating": req.Rating})
}

// Пишет событие аналитики в Postgres для текущего Telegram-пользователя.
func logAnalytics(c *gin.Context, eventType string, payload map[string]any) {
	tgUser, err := ctxTelegramUser(c)
	if err != nil || chatStore == nil {
		return
	}
	if err := chatStore.LogEvent(c.Request.Context(), tgUser.ID, eventType, payload); err != nil {
		log.Printf("LogEvent %s: %v", eventType, err)
	}
}
