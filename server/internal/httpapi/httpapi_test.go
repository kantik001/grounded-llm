package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/httpapi"
)

func TestCORS_Preflight(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(httpapi.CORS([]string{"http://localhost:3000"}))
	r.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/ping", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("preflight status = %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("Allow-Origin = %q", got)
	}
}

func TestRegisterPublic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	called := false
	httpapi.RegisterPublic(r, httpapi.PublicHandlers{
		Health:     func(c *gin.Context) { called = true; c.Status(http.StatusOK) },
		Ready:      func(c *gin.Context) { c.Status(http.StatusOK) },
		Metrics:    func(c *gin.Context) { c.Status(http.StatusOK) },
		Domains:    func(c *gin.Context) { c.Status(http.StatusOK) },
		Onboarding: func(c *gin.Context) { c.Status(http.StatusOK) },
		Branding:   func(c *gin.Context) { c.Status(http.StatusOK) },
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	if w.Code != http.StatusOK || !called {
		t.Fatalf("health not mounted: code=%d called=%v", w.Code, called)
	}
}

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(httpapi.RequestID())
	r.GET("/x", func(c *gin.Context) {
		if httpapi.CtxRequestID(c) == "" {
			t.Fatal("missing request id")
		}
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))
	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("missing response header")
	}
}

func TestOpenAPIEmbedded(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/openapi.json", httpapi.OpenAPI)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/openapi.json", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
	if len(w.Body.Bytes()) < 100 {
		t.Fatal("expected embedded openapi body")
	}
}
