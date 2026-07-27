package main

import (
	"strings"
	"testing"
)

func TestFormatGuardrailsViolations(t *testing.T) {
	got := formatGuardrailsViolations([]string{
		"numeric_unmatched:72",
		"pii:email:u***@example.com",
	})
	if !strings.Contains(got, "72") || !strings.Contains(got, "PII") {
		t.Fatalf("got %q", got)
	}
}

func TestNormalizeGuardrailsMode(t *testing.T) {
	if normalizeGuardrailsMode("") != GuardrailsModeLocal {
		t.Fatal("default local")
	}
	if normalizeGuardrailsMode("REMOTE") != GuardrailsModeRemote {
		t.Fatal("remote")
	}
	if normalizeGuardrailsMode("hybrid") != GuardrailsModeHybrid {
		t.Fatal("hybrid")
	}
}

func TestVerifyRAGAnswer_LocalUnaffected(t *testing.T) {
	config = &Config{GuardrailsMode: "local", DefaultLocale: "en"}
	fragments := []RAGFragment{{Filename: "Table", Content: "Average value 77 and range 3-72."}}
	answer := appendRAGDisclaimer("Average 77.", "en")
	ok, reason := verifyRAGAnswer(answer, fragments, "en")
	if !ok {
		t.Fatalf("expected pass: %s", reason)
	}
}
