package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Health is GET /health — liveness (DB ping is best-effort).
func Health(c *gin.Context) {
	s := requireServices()
	payload := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
	}
	st := s.Store()
	if st != nil && st.Pool != nil {
		if err := st.Pool.Ping(c.Request.Context()); err != nil {
			payload["status"] = "degraded"
			payload["database"] = "unreachable"
		} else {
			payload["database"] = "ok"
		}
	}
	c.JSON(http.StatusOK, payload)
}

// Ready is GET /ready — readiness (DB + Python RAG when required).
func Ready(c *gin.Context) {
	s := requireServices()
	ctx := c.Request.Context()
	checks := gin.H{}
	ready := true

	st := s.Store()
	if st == nil || st.Pool == nil {
		checks["database"] = "unconfigured"
		ready = false
	} else if err := st.Pool.Ping(ctx); err != nil {
		checks["database"] = "unreachable"
		ready = false
	} else {
		checks["database"] = "ok"
	}

	cfg := s.Config()
	if cfg != nil && cfg.RAGMock {
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
	cfg := requireServices().Config()
	if cfg == nil {
		return "unconfigured", fmt.Errorf("config not loaded")
	}
	base := strings.TrimRight(cfg.PythonBaseURL, "/")
	if base == "" {
		return "unconfigured", fmt.Errorf("PYTHON_BASE_URL not set")
	}
	url := base + "/ready"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "error", err
	}
	if cfg.RAGServiceToken != "" {
		req.Header.Set("X-RAG-Service-Token", cfg.RAGServiceToken)
	}
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
