package domain

import (
	"os"
	"path/filepath"
	"testing"
)

// ConfigPathForTest resolves domains.json for integration tests.
func ConfigPathForTest(t *testing.T) string {
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
