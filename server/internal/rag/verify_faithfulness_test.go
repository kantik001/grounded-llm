package rag

import (
	"strings"
	"testing"
)

const faithCtx = `Возврат товара возможен в течение 14 дней с момента покупки.
Для возврата необходим чек и заявление. Обмен производится в магазине покупки.
Refunds are processed within 14 days of purchase with a valid receipt.`

func TestUnsupportedSentencesPassesGroundedAnswer(t *testing.T) {
	body := "Возврат товара возможен в течение 14 дней. Для возврата необходим чек и заявление."
	if got := UnsupportedSentences(body, faithCtx, 0.5); len(got) != 0 {
		t.Fatalf("grounded answer flagged: %v", got)
	}
}

func TestUnsupportedSentencesCatchesFabrication(t *testing.T) {
	body := "Возврат товара возможен в течение 14 дней. " +
		"Компенсация выплачивается биткоинами через мобильное приложение банка."
	got := UnsupportedSentences(body, faithCtx, 0.5)
	if len(got) != 1 {
		t.Fatalf("want 1 unsupported sentence, got %v", got)
	}
	if !strings.Contains(got[0], "биткоинами") {
		t.Fatalf("wrong sentence flagged: %q", got[0])
	}
}

func TestUnsupportedSentencesToleratesInflection(t *testing.T) {
	// "возврата"/"возврат", "чеком"/"чек" — stem matching must count these.
	body := "Для возврата товара потребуется заявление вместе с чеком."
	if got := UnsupportedSentences(body, faithCtx, 0.5); len(got) != 0 {
		t.Fatalf("inflected grounded answer flagged: %v", got)
	}
}

func TestUnsupportedSentencesSkipsShortSentences(t *testing.T) {
	if got := UnsupportedSentences("Да, конечно.", faithCtx, 0.5); len(got) != 0 {
		t.Fatalf("short sentence must be skipped: %v", got)
	}
}

func TestVerifyFaithfulnessModes(t *testing.T) {
	fabricated := "Компенсация выплачивается биткоинами через мобильное приложение согласно новому регламенту."

	t.Setenv("VERIFY_FAITHFULNESS", "enforce")
	ok, reason := VerifyFaithfulnessLocal(fabricated, faithCtx)
	if ok {
		t.Fatal("enforce mode must fail on fabricated claim")
	}
	if reason == "" {
		t.Fatal("expected reason")
	}

	t.Setenv("VERIFY_FAITHFULNESS", "warn")
	ok, reason = VerifyFaithfulnessLocal(fabricated, faithCtx)
	if !ok || reason == "" {
		t.Fatalf("warn mode must pass with reason: ok=%v reason=%q", ok, reason)
	}

	t.Setenv("VERIFY_FAITHFULNESS", "off")
	ok, reason = VerifyFaithfulnessLocal(fabricated, faithCtx)
	if !ok || reason != "" {
		t.Fatalf("off mode must pass silently: ok=%v reason=%q", ok, reason)
	}
}

func TestVerifyAnswerLocalEnforceIntegration(t *testing.T) {
	t.Setenv("VERIFY_FAITHFULNESS", "enforce")
	// Numbers grounded, but the claim itself is fabricated.
	body := "Компенсация 14 процентов выплачивается биткоинами через приложение партнёрского банка."
	ok, reason := VerifyAnswerLocal(body, faithCtx)
	if ok {
		t.Fatalf("expected failure, got pass (%s)", reason)
	}
	if !strings.Contains(reason, "Unsupported claim") {
		t.Fatalf("reason=%q", reason)
	}
}
