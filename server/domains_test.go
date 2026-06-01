package main

import (
	"os"
	"path/filepath"
	"testing"
)

func domainsConfigForTest(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	candidates := []string{
		filepath.Join(wd, "..", "config", "domains.json"),
		filepath.Join(wd, "config", "domains.json"),
	}
	if env := os.Getenv("DOMAINS_CONFIG_PATH"); env != "" {
		candidates = append([]string{env}, candidates...)
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	t.Skip("config/domains.json not found")
	return ""
}

func TestNormalizeDomainID(t *testing.T) {
	t.Setenv("DOMAINS_CONFIG_PATH", domainsConfigForTest(t))
	domainCatalog = domainsFile{}
	if err := loadDomainCatalog(); err != nil {
		t.Fatalf("loadDomainCatalog: %v", err)
	}

	id, err := normalizeDomainID("default")
	if err != nil || id != "default" {
		t.Fatalf("default: id=%q err=%v", id, err)
	}

	_, err = normalizeDomainID("unknown_domain_xyz")
	if err == nil {
		t.Fatal("expected error for unknown domain")
	}
}
