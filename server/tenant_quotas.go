package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"
)

// TenantQuotaLimits — optional caps per tenant (0 = unlimited).
type TenantQuotaLimits struct {
	MessagesPerDay int `json:"messages_per_day"`
	StorageMB      int `json:"storage_mb"`
	MaxDomains     int `json:"max_domains"`
}

type tenantQuotaFileEntry struct {
	TenantID         string `json:"tenant_id"`
	MessagesPerDay   int    `json:"messages_per_day"`
	StorageMB        int    `json:"storage_mb"`
	MaxDomains       int    `json:"max_domains"`
}

var tenantQuotaRegistry map[string]TenantQuotaLimits

func loadTenantQuotas() {
	tenantQuotaRegistry = make(map[string]TenantQuotaLimits)
	path := strings.TrimSpace(os.Getenv("TENANT_QUOTAS_FILE"))
	if path == "" {
		return
	}
	body, err := os.ReadFile(path)
	if err != nil {
		log.Printf("TENANT_QUOTAS_FILE read error: %v", err)
		return
	}
	var entries []tenantQuotaFileEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		log.Printf("TENANT_QUOTAS_FILE parse error: %v", err)
		return
	}
	for _, e := range entries {
		id := normalizeTenantID(e.TenantID)
		if id == "" {
			continue
		}
		tenantQuotaRegistry[id] = TenantQuotaLimits{
			MessagesPerDay: e.MessagesPerDay,
			StorageMB:      e.StorageMB,
			MaxDomains:     e.MaxDomains,
		}
	}
	log.Printf("Tenant quotas: %d tenant(s) configured", len(tenantQuotaRegistry))
}

func quotaLimitsForTenant(tenantID string) (TenantQuotaLimits, bool) {
	lim, ok := tenantQuotaRegistry[normalizeTenantID(tenantID)]
	return lim, ok
}

func limitActive(v int) bool {
	return v > 0
}
