package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateKnowledgeFileContent_TXT(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "ok.txt")
	if err := os.WriteFile(p, []byte("hello policy"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := validateKnowledgeFileContent(p, "ok.txt"); err != nil {
		t.Fatalf("txt ok: %v", err)
	}
}

func TestValidateKnowledgeFileContent_PDFMagic(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "doc.pdf")
	if err := os.WriteFile(p, []byte("%PDF-1.4 fake"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := validateKnowledgeFileContent(p, "doc.pdf"); err != nil {
		t.Fatalf("pdf ok: %v", err)
	}
	bad := filepath.Join(dir, "bad.pdf")
	if err := os.WriteFile(bad, []byte("not-a-pdf"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := validateKnowledgeFileContent(bad, "bad.pdf"); err == nil {
		t.Fatal("expected PDF magic failure")
	}
}

func TestValidateKnowledgeFileContent_RejectBinaryTXT(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bin.txt")
	if err := os.WriteFile(p, []byte{'a', 0, 'b'}, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := validateKnowledgeFileContent(p, "bin.txt"); err == nil {
		t.Fatal("expected null-byte rejection")
	}
}
