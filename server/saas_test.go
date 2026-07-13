package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestLoadPlans(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "plans.yaml")
	body := `version: 1
currency: USD
plans:
  starter:
    label: Starter
    price_monthly: 0
    quotas:
      messages_per_day: 100
      storage_mb: 256
      domains: 1
`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PLANS_FILE", path)
	if err := loadPlans(); err != nil {
		t.Fatal(err)
	}
	p, ok := planByID("starter")
	if !ok || p.Label != "Starter" {
		t.Fatalf("plan: %+v ok=%v", p, ok)
	}
	lim := planQuotasToLimits(p.Quotas)
	if lim.MessagesPerDay != 100 || lim.StorageMB != 256 || lim.MaxDomains != 1 {
		t.Fatalf("limits: %+v", lim)
	}
}

func TestSignupDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("SAAS_SIGNUP_ENABLED", "false")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/signup", strings.NewReader(`{"org_name":"Acme","email":"a@b.com","plan":"starter"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	handleSignup(c)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestSignupCreatesTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	registry := filepath.Join(dir, "tenants.json")
	quotas := filepath.Join(dir, "tenant_quotas.json")
	plans := filepath.Join(dir, "plans.yaml")
	dataDir := filepath.Join(dir, "data")

	planYAML := `version: 1
currency: USD
plans:
  starter:
    label: Starter
    price_monthly: 0
    quotas:
      messages_per_day: 200
      storage_mb: 512
      domains: 1
`
	for path, content := range map[string]string{plans: planYAML} {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	t.Setenv("SAAS_SIGNUP_ENABLED", "true")
	t.Setenv("TENANTS_REGISTRY_FILE", registry)
	t.Setenv("TENANT_QUOTAS_FILE", quotas)
	t.Setenv("PLANS_FILE", plans)

	config = &Config{DataDir: dataDir, DefaultTenantID: "default"}
	allowedTenants = map[string]struct{}{"default": {}}
	tenantQuotaRegistry = make(map[string]TenantQuotaLimits)
	domainCatalog = domainsFile{
		DefaultDomain: "default",
		Domains: map[string]DomainInfo{
			"default": {Name: "Default"},
		},
	}

	if err := loadPlans(); err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/signup", strings.NewReader(`{"org_name":"Acme Corp","email":"admin@acme.com","plan":"starter"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	handleSignup(c)
	if w.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	tenantID, _ := resp["tenant_id"].(string)
	if tenantID == "" {
		t.Fatalf("missing tenant_id: %v", resp)
	}
	if _, ok := allowedTenants[tenantID]; !ok {
		t.Fatalf("tenant not allowed: %s", tenantID)
	}
	lim, ok := quotaLimitsForTenant(tenantID)
	if !ok || lim.MessagesPerDay != 200 {
		t.Fatalf("quotas: %+v ok=%v", lim, ok)
	}
}

func TestVerifyStripeSignature(t *testing.T) {
	secret := "whsec_test"
	payload := []byte(`{"id":"evt_1"}`)
	ts := time.Now().Unix()
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(strconv.FormatInt(ts, 10)))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write(payload)
	sig := hex.EncodeToString(mac.Sum(nil))
	header := "t=" + strconv.FormatInt(ts, 10) + ",v1=" + sig
	if err := verifyStripeSignature(payload, header, secret, 5*time.Minute); err != nil {
		t.Fatalf("verify: %v", err)
	}
	if err := verifyStripeSignature(payload, header, secret+"x", 5*time.Minute); err == nil {
		t.Fatal("expected mismatch")
	}
}

func TestStripeWebhookUpdatesPlan(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	registry := filepath.Join(dir, "tenants.json")
	quotas := filepath.Join(dir, "tenant_quotas.json")
	plans := filepath.Join(dir, "plans.yaml")

	planYAML := `version: 1
currency: USD
plans:
  starter:
    label: Starter
    price_monthly: 0
    quotas:
      messages_per_day: 200
      storage_mb: 512
      domains: 1
  business:
    label: Business
    price_monthly: 299
    quotas:
      messages_per_day: 5000
      storage_mb: 10240
      domains: 10
`
	if err := os.WriteFile(plans, []byte(planYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	entry := `[{"tenant_id":"acme-abc","org_name":"Acme","email":"a@b.com","plan":"starter","created_at":"2026-01-01T00:00:00Z"}]`
	if err := os.WriteFile(registry, []byte(entry), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("STRIPE_WEBHOOK_SECRET", "whsec_test")
	t.Setenv("TENANTS_REGISTRY_FILE", registry)
	t.Setenv("TENANT_QUOTAS_FILE", quotas)
	t.Setenv("PLANS_FILE", plans)

	config = &Config{DefaultTenantID: "default"}
	allowedTenants = map[string]struct{}{"acme-abc": {}}
	tenantQuotaRegistry = make(map[string]TenantQuotaLimits)
	loadTenantRegistry()
	if err := loadPlans(); err != nil {
		t.Fatal(err)
	}

	event := map[string]any{
		"type": "checkout.session.completed",
		"data": map[string]any{
			"object": map[string]any{
				"customer": "cus_123",
				"metadata": map[string]string{
					"tenant_id": "acme-abc",
					"plan":      "business",
				},
			},
		},
	}
	payload, _ := json.Marshal(event)
	ts := time.Now().Unix()
	mac := hmac.New(sha256.New, []byte("whsec_test"))
	_, _ = mac.Write([]byte(strconv.FormatInt(ts, 10)))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write(payload)
	sig := hex.EncodeToString(mac.Sum(nil))
	header := "t=" + strconv.FormatInt(ts, 10) + ",v1=" + sig

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/billing/stripe/webhook", strings.NewReader(string(payload)))
	c.Request.Header.Set("Stripe-Signature", header)
	handleStripeWebhook(c)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	lim, ok := quotaLimitsForTenant("acme-abc")
	if !ok || lim.MessagesPerDay != 5000 {
		t.Fatalf("quotas: %+v ok=%v", lim, ok)
	}
}

func TestProvisionSignupAdminUser(t *testing.T) {
	dir := t.TempDir()
	usersFile := filepath.Join(dir, "admin_users.json")
	t.Setenv("ADMIN_USERS_FILE", usersFile)

	adminUserRegistry = make(map[string]adminUserRecord)
	user, pass, err := provisionSignupAdminUser("acme-demo")
	if err != nil {
		t.Fatal(err)
	}
	if user != "acme-demo-admin" || len(pass) < 8 {
		t.Fatalf("user=%q pass len=%d", user, len(pass))
	}
	if _, ok := adminUserRegistry[user]; !ok {
		t.Fatal("user not in registry")
	}
	if _, err := os.Stat(usersFile); err != nil {
		t.Fatal(err)
	}
}

func TestStripeCheckoutMock(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/checkout/sessions" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{
			"id":  "cs_test",
			"url": "https://checkout.stripe.test/session",
		})
	}))
	defer mock.Close()

	dir := t.TempDir()
	registry := filepath.Join(dir, "tenants.json")
	plans := filepath.Join(dir, "plans.yaml")
	planYAML := `version: 1
currency: USD
plans:
  business:
    label: Business
    price_monthly: 299
    stripe_price_id: price_test
    quotas:
      messages_per_day: 5000
      storage_mb: 10240
      domains: 10
`
	if err := os.WriteFile(plans, []byte(planYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	entry := `[{"tenant_id":"acme-abc","org_name":"Acme","email":"a@b.com","plan":"starter","created_at":"2026-01-01T00:00:00Z"}]`
	if err := os.WriteFile(registry, []byte(entry), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("SAAS_SIGNUP_ENABLED", "true")
	t.Setenv("STRIPE_SECRET_KEY", "sk_test_x")
	t.Setenv("STRIPE_API_BASE", mock.URL)
	t.Setenv("TENANTS_REGISTRY_FILE", registry)
	t.Setenv("PLANS_FILE", plans)

	allowedTenants = map[string]struct{}{"acme-abc": {}}
	if err := loadPlans(); err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := `{"tenant_id":"acme-abc","plan":"business","email":"a@b.com"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/billing/stripe/checkout", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	handleStripeCheckout(c)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["checkout_url"] != "https://checkout.stripe.test/session" {
		t.Fatalf("checkout_url=%v", resp["checkout_url"])
	}
}
