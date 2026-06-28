package main

import (
	"strings"
	"testing"
)

func TestParseAnalyticsDays(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"", 7},
		{"7", 7},
		{"30", 30},
		{"0", 7},
		{"-1", 7},
		{"abc", 7},
		{"120", 90},
	}
	for _, tc := range cases {
		if got := parseAnalyticsDays(tc.in); got != tc.want {
			t.Fatalf("parseAnalyticsDays(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestComputeVerifyPassRate(t *testing.T) {
	if got := computeVerifyPassRate(0, 0); got != 0 {
		t.Fatalf("empty: got %v", got)
	}
	if got := computeVerifyPassRate(8, 2); got != 80 {
		t.Fatalf("80%%: got %v", got)
	}
}

func TestQuestionPreviewForAnalytics(t *testing.T) {
	if got := questionPreviewForAnalytics("  hello   world  "); got != "hello world" {
		t.Fatalf("normalize: %q", got)
	}
	long := strings.Repeat("x", 100)
	got := questionPreviewForAnalytics(long)
	if len(got) != 83 {
		t.Fatalf("truncate len: got %d want 83", len(got))
	}
	if got[len(got)-3:] != "…" {
		t.Fatalf("truncate suffix: %q", got)
	}
}

func TestShouldRecordRAGAnalytics(t *testing.T) {
	if !shouldRecordRAGAnalytics(RAGAnswerResult{SoftFail: true}) {
		t.Fatal("soft fail should record")
	}
	if !shouldRecordRAGAnalytics(RAGAnswerResult{OK: true, VerifyPass: true}) {
		t.Fatal("verified answer should record")
	}
	if shouldRecordRAGAnalytics(RAGAnswerResult{OK: false, ErrMsg: "LLM down"}) {
		t.Fatal("generic error should not record")
	}
}
