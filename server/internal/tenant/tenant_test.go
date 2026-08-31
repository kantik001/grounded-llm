package tenant

import (
	"context"
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

func TestKbDataDirTenantNested(t *testing.T) {
	dir := t.TempDir()
	bindTestConfig(t, dir, "default")
	got := KBDataDir("default", "hr")
	want := filepath.Join(dir, "default", "hr")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestNormalizeTenantID(t *testing.T) {
	if NormalizeTenantID(" AcMe ") != "acme" {
		t.Fatal()
	}
}

func TestCheckStorageQuotaWithoutStore(t *testing.T) {
	ResetQuotas()
	BindStore(nil)
	bindTestConfig(t, t.TempDir(), "default")
	SetQuotaForTest("default", QuotaLimits{StorageMB: 1})
	if err := CheckStorageQuota(context.Background(), "default", 2*1024*1024); err != nil {
		t.Fatalf("without store quotas are not enforced: %v", err)
	}
}

func TestCheckDomainQuotaWithoutStore(t *testing.T) {
	ResetQuotas()
	BindStore(nil)
	bindTestConfig(t, t.TempDir(), "default")
	SetQuotaForTest("acme", QuotaLimits{MaxDomains: 1})
	if err := CheckDomainQuota(context.Background(), "acme", "legal"); err != nil {
		t.Fatalf("without store quotas are not enforced: %v", err)
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
