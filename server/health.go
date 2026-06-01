package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GET /health: статус сервиса и доступность PostgreSQL.
func handleHealthCheck(c *gin.Context) {
	payload := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
	}
	if chatStore != nil && chatStore.pool != nil {
		if err := chatStore.pool.Ping(c.Request.Context()); err != nil {
			payload["status"] = "degraded"
			payload["database"] = "unreachable"
		} else {
			payload["database"] = "ok"
		}
	}
	c.JSON(http.StatusOK, payload)
}
