package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestCountTenantKBDomains(t *testing.T) {
	dir := t.TempDir()
	tenant := filepath.Join(dir, "default")
	if err := os.MkdirAll(filepath.Join(tenant, "it_support"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tenant, "it_support", "a.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tenant, "policy_en.txt"), []byte("y"), 0o644); err != nil {
		t.Fatal(err)
	}
	config = &Config{DataDir: dir, DefaultTenantID: "default"}
	got := countTenantKBDomains(dir, "default")
	if got != 2 {
		t.Fatalf("domains=%d want 2", got)
	}
}

func TestCheckStorageQuota(t *testing.T) {
	dir := t.TempDir()
	config = &Config{DataDir: dir, DefaultTenantID: "default"}
	tenantQuotaRegistry = map[string]TenantQuotaLimits{
		"default": {StorageMB: 1},
	}
	if err := checkStorageQuota("default", 2*1024*1024); err == nil {
		t.Fatal("expected storage quota error")
	}
	if err := checkStorageQuota("default", 100); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestCheckDomainQuotaNewDomain(t *testing.T) {
	dir := t.TempDir()
	tenant := filepath.Join(dir, "acme")
	if err := os.MkdirAll(filepath.Join(tenant, "hr"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tenant, "hr", "a.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	config = &Config{DataDir: dir, DefaultTenantID: "default"}
	tenantQuotaRegistry = map[string]TenantQuotaLimits{
		"acme": {MaxDomains: 1},
	}
	if err := checkDomainQuota("acme", "legal"); err == nil {
		t.Fatal("expected domain quota error")
	}
	if err := checkDomainQuota("acme", "hr"); err != nil {
		t.Fatalf("existing domain should pass: %v", err)
	}
}

func TestCheckMessageQuotaUnlimited(t *testing.T) {
	tenantQuotaRegistry = nil
	if err := checkMessageQuota(context.Background(), "default"); err != nil {
		t.Fatalf("no quotas: %v", err)
	}
}

func TestQuotaLimitsForTenant(t *testing.T) {
	tenantQuotaRegistry = map[string]TenantQuotaLimits{
		"default": {MessagesPerDay: 100},
	}
	lim, ok := quotaLimitsForTenant("default")
	if !ok || lim.MessagesPerDay != 100 {
		t.Fatalf("got %+v ok=%v", lim, ok)
	}
}
