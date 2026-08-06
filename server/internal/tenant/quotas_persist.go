package tenant

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"grounded_llm_server/internal/store"
)

func quotasFilePath() string {
	return strings.TrimSpace(os.Getenv("TENANT_QUOTAS_FILE"))
}

// UpsertQuota writes or updates quota limits for a tenant.
// Prefer Postgres when available; also dual-writes JSON when TENANT_QUOTAS_FILE is set.
func UpsertQuota(tenantID string, limits QuotaLimits) error {
	id := NormalizeTenantID(tenantID)
	path := quotasFilePath()
	pg := UsePostgresBackend()

	if pg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := saasStore().UpsertTenantQuota(ctx, id, store.TenantQuotaLimits{
			MessagesPerDay: limits.MessagesPerDay,
			StorageMB:      limits.StorageMB,
			MaxDomains:     limits.MaxDomains,
		}); err != nil {
			return err
		}
	}

	if path == "" {
		if quotaRegistry == nil {
			quotaRegistry = make(map[string]QuotaLimits)
		}
		quotaRegistry[id] = limits
		return nil
	}

	registryMu.Lock()
	defer registryMu.Unlock()

	var entries []quotaFileEntry
	if body, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(body, &entries)
	}

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

// LoadQuotas hydrates in-memory quotas from JSON and/or Postgres.
func LoadQuotasFromStore() {
	if !UsePostgresBackend() {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	rows, err := saasStore().ListTenantQuotas(ctx)
	if err != nil {
		log.Printf("tenant_quotas load error: %v", err)
		return
	}
	if quotaRegistry == nil {
		quotaRegistry = make(map[string]QuotaLimits)
	}
	for id, lim := range rows {
		quotaRegistry[NormalizeTenantID(id)] = QuotaLimits{
			MessagesPerDay: lim.MessagesPerDay,
			StorageMB:      lim.StorageMB,
			MaxDomains:     lim.MaxDomains,
		}
	}
	if len(rows) > 0 {
		log.Printf("Tenant quotas: merged %d row(s) from Postgres", len(rows))
	}
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
