package main

import (
	"strings"
	"testing"
)

func TestExtractNumbersFromText(t *testing.T) {
	nums := extractNumbersFromText("Growth 748.5 cm and 31.8%")
	if len(nums) != 2 {
		t.Fatalf("expected 2 numbers, got %v", nums)
	}
	if nums[0] != 748.5 || nums[1] != 31.8 {
		t.Fatalf("unexpected values: %v", nums)
	}
}

func TestVerifyRAGAnswer_NoNumbersOK(t *testing.T) {
	fragments := []RAGFragment{{Filename: "Article", Content: "Spots on leaves."}}
	answer := appendRAGDisclaimer("Spots appear on the leaves.", "en")
	ok, reason := verifyRAGAnswer(answer, fragments, "en")
	if !ok {
		t.Fatalf("expected pass, got: %s", reason)
	}
}

func TestVerifyRAGAnswer_NumberInContextOK(t *testing.T) {
	fragments := []RAGFragment{{Filename: "Table", Content: "Average value 77 and range 3-72."}}
	answer := appendRAGDisclaimer("Average 77.", "en")
	ok, reason := verifyRAGAnswer(answer, fragments, "en")
	if !ok {
		t.Fatalf("expected pass, got: %s", reason)
	}
}

func TestVerifyRAGAnswer_HallucinatedNumberFails(t *testing.T) {
	fragments := []RAGFragment{{Filename: "Article", Content: "No digits in text."}}
	answer := appendRAGDisclaimer("Margin 72%.", "en")
	ok, reason := verifyRAGAnswer(answer, fragments, "en")
	if ok {
		t.Fatal("expected verification to fail for hallucinated number")
	}
	if reason == "" {
		t.Fatal("expected non-empty reason")
	}
}

func TestAppendRAGDisclaimer_StripsSourceAndAddsDisclaimer(t *testing.T) {
	raw := "Answer body.\n\nSource: \"Secret article\""
	out := appendRAGDisclaimer(raw, "en")
	if strings.Contains(out, "Source:") || strings.Contains(out, "Secret article") {
		t.Fatalf("source attribution should be removed: %q", out)
	}
	if !strings.Contains(out, "Not a substitute for official expert advice") {
		t.Fatalf("expected disclaimer, got: %q", out)
	}
}

func TestAppendRAGDisclaimer_RussianLocale(t *testing.T) {
	initLocaleConfig(&Config{DefaultLocale: "ru"})
	out := appendRAGDisclaimer("Ответ.", "ru")
	if !strings.Contains(out, "Не заменяет официальную консультацию") {
		t.Fatalf("expected RU disclaimer from locale bundle, got: %q", out)
	}
}

func TestCleanRAGAnswer_StripsIntroPhrase(t *testing.T) {
	out := cleanRAGAnswer("I think the crop is at risk.")
	if strings.Contains(out, "I think") {
		t.Fatalf("intro should be stripped, got: %q", out)
	}
	if !strings.Contains(out, "crop") {
		t.Fatalf("got %q", out)
	}
}
