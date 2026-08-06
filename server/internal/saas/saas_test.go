package saas

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
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type testHost struct {
	mu         sync.Mutex
	allowed    map[string]struct{}
	registry   []RegistryEntry
	quotas     map[string]QuotaLimits
	dataDir    string
	seenEvents map[string]struct{}
}

func (h *testHost) NormalizeTenantID(raw string) string {
	return strings.TrimSpace(strings.ToLower(raw))
}

func (h *testHost) TenantRegistered(tenantID string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	_, ok := h.allowed[h.NormalizeTenantID(tenantID)]
	return ok
}

func (h *testHost) TenantIDFree(candidate string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	_, ok := h.allowed[h.NormalizeTenantID(candidate)]
	return !ok
}

func (h *testHost) TenantsRegistryConfigured() bool { return true }

func (h *testHost) RegisterTenant(entry RegistryEntry) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	id := h.NormalizeTenantID(entry.TenantID)
	h.registry = append(h.registry, entry)
	h.allowed[id] = struct{}{}
	return nil
}

func (h *testHost) NewTenantEntry(tenantID, orgName, email, plan string) RegistryEntry {
	return RegistryEntry{
		TenantID:  h.NormalizeTenantID(tenantID),
		OrgName:   orgName,
		Email:     email,
		Plan:      plan,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func (h *testHost) ApplyPlanQuotas(tenantID string, limits QuotaLimits) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.quotas[h.NormalizeTenantID(tenantID)] = limits
	return nil
}

func (h *testHost) UpdateTenantPlan(tenantID, plan string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	id := h.NormalizeTenantID(tenantID)
	for i, e := range h.registry {
		if h.NormalizeTenantID(e.TenantID) == id {
			h.registry[i].Plan = plan
			return nil
		}
	}
	return errHostNotConfigured
}

func (h *testHost) UpdateTenantStripeCustomer(tenantID, customerID string) error {
	return nil
}

func (h *testHost) TenantEmail(tenantID string) string {
	h.mu.Lock()
	defer h.mu.Unlock()
	id := h.NormalizeTenantID(tenantID)
	for _, e := range h.registry {
		if h.NormalizeTenantID(e.TenantID) == id {
			return e.Email
		}
	}
	return ""
}

func (h *testHost) ProvisionDataDir(tenantID string) error {
	dir := filepath.Join(h.dataDir, h.NormalizeTenantID(tenantID), h.DefaultDomainID())
	return os.MkdirAll(dir, 0o755)
}

func (h *testHost) AdminProvisioningEnabled() bool { return false }

func (h *testHost) ProvisionAdminUser(tenantID string) (string, string, error) {
	return "", "", nil
}

func (h *testHost) DefaultDomainID() string { return "default" }

func (h *testHost) ClaimStripeEvent(eventID, eventType string) (bool, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.seenEvents == nil {
		h.seenEvents = make(map[string]struct{})
	}
	if eventID == "" {
		return true, nil
	}
	if _, ok := h.seenEvents[eventID]; ok {
		return false, nil
	}
	h.seenEvents[eventID] = struct{}{}
	return true, nil
}

func setupTestHost(t *testing.T) *testHost {
	t.Helper()
	h := &testHost{
		allowed: map[string]struct{}{"default": {}},
		quotas:  make(map[string]QuotaLimits),
		dataDir: t.TempDir(),
	}
	BindHost(h)
	return h
}

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
	if err := LoadPlans(); err != nil {
		t.Fatal(err)
	}
	p, ok := PlanByID("starter")
	if !ok || p.Label != "Starter" {
		t.Fatalf("plan: %+v ok=%v", p, ok)
	}
	lim := quotasToLimits(p.Quotas)
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
`
	if err := os.WriteFile(plans, []byte(planYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("SAAS_SIGNUP_ENABLED", "true")
	t.Setenv("PLANS_FILE", plans)

	h := setupTestHost(t)

	if err := LoadPlans(); err != nil {
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
	if !h.TenantRegistered(tenantID) {
		t.Fatalf("tenant not allowed: %s", tenantID)
	}
	lim, ok := h.quotas[tenantID]
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
	if err := VerifyStripeSignature(payload, header, secret, 5*time.Minute); err != nil {
		t.Fatalf("verify: %v", err)
	}
	if err := VerifyStripeSignature(payload, header, secret+"x", 5*time.Minute); err == nil {
		t.Fatal("expected mismatch")
	}
}

func TestStripeWebhookUpdatesPlan(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
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

	t.Setenv("STRIPE_WEBHOOK_SECRET", "whsec_test")
	t.Setenv("PLANS_FILE", plans)

	h := setupTestHost(t)
	h.registry = []RegistryEntry{{
		TenantID: "acme-abc",
		OrgName:  "Acme",
		Email:    "a@b.com",
		Plan:     "starter",
	}}
	h.allowed["acme-abc"] = struct{}{}

	if err := LoadPlans(); err != nil {
		t.Fatal(err)
	}

	event := map[string]any{
		"id":   "evt_test_1",
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
	lim, ok := h.quotas["acme-abc"]
	if !ok || lim.MessagesPerDay != 5000 {
		t.Fatalf("quotas: %+v ok=%v", lim, ok)
	}

	// Replay same event id — must stay idempotent (no double apply / still 200).
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest(http.MethodPost, "/api/v1/billing/stripe/webhook", strings.NewReader(string(payload)))
	c2.Request.Header.Set("Stripe-Signature", header)
	handleStripeWebhook(c2)
	if w2.Code != http.StatusOK {
		t.Fatalf("replay status=%d body=%s", w2.Code, w2.Body.String())
	}
	var replay map[string]any
	_ = json.Unmarshal(w2.Body.Bytes(), &replay)
	if replay["duplicate"] != true {
		t.Fatalf("expected duplicate=true on replay, got %v", replay)
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

	t.Setenv("SAAS_SIGNUP_ENABLED", "true")
	t.Setenv("STRIPE_SECRET_KEY", "sk_test_x")
	t.Setenv("STRIPE_API_BASE", mock.URL)
	t.Setenv("PLANS_FILE", plans)

	h := setupTestHost(t)
	h.registry = []RegistryEntry{{
		TenantID: "acme-abc",
		OrgName:  "Acme",
		Email:    "a@b.com",
		Plan:     "starter",
	}}
	h.allowed["acme-abc"] = struct{}{}

	if err := LoadPlans(); err != nil {
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
