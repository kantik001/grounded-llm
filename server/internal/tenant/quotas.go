package tenant

import (
	"encoding/json"
	"log"
	"os"
	"strings"
)

// QuotaLimits holds optional caps per tenant (0 = unlimited).
type QuotaLimits struct {
	MessagesPerDay int `json:"messages_per_day"`
	StorageMB      int `json:"storage_mb"`
	MaxDomains     int `json:"max_domains"`
}

type quotaFileEntry struct {
	TenantID       string `json:"tenant_id"`
	MessagesPerDay int    `json:"messages_per_day"`
	StorageMB      int    `json:"storage_mb"`
	MaxDomains     int    `json:"max_domains"`
}

var quotaRegistry map[string]QuotaLimits

// LoadQuotas reads TENANT_QUOTAS_FILE into memory.
func LoadQuotas() {
	quotaRegistry = make(map[string]QuotaLimits)
	path := strings.TrimSpace(os.Getenv("TENANT_QUOTAS_FILE"))
	if path == "" {
		return
	}
	body, err := os.ReadFile(path)
	if err != nil {
		log.Printf("TENANT_QUOTAS_FILE read error: %v", err)
		return
	}
	var entries []quotaFileEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		log.Printf("TENANT_QUOTAS_FILE parse error: %v", err)
		return
	}
	for _, e := range entries {
		id := NormalizeTenantID(e.TenantID)
		if id == "" {
			continue
		}
		quotaRegistry[id] = QuotaLimits{
			MessagesPerDay: e.MessagesPerDay,
			StorageMB:      e.StorageMB,
			MaxDomains:     e.MaxDomains,
		}
	}
	log.Printf("Tenant quotas: %d tenant(s) configured", len(quotaRegistry))
}

// QuotaLimitsFor returns configured limits for tenantID.
func QuotaLimitsFor(tenantID string) (QuotaLimits, bool) {
	lim, ok := quotaRegistry[NormalizeTenantID(tenantID)]
	return lim, ok
}

// ConfiguredQuotaCount returns the number of tenants with quota limits loaded.
func ConfiguredQuotaCount() int {
	return len(quotaRegistry)
}

// ResetQuotas clears in-memory quota limits (tests).
func ResetQuotas() {
	quotaRegistry = nil
}

// SetQuotaForTest sets in-memory quota limits without persisting (tests).
func SetQuotaForTest(tenantID string, limits QuotaLimits) {
	if quotaRegistry == nil {
		quotaRegistry = make(map[string]QuotaLimits)
	}
	quotaRegistry[NormalizeTenantID(tenantID)] = limits
}

func limitActive(v int) bool {
	return v > 0
}
