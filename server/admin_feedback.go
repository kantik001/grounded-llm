package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /admin/feedback — сводка 👍/👎 по сообщениям ассистента.
func handleAdminFeedbackSummary(c *gin.Context) {
	if chatStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "БД недоступна"})
		return
	}
	summary, err := chatStore.FeedbackSummary(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "feedback": summary})
}
