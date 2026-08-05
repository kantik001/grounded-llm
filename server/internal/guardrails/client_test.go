package guardrails

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

func TestNormalizeMode(t *testing.T) {
	if NormalizeMode("") != ModeLocal {
		t.Fatal("default local")
	}
	if NormalizeMode("REMOTE") != ModeRemote {
		t.Fatal("remote")
	}
	if NormalizeMode("hybrid") != ModeHybrid {
		t.Fatal("hybrid")
	}
}
