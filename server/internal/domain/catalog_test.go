package domain

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeID(t *testing.T) {
	t.Setenv("DOMAINS_CONFIG_PATH", ConfigPathForTest(t))
	ResetCatalog()
	if err := LoadCatalog(); err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}

	id, err := NormalizeID("default")
	if err != nil || id != "default" {
		t.Fatalf("default: id=%q err=%v", id, err)
	}

	_, err = NormalizeID("unknown_domain_xyz")
	if err == nil {
		t.Fatal("expected error for unknown domain")
	}
}

func TestConfigPathForTest(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(wd, "..", "config", "domains.json")
	if _, err := os.Stat(p); err != nil {
		t.Skip("domains.json not found")
	}
	t.Setenv("DOMAINS_CONFIG_PATH", p)
	got := ConfigPathForTest(t)
	if got != p {
		t.Fatalf("expected %q, got %q", p, got)
	}
}
