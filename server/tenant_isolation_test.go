package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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
	acmeDoc := filepath.Join(dir, "acme", "default", "secret_acme.txt")
	betaDoc := filepath.Join(dir, "beta", "default", "secret_beta.txt")
	if err := os.MkdirAll(filepath.Dir(acmeDoc), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(betaDoc), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(acmeDoc, []byte("acme-only"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(betaDoc, []byte("beta-only"), 0o644); err != nil {
		t.Fatal(err)
	}

	config = &Config{DataDir: dir, DefaultTenantID: "default"}
	acmeDir := kbDataDir("acme", "default")
	betaDir := kbDataDir("beta", "default")
	if acmeDir == betaDir {
		t.Fatalf("tenants must not share KB path: %s", acmeDir)
	}

	acmeListing, err := os.ReadDir(acmeDir)
	if err != nil {
		t.Fatal(err)
	}
	betaListing, err := os.ReadDir(betaDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(acmeListing) != 1 || acmeListing[0].Name() != "secret_acme.txt" {
		t.Fatalf("acme listing=%v", acmeListing)
	}
	if len(betaListing) != 1 || betaListing[0].Name() != "secret_beta.txt" {
		t.Fatalf("beta listing=%v", betaListing)
	}
	if !filepath.IsAbs(acmeDir) || !filepath.IsAbs(betaDir) {
		t.Fatal("expected absolute KB paths")
	}
	if filepath.Dir(acmeDir) == filepath.Dir(betaDir) {
		// same domain leaf name is OK; parents must differ by tenant
		if filepath.Base(filepath.Dir(acmeDir)) == filepath.Base(filepath.Dir(betaDir)) {
			t.Fatalf("tenant parent collision: acme=%s beta=%s", acmeDir, betaDir)
		}
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
