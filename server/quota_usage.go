package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type TenantQuotaUsage struct {
	MessagesToday int64   `json:"messages_today"`
	StorageBytes  int64   `json:"storage_bytes"`
	StorageMB     float64 `json:"storage_mb"`
	Domains       int     `json:"domains"`
}

type TenantQuotaStatus struct {
	TenantID string            `json:"tenant_id"`
	Limits   TenantQuotaLimits `json:"limits"`
	Usage    TenantQuotaUsage  `json:"usage"`
	Enforced bool              `json:"enforced"`
}

// CountTenantUserMessagesToday counts user-role messages since UTC midnight.
func (st *ChatStore) CountTenantUserMessagesToday(ctx context.Context, tenantID string) (int64, error) {
	if st == nil || st.pool == nil {
		return 0, fmt.Errorf("chat store not initialized")
	}
	var n int64
	err := st.pool.QueryRow(ctx, `
		SELECT COUNT(*)::bigint FROM messages m
		JOIN chat_sessions cs ON cs.id = m.session_id
		WHERE cs.tenant_id = $1 AND m.role = 'user'
		  AND m.created_at >= date_trunc('day', NOW() AT TIME ZONE 'UTC')`,
		normalizeTenantID(tenantID),
	).Scan(&n)
	return n, err
}

func tenantStorageBytes(dataDir, tenantID string) int64 {
	tenantID = normalizeTenantID(tenantID)
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
	tenantID = normalizeTenantID(tenantID)
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

func tenantHasKBDomains(dataDir, tenantID, domainID string) bool {
	dir := kbDataDir(tenantID, domainID)
	return hasKnowledgeFiles(dir)
}

func collectTenantQuotaUsage(ctx context.Context, tenantID string) (TenantQuotaUsage, error) {
	usage := TenantQuotaUsage{
		Domains: countTenantKBDomains(config.DataDir, tenantID),
	}
	if chatStore != nil {
		n, err := chatStore.CountTenantUserMessagesToday(ctx, tenantID)
		if err != nil {
			return usage, err
		}
		usage.MessagesToday = n
	}
	usage.StorageBytes = tenantStorageBytes(config.DataDir, tenantID)
	usage.StorageMB = float64(usage.StorageBytes) / (1024 * 1024)
	return usage, nil
}

func buildTenantQuotaStatus(ctx context.Context, tenantID string) (TenantQuotaStatus, error) {
	limits, enforced := quotaLimitsForTenant(tenantID)
	usage, err := collectTenantQuotaUsage(ctx, tenantID)
	if err != nil {
		return TenantQuotaStatus{}, err
	}
	return TenantQuotaStatus{
		TenantID: normalizeTenantID(tenantID),
		Limits:   limits,
		Usage:    usage,
		Enforced: enforced,
	}, nil
}

func checkMessageQuota(ctx context.Context, tenantID string) error {
	limits, ok := quotaLimitsForTenant(tenantID)
	if !ok || !limitActive(limits.MessagesPerDay) {
		return nil
	}
	if chatStore == nil {
		return nil
	}
	n, err := chatStore.CountTenantUserMessagesToday(ctx, tenantID)
	if err != nil {
		return err
	}
	if n >= int64(limits.MessagesPerDay) {
		return fmt.Errorf("daily message quota exceeded for tenant %s (%d/%d)", tenantID, n, limits.MessagesPerDay)
	}
	return nil
}

func checkStorageQuota(tenantID string, additionalBytes int64) error {
	limits, ok := quotaLimitsForTenant(tenantID)
	if !ok || !limitActive(limits.StorageMB) {
		return nil
	}
	maxBytes := int64(limits.StorageMB) * 1024 * 1024
	used := tenantStorageBytes(config.DataDir, tenantID)
	if used+additionalBytes > maxBytes {
		usedMB := float64(used) / (1024 * 1024)
		return fmt.Errorf("storage quota exceeded for tenant %s (%.1f/%d MB)", tenantID, usedMB, limits.StorageMB)
	}
	return nil
}

func checkDomainQuota(tenantID, domainID string) error {
	limits, ok := quotaLimitsForTenant(tenantID)
	if !ok || !limitActive(limits.MaxDomains) {
		return nil
	}
	if tenantHasKBDomains(config.DataDir, tenantID, domainID) {
		return nil
	}
	count := countTenantKBDomains(config.DataDir, tenantID)
	if count >= limits.MaxDomains {
		return fmt.Errorf("domain quota exceeded for tenant %s (%d/%d)", tenantID, count, limits.MaxDomains)
	}
	return nil
}
