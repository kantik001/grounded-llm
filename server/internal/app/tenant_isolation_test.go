package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestTenantAllowlistRejectsUnknown(t *testing.T) {
	t.Setenv("ALLOWED_TENANTS", "default,acme")
	config = &Config{DefaultTenantID: "default", DataDir: t.TempDir()}
	initTenantConfig(config)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(tenantMiddleware(config))
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"tenant": ctxTenantID(c)})
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("X-Tenant-ID", "evil")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400 for unknown tenant, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestTenantAllowlistAcceptsKnown(t *testing.T) {
	t.Setenv("ALLOWED_TENANTS", "default,acme")
	config = &Config{DefaultTenantID: "default", DataDir: t.TempDir()}
	initTenantConfig(config)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(tenantMiddleware(config))
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"tenant": ctxTenantID(c)})
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("X-Tenant-ID", "AcMe")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["tenant"] != "acme" {
		t.Fatalf("tenant=%v", body["tenant"])
	}
}

func TestKbDataDirIsolatesTenants(t *testing.T) {
	dir := t.TempDir()
	config = &Config{DataDir: dir, DefaultTenantID: "default"}
	acmeDir := kbDataDir("acme", "default")
	betaDir := kbDataDir("beta", "default")
	if acmeDir == betaDir {
		t.Fatalf("tenants must not share KB path: %s", acmeDir)
	}
	if acmeDir != filepath.Join(dir, "acme", "default") {
		t.Fatalf("unexpected acme dir: %s", acmeDir)
	}
	if betaDir != filepath.Join(dir, "beta", "default") {
		t.Fatalf("unexpected beta dir: %s", betaDir)
	}
}

func TestAdminTenantIDDefaultsAndOverrides(t *testing.T) {
	config = &Config{DefaultTenantID: "default"}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin/articles", func(c *gin.Context) {
		c.String(http.StatusOK, adminTenantID(c))
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/admin/articles", nil))
	if w.Body.String() != "default" {
		t.Fatalf("default tenant=%q", w.Body.String())
	}

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/admin/articles?tenant_id=Acme", nil))
	if w2.Body.String() != "acme" {
		t.Fatalf("override tenant=%q", w2.Body.String())
	}
}
