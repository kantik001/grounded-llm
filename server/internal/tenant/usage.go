package tenant

import (
	"context"
	"fmt"
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

// CollectQuotaUsage gathers usage metrics for tenantID.
func CollectQuotaUsage(ctx context.Context, tenantID string) (QuotaUsage, error) {
	usage := QuotaUsage{}
	st := chatStore()
	if st == nil {
		return usage, nil
	}
	n, err := st.CountTenantUserMessagesToday(ctx, tenantID)
	if err != nil {
		return usage, err
	}
	usage.MessagesToday = n
	storage, err := st.TenantKBStorageBytes(ctx, tenantID)
	if err != nil {
		return usage, err
	}
	usage.StorageBytes = storage
	usage.StorageMB = float64(usage.StorageBytes) / (1024 * 1024)
	domains, err := st.CountTenantKBDomains(ctx, tenantID)
	if err != nil {
		return usage, err
	}
	usage.Domains = domains
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
func CheckStorageQuota(ctx context.Context, tenantID string, additionalBytes int64) error {
	limits, ok := QuotaLimitsFor(tenantID)
	if !ok || !limitActive(limits.StorageMB) {
		return nil
	}
	st := chatStore()
	if st == nil {
		return nil
	}
	maxBytes := int64(limits.StorageMB) * 1024 * 1024
	used, err := st.TenantKBStorageBytes(ctx, tenantID)
	if err != nil {
		return err
	}
	if used+additionalBytes > maxBytes {
		usedMB := float64(used) / (1024 * 1024)
		return fmt.Errorf("storage quota exceeded for tenant %s (%.1f/%d MB)", tenantID, usedMB, limits.StorageMB)
	}
	return nil
}

// CheckDomainQuota returns an error when adding domainID would exceed max domains.
func CheckDomainQuota(ctx context.Context, tenantID, domainID string) error {
	limits, ok := QuotaLimitsFor(tenantID)
	if !ok || !limitActive(limits.MaxDomains) {
		return nil
	}
	st := chatStore()
	if st == nil {
		return nil
	}
	has, err := st.TenantHasKBDomain(ctx, tenantID, domainID)
	if err != nil {
		return err
	}
	if has {
		return nil
	}
	count, err := st.CountTenantKBDomains(ctx, tenantID)
	if err != nil {
		return err
	}
	if count >= limits.MaxDomains {
		return fmt.Errorf("domain quota exceeded for tenant %s (%d/%d)", tenantID, count, limits.MaxDomains)
	}
	return nil
}
