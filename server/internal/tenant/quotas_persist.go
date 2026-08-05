package tenant

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func quotasFilePath() string {
	return strings.TrimSpace(os.Getenv("TENANT_QUOTAS_FILE"))
}

// UpsertQuota writes or updates quota limits for a tenant.
func UpsertQuota(tenantID string, limits QuotaLimits) error {
	path := quotasFilePath()
	if path == "" {
		if quotaRegistry == nil {
			quotaRegistry = make(map[string]QuotaLimits)
		}
		quotaRegistry[NormalizeTenantID(tenantID)] = limits
		return nil
	}

	registryMu.Lock()
	defer registryMu.Unlock()

	var entries []quotaFileEntry
	if body, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(body, &entries)
	}

	id := NormalizeTenantID(tenantID)
	updated := false
	for i, e := range entries {
		if NormalizeTenantID(e.TenantID) == id {
			entries[i].MessagesPerDay = limits.MessagesPerDay
			entries[i].StorageMB = limits.StorageMB
			entries[i].MaxDomains = limits.MaxDomains
			updated = true
			break
		}
	}
	if !updated {
		entries = append(entries, quotaFileEntry{
			TenantID:       id,
			MessagesPerDay: limits.MessagesPerDay,
			StorageMB:      limits.StorageMB,
			MaxDomains:     limits.MaxDomains,
		})
	}

	body, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, body, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		return err
	}

	if quotaRegistry == nil {
		quotaRegistry = make(map[string]QuotaLimits)
	}
	quotaRegistry[id] = limits
	return nil
}

// ApplyPlanQuotasFunc applies plan-derived limits; wired from app (plans stay in app).
type ApplyPlanQuotasFunc func(tenantID, planID string) error

var applyPlanQuotas ApplyPlanQuotasFunc

// BindApplyPlanQuotas wires plan quota application from the app layer.
func BindApplyPlanQuotas(fn ApplyPlanQuotasFunc) {
	applyPlanQuotas = fn
}

// ApplyPlanQuotas updates tenant quotas from a billing plan id.
func ApplyPlanQuotas(tenantID, planID string) error {
	if applyPlanQuotas == nil {
		return fmt.Errorf("plan quotas not configured")
	}
	return applyPlanQuotas(tenantID, planID)
}
