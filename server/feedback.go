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

func handleFeedback(c *gin.Context) {
	tgUser, err := ctxTelegramUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": err.Error()})
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
	if err := chatStore.SaveMessageFeedback(ctx, tgUser.ID, req.MessageID, req.Rating); err != nil {
		if err == errSessionNotFound {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Message not found"})
			return
		}
		log.Printf("SaveMessageFeedback: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to save rating"})
		return
	}
	_ = chatStore.LogEvent(ctx, tgUser.ID, "message_feedback", map[string]any{
		"message_id": req.MessageID,
		"rating":     req.Rating,
		"session_id": req.SessionID,
	})
	c.JSON(http.StatusOK, gin.H{"success": true, "message_id": req.MessageID, "rating": req.Rating})
}
