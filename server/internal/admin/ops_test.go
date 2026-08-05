package admin

import (
	"os"
	"path/filepath"
	"testing"

	cfgpkg "grounded_llm_server/internal/config"
)

func TestValidateTenantPurgeTarget(t *testing.T) {
	BindConfig(func() *cfgpkg.Config {
		return &cfgpkg.Config{DefaultTenantID: "default"}
	})

	if err := ValidateTenantPurgeTarget("acme", true, false); err != nil {
		t.Fatalf("acme: %v", err)
	}
	if err := ValidateTenantPurgeTarget("acme", false, false); err == nil {
		t.Fatal("expected confirm error")
	}
	if err := ValidateTenantPurgeTarget("default", true, false); err == nil {
		t.Fatal("expected default tenant guard")
	}
	if err := ValidateTenantPurgeTarget("default", true, true); err != nil {
		t.Fatalf("default with purge_default: %v", err)
	}
	if err := ValidateTenantPurgeTarget("../bad", true, false); err == nil {
		t.Fatal("expected invalid tenant_id")
	}
}

func TestProvisionSignupAdminUser(t *testing.T) {
	dir := t.TempDir()
	usersFile := filepath.Join(dir, "admin_users.json")
	t.Setenv("ADMIN_USERS_FILE", usersFile)

	userRegistry = make(map[string]userRecord)
	user, pass, err := ProvisionSignupAdminUser("acme-demo")
	if err != nil {
		t.Fatal(err)
	}
	if user != "acme-demo-admin" || len(pass) < 8 {
		t.Fatalf("user=%q pass len=%d", user, len(pass))
	}
	if _, ok := userRegistry[user]; !ok {
		t.Fatal("user not in registry")
	}
	if _, err := os.Stat(usersFile); err != nil {
		t.Fatal(err)
	}
}

func TestIsReindexTerminal(t *testing.T) {
	if !IsReindexTerminal(storeReindexSucceeded()) {
		t.Fatal("succeeded should be terminal")
	}
	if !IsReindexTerminal(storeReindexFailed()) {
		t.Fatal("failed should be terminal")
	}
	if IsReindexTerminal(storeReindexRunning()) {
		t.Fatal("running should not be terminal")
	}
}

func storeReindexSucceeded() string { return "succeeded" }
func storeReindexFailed() string    { return "failed" }
func storeReindexRunning() string   { return "running" }

func TestReindexStatusLabel(t *testing.T) {
	cases := map[string]string{
		"pending":   "queued",
		"running":   "running",
		"succeeded": "succeeded",
		"failed":    "failed",
	}
	for in, want := range cases {
		if got := ReindexStatusLabel(in); got != want {
			t.Fatalf("%q: got %q want %q", in, got, want)
		}
	}
}

func TestReindexAcceptedMessage(t *testing.T) {
	if ReindexAcceptedMessage(false) == "" {
		t.Fatal("expected message for new job")
	}
	if ReindexAcceptedMessage(true) == ReindexAcceptedMessage(false) {
		t.Fatal("already running message should differ")
	}
}

func TestUploadValidateTxt(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "ok.txt")
	if err := os.WriteFile(p, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ValidateKnowledgeFileContent(p, "ok.txt"); err != nil {
		t.Fatal(err)
	}
}
