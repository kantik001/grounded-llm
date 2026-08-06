package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/config"
	"grounded_llm_server/internal/tenant"
)

func setupBindingRouter(t *testing.T, cfg *config.Config) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(tenant.Middleware(cfg))
	r.Use(CombinedMiddleware(cfg))
	r.GET("/whoami", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"tenant": tenant.CtxID(c)})
	})
	return r
}

func TestAPIKeyBoundTenantRejectsForeignHeader(t *testing.T) {
	t.Setenv("ALLOWED_TENANTS", "default,acme,beta")
	cfg := &config.Config{DefaultTenantID: "default"}
	tenant.InitAllowlist(cfg)
	Registry = map[string]KeyRecord{
		HashKey("acme-key"): {Label: "acme", Roles: []string{RoleChatOnly}, Tenant: "acme"},
	}
	r := setupBindingRouter(t, cfg)

	req := httptest.NewRequest(http.MethodGet, "/whoami", nil)
	req.Header.Set(HeaderAPIKey, "acme-key")
	req.Header.Set("X-Tenant-ID", "beta")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403 for cross-tenant key use, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestAPIKeyBoundTenantPinsContext(t *testing.T) {
	t.Setenv("ALLOWED_TENANTS", "default,acme")
	cfg := &config.Config{DefaultTenantID: "default"}
	tenant.InitAllowlist(cfg)
	Registry = map[string]KeyRecord{
		HashKey("acme-key"): {Label: "acme", Roles: []string{RoleChatOnly}, Tenant: "acme"},
	}
	r := setupBindingRouter(t, cfg)

	req := httptest.NewRequest(http.MethodGet, "/whoami", nil)
	req.Header.Set(HeaderAPIKey, "acme-key")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if body := w.Body.String(); body != `{"tenant":"acme"}` {
		t.Fatalf("body=%s", body)
	}
}

func TestUnboundAPIKeyDefaultsToDefaultTenant(t *testing.T) {
	t.Setenv("ALLOWED_TENANTS", "default,acme")
	cfg := &config.Config{DefaultTenantID: "default"}
	tenant.InitAllowlist(cfg)
	Registry = map[string]KeyRecord{
		HashKey("plain"): {Label: "plain", Roles: []string{RoleChatOnly}},
	}
	r := setupBindingRouter(t, cfg)

	// No explicit tenant → default.
	req := httptest.NewRequest(http.MethodGet, "/whoami", nil)
	req.Header.Set(HeaderAPIKey, "plain")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK || w.Body.String() != `{"tenant":"default"}` {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}

	// Explicit foreign tenant → rejected (key is scoped to default).
	req2 := httptest.NewRequest(http.MethodGet, "/whoami", nil)
	req2.Header.Set(HeaderAPIKey, "plain")
	req2.Header.Set("X-Tenant-ID", "acme")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d body=%s", w2.Code, w2.Body.String())
	}
}

func TestWildcardAPIKeyAllowsAnyAllowlistedTenant(t *testing.T) {
	t.Setenv("ALLOWED_TENANTS", "default,acme")
	cfg := &config.Config{DefaultTenantID: "default"}
	tenant.InitAllowlist(cfg)
	Registry = map[string]KeyRecord{
		HashKey("gw"): {Label: "gateway", Roles: []string{RoleChatOnly}, Tenant: "*"},
	}
	r := setupBindingRouter(t, cfg)

	req := httptest.NewRequest(http.MethodGet, "/whoami", nil)
	req.Header.Set(HeaderAPIKey, "gw")
	req.Header.Set("X-Tenant-ID", "acme")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK || w.Body.String() != `{"tenant":"acme"}` {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}
