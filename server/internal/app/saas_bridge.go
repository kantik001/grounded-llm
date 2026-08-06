package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/saas"
	"grounded_llm_server/internal/tenant"
)

func init() {
	saas.BindConfig(func() *Config { return config })
	saas.BindHost(saasHost{})
}

func loadPlans() error {
	return saas.LoadPlans()
}

func saasSignupEnabled() bool {
	return saas.SignupEnabled()
}

func stripeWebhookSecret() string {
	return saas.StripeWebhookSecret()
}

func stripeSecretKey() string {
	return saas.StripeSecretKey()
}

func registerSaaSRoutes(router *gin.Engine, rl *rateLimiter) {
	var lim saas.RateLimitMiddleware
	if rl != nil {
		lim = saas.RateLimitMiddleware(rateLimitMiddleware(rl))
	}
	saas.RegisterRoutes(router, lim)
}

type saasHost struct{}

func (saasHost) NormalizeTenantID(raw string) string { return normalizeTenantID(raw) }

func (saasHost) TenantRegistered(tenantID string) bool {
	return tenant.OnAllowlist(normalizeTenantID(tenantID))
}

func (saasHost) TenantIDFree(candidate string) bool {
	return !tenant.OnAllowlist(normalizeTenantID(candidate))
}

func (saasHost) RegisterTenant(entry saas.RegistryEntry) error {
	return registerTenantEntry(tenant.RegistryEntry{
		TenantID:  entry.TenantID,
		OrgName:   entry.OrgName,
		Email:     entry.Email,
		Plan:      entry.Plan,
		CreatedAt: entry.CreatedAt,
	})
}

func (saasHost) NewTenantEntry(tenantID, orgName, email, plan string) saas.RegistryEntry {
	e := newTenantRegistryEntry(tenantID, orgName, email, plan)
	return saas.RegistryEntry{
		TenantID:  e.TenantID,
		OrgName:   e.OrgName,
		Email:     e.Email,
		Plan:      e.Plan,
		CreatedAt: e.CreatedAt,
	}
}

func (saasHost) ApplyPlanQuotas(tenantID string, limits saas.QuotaLimits) error {
	return upsertTenantQuota(tenantID, TenantQuotaLimits{
		MessagesPerDay: limits.MessagesPerDay,
		StorageMB:      limits.StorageMB,
		MaxDomains:     limits.MaxDomains,
	})
}

func (saasHost) UpdateTenantPlan(tenantID, plan string) error {
	return updateTenantPlan(tenantID, plan)
}

func (saasHost) UpdateTenantStripeCustomer(tenantID, customerID string) error {
	return updateTenantStripeCustomer(tenantID, customerID)
}

func (saasHost) TenantEmail(tenantID string) string {
	return tenantSignupEmail(tenantID)
}

func (saasHost) ProvisionDataDir(tenantID string) error {
	if config == nil {
		return fmt.Errorf("server not configured")
	}
	domainID := defaultDomainID()
	dir := filepath.Join(config.DataDir, normalizeTenantID(tenantID), domainID)
	return os.MkdirAll(dir, 0o755)
}

func (saasHost) AdminProvisioningEnabled() bool {
	return saasProvisionAdmin()
}

func (saasHost) ProvisionAdminUser(tenantID string) (username, password string, err error) {
	return provisionSignupAdminUser(tenantID)
}

func (saasHost) DefaultDomainID() string {
	return defaultDomainID()
}

func (saasHost) ClaimStripeEvent(eventID, eventType string) (bool, error) {
	return tenant.ClaimStripeEvent(eventID, eventType)
}

func (saasHost) TenantsRegistryConfigured() bool {
	return tenant.TenantsRegistryConfigured()
}
