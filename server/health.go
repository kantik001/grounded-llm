package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GET /health — liveness (process up; DB ping is best-effort).
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

// GET /ready — readiness for load balancers / Kubernetes (DB + Python RAG when required).
func handleReadiness(c *gin.Context) {
	ctx := c.Request.Context()
	checks := gin.H{}
	ready := true

	if chatStore == nil || chatStore.pool == nil {
		checks["database"] = "unconfigured"
		ready = false
	} else if err := chatStore.pool.Ping(ctx); err != nil {
		checks["database"] = "unreachable"
		ready = false
	} else {
		checks["database"] = "ok"
	}

	if config.RAGMock {
		checks["python_rag"] = "mock"
	} else {
		ragStatus, err := probePythonRAG(ctx)
		if err != nil {
			checks["python_rag"] = ragStatus
			ready = false
		} else {
			checks["python_rag"] = ragStatus
		}
	}

	payload := gin.H{
		"status": map[bool]string{true: "ready", false: "not_ready"}[ready],
		"checks": checks,
	}
	if !ready {
		c.JSON(http.StatusServiceUnavailable, payload)
		return
	}
	c.JSON(http.StatusOK, payload)
}

func probePythonRAG(ctx context.Context) (string, error) {
	base := pythonServiceBaseURL()
	if base == "" {
		return "unconfigured", fmt.Errorf("PYTHON_BASE_URL not set")
	}
	url := base + "/ready"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "error", err
	}
	setPythonServiceHeaders(req)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "unreachable", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("http_%d", resp.StatusCode), fmt.Errorf("python /ready returned %d", resp.StatusCode)
	}
	return "ok", nil
}
