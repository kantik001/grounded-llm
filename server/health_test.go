package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthCheck(t *testing.T) {
	config = &Config{}
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)
	handleHealthCheck(c)
	if w.Code != http.StatusOK {
		t.Fatalf("health status = %d", w.Code)
	}
}

func TestReadiness_RAGMock(t *testing.T) {
	config = &Config{RAGMock: true}
	chatStore = &ChatStore{}
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/ready", nil)
	handleReadiness(c)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("ready without DB should be 503, got %d", w.Code)
	}
}
