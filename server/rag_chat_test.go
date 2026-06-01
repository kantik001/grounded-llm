package main

import (
	"strings"
	"testing"
)

func TestExtractNumbersFromText(t *testing.T) {
	nums := extractNumbersFromText("Прирост 748,5 см и 31.8%")
	if len(nums) != 2 {
		t.Fatalf("expected 2 numbers, got %v", nums)
	}
	if nums[0] != 748.5 || nums[1] != 31.8 {
		t.Fatalf("unexpected values: %v", nums)
	}
}

func TestVerifyRAGAnswer_NoNumbersOK(t *testing.T) {
	fragments := []RAGFragment{{Filename: "Статья", Content: "Парша проявляется пятнами."}}
	answer := appendRAGDisclaimer("Парша проявляется пятнами на листьях.")
	ok, reason := verifyRAGAnswer(answer, fragments)
	if !ok {
		t.Fatalf("expected pass, got: %s", reason)
	}
}

func TestVerifyRAGAnswer_NumberInContextOK(t *testing.T) {
	fragments := []RAGFragment{{Filename: "Таблица", Content: "Среднее значение 77 и повторность 3-72."}}
	answer := appendRAGDisclaimer("Среднее 77.")
	ok, reason := verifyRAGAnswer(answer, fragments)
	if !ok {
		t.Fatalf("expected pass, got: %s", reason)
	}
}

func TestVerifyRAGAnswer_HallucinatedNumberFails(t *testing.T) {
	fragments := []RAGFragment{{Filename: "Статья", Content: "Без цифр в тексте."}}
	answer := appendRAGDisclaimer("Рентабельность 72%.")
	ok, reason := verifyRAGAnswer(answer, fragments)
	if ok {
		t.Fatal("expected verification to fail for hallucinated number")
	}
	if reason == "" {
		t.Fatal("expected non-empty reason")
	}
}

func TestAppendRAGDisclaimer_StripsSourceAndAddsDisclaimer(t *testing.T) {
	raw := "Ответ по теме.\n\nИсточник: \"Секретная статья\""
	out := appendRAGDisclaimer(raw)
	if strings.Contains(out, "Источник:") || strings.Contains(out, "Секретная статья") {
		t.Fatalf("source attribution should be removed: %q", out)
	}
	if !strings.Contains(out, "Не заменяет официальную консультацию") {
		t.Fatalf("expected disclaimer, got: %q", out)
	}
}

func TestCleanRAGAnswer_StripsIntroPhrase(t *testing.T) {
	out := cleanRAGAnswer("Я думаю, что парша опасна для урожая.")
	if strings.Contains(out, "Я думаю") {
		t.Fatalf("intro should be stripped, got: %q", out)
	}
	if !strings.Contains(out, "парша") {
		t.Fatalf("got %q", out)
	}
}
