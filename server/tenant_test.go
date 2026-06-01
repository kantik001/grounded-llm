package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKbDataDirLegacyFallback(t *testing.T) {
	dir := t.TempDir()
	legacy := filepath.Join(dir, "default")
	if err := os.MkdirAll(legacy, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacy, "doc.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	config = &Config{DataDir: dir, DefaultTenantID: "default"}
	got := kbDataDir("default", "default")
	if got != legacy {
		t.Fatalf("got %q want %q", got, legacy)
	}
}

func TestNormalizeTenantID(t *testing.T) {
	if normalizeTenantID(" AcMe ") != "acme" {
		t.Fatal()
	}
}
