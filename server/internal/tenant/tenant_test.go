package tenant

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"grounded_llm_server/internal/config"
)

func bindTestConfig(t *testing.T, dataDir, defaultTenant string) {
	t.Helper()
	cfg := &config.Config{DataDir: dataDir, DefaultTenantID: defaultTenant}
	BindConfig(func() *config.Config { return cfg })
	InitAllowlist(cfg)
}

func TestKbDataDirLegacyFallback(t *testing.T) {
	dir := t.TempDir()
	legacy := filepath.Join(dir, "default")
	if err := os.MkdirAll(legacy, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacy, "doc.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	bindTestConfig(t, dir, "default")
	got := KBDataDir("default", "default")
	if got != legacy {
		t.Fatalf("got %q want %q", got, legacy)
	}
}

func TestNormalizeTenantID(t *testing.T) {
	if NormalizeTenantID(" AcMe ") != "acme" {
		t.Fatal()
	}
}

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
	bindTestConfig(t, dir, "default")
	got := countTenantKBDomains(dir, "default")
	if got != 2 {
		t.Fatalf("domains=%d want 2", got)
	}
}

func TestCheckStorageQuota(t *testing.T) {
	dir := t.TempDir()
	bindTestConfig(t, dir, "default")
	SetQuotaForTest("default", QuotaLimits{StorageMB: 1})
	if err := CheckStorageQuota("default", 2*1024*1024); err == nil {
		t.Fatal("expected storage quota error")
	}
	if err := CheckStorageQuota("default", 100); err != nil {
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
	bindTestConfig(t, dir, "default")
	SetQuotaForTest("acme", QuotaLimits{MaxDomains: 1})
	if err := CheckDomainQuota("acme", "legal"); err == nil {
		t.Fatal("expected domain quota error")
	}
	if err := CheckDomainQuota("acme", "hr"); err != nil {
		t.Fatalf("existing domain should pass: %v", err)
	}
}

func TestCheckMessageQuotaUnlimited(t *testing.T) {
	ResetQuotas()
	if err := CheckMessageQuota(context.Background(), "default"); err != nil {
		t.Fatalf("no quotas: %v", err)
	}
}

func TestQuotaLimitsForTenant(t *testing.T) {
	SetQuotaForTest("default", QuotaLimits{MessagesPerDay: 100})
	lim, ok := QuotaLimitsFor("default")
	if !ok || lim.MessagesPerDay != 100 {
		t.Fatalf("got %+v ok=%v", lim, ok)
	}
}
