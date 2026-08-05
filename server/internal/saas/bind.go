package saas

import (
	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/config"
)

var lookupConfig = config.Current

// BindConfig overrides where this package reads process config (app sets app.config).
func BindConfig(fn func() *config.Config) {
	if fn != nil {
		lookupConfig = fn
	}
}

func cfg() *config.Config {
	return lookupConfig()
}

// QuotaLimits mirrors tenant quota caps applied after signup or billing events.
type QuotaLimits struct {
	MessagesPerDay int
	StorageMB      int
	MaxDomains     int
}

// RegistryEntry is a tenant row written to TENANTS_REGISTRY_FILE.
type RegistryEntry struct {
	TenantID  string
	OrgName   string
	Email     string
	Plan      string
	CreatedAt string
}

// Host wires tenant registry and provisioning helpers owned by app.
type Host interface {
	NormalizeTenantID(raw string) string
	TenantRegistered(tenantID string) bool
	TenantIDFree(candidate string) bool
	TenantsRegistryConfigured() bool
	RegisterTenant(entry RegistryEntry) error
	NewTenantEntry(tenantID, orgName, email, plan string) RegistryEntry
	ApplyPlanQuotas(tenantID string, limits QuotaLimits) error
	UpdateTenantPlan(tenantID, plan string) error
	UpdateTenantStripeCustomer(tenantID, customerID string) error
	TenantEmail(tenantID string) string
	ProvisionDataDir(tenantID string) error
	AdminProvisioningEnabled() bool
	ProvisionAdminUser(tenantID string) (username, password string, err error)
	DefaultDomainID() string
}

var host Host

// BindHost wires tenant helpers from the app layer.
func BindHost(h Host) {
	host = h
}

func normalizeTenantID(raw string) string {
	if host != nil {
		return host.NormalizeTenantID(raw)
	}
	return raw
}

func tenantRegistered(tenantID string) bool {
	if host != nil {
		return host.TenantRegistered(tenantID)
	}
	return false
}

func tenantIDFree(candidate string) bool {
	if host != nil {
		return host.TenantIDFree(candidate)
	}
	return true
}

func tenantsRegistryConfigured() bool {
	if host != nil {
		return host.TenantsRegistryConfigured()
	}
	return false
}

func registerTenant(entry RegistryEntry) error {
	if host == nil {
		return errHostNotConfigured
	}
	return host.RegisterTenant(entry)
}

func newTenantEntry(tenantID, orgName, email, plan string) RegistryEntry {
	if host != nil {
		return host.NewTenantEntry(tenantID, orgName, email, plan)
	}
	return RegistryEntry{TenantID: tenantID, OrgName: orgName, Email: email, Plan: plan}
}

func applyPlanQuotas(tenantID string, limits QuotaLimits) error {
	if host == nil {
		return errHostNotConfigured
	}
	return host.ApplyPlanQuotas(tenantID, limits)
}

func updateTenantPlan(tenantID, plan string) error {
	if host == nil {
		return errHostNotConfigured
	}
	return host.UpdateTenantPlan(tenantID, plan)
}

func updateTenantStripeCustomer(tenantID, customerID string) error {
	if host == nil {
		return errHostNotConfigured
	}
	return host.UpdateTenantStripeCustomer(tenantID, customerID)
}

func tenantEmail(tenantID string) string {
	if host != nil {
		return host.TenantEmail(tenantID)
	}
	return ""
}

func provisionDataDir(tenantID string) error {
	if host == nil {
		return errHostNotConfigured
	}
	return host.ProvisionDataDir(tenantID)
}

func adminProvisioningEnabled() bool {
	if host != nil {
		return host.AdminProvisioningEnabled()
	}
	return false
}

func provisionAdminUser(tenantID string) (username, password string, err error) {
	if host == nil {
		return "", "", errHostNotConfigured
	}
	return host.ProvisionAdminUser(tenantID)
}

func defaultDomainID() string {
	if host != nil {
		return host.DefaultDomainID()
	}
	return "default"
}

type hostError string

func (e hostError) Error() string { return string(e) }

const errHostNotConfigured = hostError("saas host not configured")

// RateLimitMiddleware is optional per-route rate limiting from app.
type RateLimitMiddleware gin.HandlerFunc
