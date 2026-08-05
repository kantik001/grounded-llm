package analytics

import (
	"strings"
	"testing"

	"grounded_llm_server/internal/rag"
)

func TestQuestionPreview(t *testing.T) {
	if got := QuestionPreview("  hello   world  "); got != "hello world" {
		t.Fatalf("normalize: %q", got)
	}
	long := strings.Repeat("x", 100)
	got := QuestionPreview(long)
	if len(got) != 83 {
		t.Fatalf("truncate len: got %d want 83", len(got))
	}
	if got[len(got)-3:] != "…" {
		t.Fatalf("truncate suffix: %q", got)
	}
}

func TestShouldRecordRAG(t *testing.T) {
	if !ShouldRecordRAG(rag.AnswerResult{SoftFail: true}) {
		t.Fatal("soft fail should record")
	}
	if !ShouldRecordRAG(rag.AnswerResult{OK: true, VerifyPass: true}) {
		t.Fatal("verified answer should record")
	}
	if ShouldRecordRAG(rag.AnswerResult{OK: false, ErrMsg: "LLM down"}) {
		t.Fatal("generic error should not record")
	}
}
