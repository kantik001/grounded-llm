package rag_test

import (
	"strings"
	"testing"

	"grounded_llm_server/internal/config"
	"grounded_llm_server/internal/rag"
	"grounded_llm_server/internal/store"
)

type stubHost struct{}

func (stubHost) NormalizeDomainID(raw string) (string, error) { return raw, nil }
func (stubHost) RequireRAGEnabled(string) error               { return nil }
func (stubHost) PublicAPIError(err error) string {
	if err == nil {
		return "Server error"
	}
	return err.Error()
}
func (stubHost) Prompts(string, string) (string, string) {
	return "system", "task"
}
func (stubHost) RAGConstraints(string) string { return "- no invent" }
func (stubHost) VerifyFailHint(string) string { return "contact admin" }
func (stubHost) Disclaimer(locale string) string {
	if locale == "ru" {
		return "Не заменяет официальную консультацию."
	}
	return "Reference information from the knowledge base. Not a substitute for official expert advice."
}

func TestMain(m *testing.M) {
	rag.BindConfig(func() *config.Config {
		return &config.Config{GuardrailsMode: "local", DefaultLocale: "en"}
	})
	rag.BindHost(stubHost{})
	m.Run()
}

func TestExtractNumbersFromText(t *testing.T) {
	nums := rag.ExtractNumbersFromText("Growth 748.5 cm and 31.8%")
	if len(nums) != 2 {
		t.Fatalf("expected 2 numbers, got %v", nums)
	}
	if nums[0] != 748.5 || nums[1] != 31.8 {
		t.Fatalf("unexpected values: %v", nums)
	}
}

func TestVerifyAnswer_NoNumbersOK(t *testing.T) {
	fragments := []store.RAGFragment{{Filename: "Article", Content: "Spots on leaves."}}
	answer := rag.AppendDisclaimer("Spots appear on the leaves.", "en")
	ok, reason := rag.VerifyAnswer(answer, fragments, "en")
	if !ok {
		t.Fatalf("expected pass, got: %s", reason)
	}
}

func TestVerifyAnswer_NumberInContextOK(t *testing.T) {
	fragments := []store.RAGFragment{{Filename: "Table", Content: "Average value 77 and range 3-72."}}
	answer := rag.AppendDisclaimer("Average 77.", "en")
	ok, reason := rag.VerifyAnswer(answer, fragments, "en")
	if !ok {
		t.Fatalf("expected pass, got: %s", reason)
	}
}

func TestVerifyAnswer_HallucinatedNumberFails(t *testing.T) {
	fragments := []store.RAGFragment{{Filename: "Article", Content: "No digits in text."}}
	answer := rag.AppendDisclaimer("Margin 72%.", "en")
	ok, reason := rag.VerifyAnswer(answer, fragments, "en")
	if ok {
		t.Fatal("expected verification to fail for hallucinated number")
	}
	if reason == "" {
		t.Fatal("expected non-empty reason")
	}
}

func TestAppendDisclaimer_StripsSourceAndAddsDisclaimer(t *testing.T) {
	raw := "Answer body.\n\nSource: \"Secret article\""
	out := rag.AppendDisclaimer(raw, "en")
	if strings.Contains(out, "Source:") || strings.Contains(out, "Secret article") {
		t.Fatalf("source attribution should be removed: %q", out)
	}
	if !strings.Contains(out, "Not a substitute for official expert advice") {
		t.Fatalf("expected disclaimer, got: %q", out)
	}
}

func TestAppendDisclaimer_RussianLocale(t *testing.T) {
	out := rag.AppendDisclaimer("Ответ.", "ru")
	if !strings.Contains(out, "Не заменяет официальную консультацию") {
		t.Fatalf("expected RU disclaimer, got: %q", out)
	}
}

func TestCleanAnswer_StripsIntroPhrase(t *testing.T) {
	out := rag.CleanAnswer("I think the crop is at risk.")
	if strings.Contains(out, "I think") {
		t.Fatalf("intro should be stripped, got: %q", out)
	}
	if !strings.Contains(out, "crop") {
		t.Fatalf("got %q", out)
	}
}

func TestVerifyAnswer_LocalUnaffected(t *testing.T) {
	fragments := []store.RAGFragment{{Filename: "Table", Content: "Average value 77 and range 3-72."}}
	answer := rag.AppendDisclaimer("Average 77.", "en")
	ok, reason := rag.VerifyAnswer(answer, fragments, "en")
	if !ok {
		t.Fatalf("expected pass: %s", reason)
	}
}
