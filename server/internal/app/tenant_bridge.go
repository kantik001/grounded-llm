package app

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/saas"
	"grounded_llm_server/internal/tenant"
)

func init() {
	tenant.BindConfig(func() *Config { return config })
	tenant.BindStore(func() *ChatStore { return chatStore })
	tenant.BindApplyPlanQuotas(applyPlanQuotasFromPlans)
}

const ctxKeyTenantID = tenant.CtxKeyTenantID

type (
	TenantQuotaLimits = tenant.QuotaLimits
	TenantQuotaUsage  = tenant.QuotaUsage
	TenantQuotaStatus = tenant.QuotaStatus
)

func initTenantConfig(cfg *Config) {
	tenant.InitAllowlist(cfg)
}

func normalizeTenantID(raw string) string {
	return tenant.NormalizeTenantID(raw)
}

func resolveTenantID(c *gin.Context, cfg *Config) (string, error) {
	return tenant.ResolveID(c, cfg)
}

func tenantMiddleware(cfg *Config) gin.HandlerFunc {
	return tenant.Middleware(cfg)
}

func ctxTenantID(c *gin.Context) string {
	return tenant.CtxID(c)
}

func kbDataDir(tenantID, domainID string) string {
	return tenant.KBDataDir(tenantID, domainID)
}

func adminTenantID(c *gin.Context) string {
	return tenant.AdminID(c)
}

func tenantsRegistryPath() string {
	return tenant.RegistryPath()
}

func loadTenantRegistry() {
	tenant.LoadRegistry()
}

func registerTenantEntry(entry tenant.RegistryEntry) error {
	return tenant.RegisterEntry(entry)
}

func newTenantRegistryEntry(tenantID, orgName, email, plan string) tenant.RegistryEntry {
	return tenant.NewRegistryEntry(tenantID, orgName, email, plan)
}

func updateTenantPlan(tenantID, plan string) error {
	return tenant.UpdatePlan(tenantID, plan)
}

func updateTenantStripeCustomer(tenantID, customerID string) error {
	return tenant.UpdateStripeCustomer(tenantID, customerID)
}

func loadTenantQuotas() {
	tenant.LoadQuotas()
}

func quotaLimitsForTenant(tenantID string) (TenantQuotaLimits, bool) {
	return tenant.QuotaLimitsFor(tenantID)
}

func upsertTenantQuota(tenantID string, limits TenantQuotaLimits) error {
	return tenant.UpsertQuota(tenantID, limits)
}

func applyPlanQuotas(tenantID, planID string) error {
	return tenant.ApplyPlanQuotas(tenantID, planID)
}

func applyPlanQuotasFromPlans(tenantID, planID string) error {
	limits, err := saas.QuotaLimitsForPlan(planID)
	if err != nil {
		return err
	}
	return tenant.UpsertQuota(tenantID, TenantQuotaLimits{
		MessagesPerDay: limits.MessagesPerDay,
		StorageMB:      limits.StorageMB,
		MaxDomains:     limits.MaxDomains,
	})
}

func buildTenantQuotaStatus(ctx context.Context, tenantID string) (TenantQuotaStatus, error) {
	return tenant.BuildQuotaStatus(ctx, tenantID)
}

func checkMessageQuota(ctx context.Context, tenantID string) error {
	return tenant.CheckMessageQuota(ctx, tenantID)
}

func checkStorageQuota(tenantID string, additionalBytes int64) error {
	return tenant.CheckStorageQuota(tenantID, additionalBytes)
}

func checkDomainQuota(tenantID, domainID string) error {
	return tenant.CheckDomainQuota(tenantID, domainID)
}

func quotaErrorResponse(c *gin.Context, err error) {
	c.JSON(http.StatusTooManyRequests, gin.H{
		"success": false,
		"error":   err.Error(),
		"code":    "quota_exceeded",
	})
}

func tenantSignupEmail(tenantID string) string {
	return tenant.SignupEmail(tenantID)
}

func tenantIsAllowed(tenantID string) bool {
	return tenant.IsAllowed(tenantID)
}

func allowTenant(tenantID string) {
	tenant.AllowTenant(tenantID)
}
