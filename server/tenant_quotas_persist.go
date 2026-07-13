package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func tenantQuotasFilePath() string {
	return strings.TrimSpace(os.Getenv("TENANT_QUOTAS_FILE"))
}

func upsertTenantQuota(tenantID string, limits TenantQuotaLimits) error {
	path := tenantQuotasFilePath()
	if path == "" {
		tenantQuotaRegistry[normalizeTenantID(tenantID)] = limits
		return nil
	}

	tenantRegistryMu.Lock()
	defer tenantRegistryMu.Unlock()

	var entries []tenantQuotaFileEntry
	if body, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(body, &entries)
	}

	id := normalizeTenantID(tenantID)
	updated := false
	for i, e := range entries {
		if normalizeTenantID(e.TenantID) == id {
			entries[i].MessagesPerDay = limits.MessagesPerDay
			entries[i].StorageMB = limits.StorageMB
			entries[i].MaxDomains = limits.MaxDomains
			updated = true
			break
		}
	}
	if !updated {
		entries = append(entries, tenantQuotaFileEntry{
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

	tenantQuotaRegistry[id] = limits
	return nil
}

func applyPlanQuotas(tenantID, planID string) error {
	plan, ok := planByID(planID)
	if !ok {
		return fmt.Errorf("unknown plan: %s", planID)
	}
	return upsertTenantQuota(tenantID, planQuotasToLimits(plan.Quotas))
}
