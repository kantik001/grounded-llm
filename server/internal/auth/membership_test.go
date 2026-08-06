package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/config"
	"grounded_llm_server/internal/tenant"
)

func TestTelegramMembershipEnforceWithoutStore(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("TENANT_MEMBERSHIP_ENFORCE", "1")
	t.Setenv("TENANT_MEMBERSHIP_AUTO_DEFAULT", "0")
	tenant.InitAllowlist(&config.Config{DefaultTenantID: "default"})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	err := tenant.BindTelegramMembership(c, 42)
	if err == nil {
		t.Fatal("expected error when enforce on and store unavailable")
	}
}

func TestTelegramMembershipOffWithoutStore(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("TENANT_MEMBERSHIP_ENFORCE", "0")
	tenant.InitAllowlist(&config.Config{DefaultTenantID: "default"})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	if err := tenant.BindTelegramMembership(c, 42); err != nil {
		t.Fatalf("off mode must pass without store: %v", err)
	}
}
