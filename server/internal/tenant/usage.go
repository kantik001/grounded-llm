package tenant

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// QuotaUsage summarizes current tenant resource consumption.
type QuotaUsage struct {
	MessagesToday int64   `json:"messages_today"`
	StorageBytes  int64   `json:"storage_bytes"`
	StorageMB     float64 `json:"storage_mb"`
	Domains       int     `json:"domains"`
}

// QuotaStatus combines limits, usage, and enforcement flag for admin APIs.
type QuotaStatus struct {
	TenantID string      `json:"tenant_id"`
	Limits   QuotaLimits `json:"limits"`
	Usage    QuotaUsage  `json:"usage"`
	Enforced bool        `json:"enforced"`
}

func tenantStorageBytes(dataDir, tenantID string) int64 {
	tenantID = NormalizeTenantID(tenantID)
	if tenantID == "" {
		tenantID = "default"
	}
	root := filepath.Join(dataDir, tenantID)
	var total int64
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !isKnowledgeFile(d.Name()) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		total += info.Size()
		return nil
	})
	return total
}

func countTenantKBDomains(dataDir, tenantID string) int {
	tenantID = NormalizeTenantID(tenantID)
	if tenantID == "" {
		tenantID = "default"
	}
	root := filepath.Join(dataDir, tenantID)
	entries, err := os.ReadDir(root)
	if err != nil {
		return 0
	}
	domains := make(map[string]struct{})
	for _, e := range entries {
		if e.IsDir() {
			if hasKnowledgeFiles(filepath.Join(root, e.Name())) {
				domains[e.Name()] = struct{}{}
			}
			continue
		}
		if isKnowledgeFile(e.Name()) {
			domains["default"] = struct{}{}
		}
	}
	return len(domains)
}

func tenantHasKBDomains(tenantID, domainID string) bool {
	dir := KBDataDir(tenantID, domainID)
	return hasKnowledgeFiles(dir)
}

// CollectQuotaUsage gathers usage metrics for tenantID.
func CollectQuotaUsage(ctx context.Context, tenantID string) (QuotaUsage, error) {
	c := cfg()
	usage := QuotaUsage{
		Domains: countTenantKBDomains(c.DataDir, tenantID),
	}
	if st := chatStore(); st != nil {
		n, err := st.CountTenantUserMessagesToday(ctx, tenantID)
		if err != nil {
			return usage, err
		}
		usage.MessagesToday = n
	}
	usage.StorageBytes = tenantStorageBytes(c.DataDir, tenantID)
	usage.StorageMB = float64(usage.StorageBytes) / (1024 * 1024)
	return usage, nil
}

// BuildQuotaStatus returns limits, usage, and enforcement for admin quota APIs.
func BuildQuotaStatus(ctx context.Context, tenantID string) (QuotaStatus, error) {
	limits, enforced := QuotaLimitsFor(tenantID)
	usage, err := CollectQuotaUsage(ctx, tenantID)
	if err != nil {
		return QuotaStatus{}, err
	}
	return QuotaStatus{
		TenantID: NormalizeTenantID(tenantID),
		Limits:   limits,
		Usage:    usage,
		Enforced: enforced,
	}, nil
}

// CheckMessageQuota returns an error when the daily message cap is exceeded.
func CheckMessageQuota(ctx context.Context, tenantID string) error {
	limits, ok := QuotaLimitsFor(tenantID)
	if !ok || !limitActive(limits.MessagesPerDay) {
		return nil
	}
	st := chatStore()
	if st == nil {
		return nil
	}
	n, err := st.CountTenantUserMessagesToday(ctx, tenantID)
	if err != nil {
		return err
	}
	if n >= int64(limits.MessagesPerDay) {
		return fmt.Errorf("daily message quota exceeded for tenant %s (%d/%d)", tenantID, n, limits.MessagesPerDay)
	}
	return nil
}

// CheckStorageQuota returns an error when storage would exceed the tenant cap.
func CheckStorageQuota(tenantID string, additionalBytes int64) error {
	limits, ok := QuotaLimitsFor(tenantID)
	if !ok || !limitActive(limits.StorageMB) {
		return nil
	}
	maxBytes := int64(limits.StorageMB) * 1024 * 1024
	used := tenantStorageBytes(cfg().DataDir, tenantID)
	if used+additionalBytes > maxBytes {
		usedMB := float64(used) / (1024 * 1024)
		return fmt.Errorf("storage quota exceeded for tenant %s (%.1f/%d MB)", tenantID, usedMB, limits.StorageMB)
	}
	return nil
}

// CheckDomainQuota returns an error when adding domainID would exceed max domains.
func CheckDomainQuota(tenantID, domainID string) error {
	limits, ok := QuotaLimitsFor(tenantID)
	if !ok || !limitActive(limits.MaxDomains) {
		return nil
	}
	if tenantHasKBDomains(tenantID, domainID) {
		return nil
	}
	count := countTenantKBDomains(cfg().DataDir, tenantID)
	if count >= limits.MaxDomains {
		return fmt.Errorf("domain quota exceeded for tenant %s (%d/%d)", tenantID, count, limits.MaxDomains)
	}
	return nil
}
