package main

import (
	"fmt"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var (
	reNumberWord = regexp.MustCompile(`\b\d+(?:\.\d+)?\b`)
	reMultiSpace = regexp.MustCompile(`\s+`)
	reThink      = regexp.MustCompile(`(?i)</?think>`)
	reAnswerTag  = regexp.MustCompile(`(?i)</?answer>`)
	reSystemTag  = regexp.MustCompile(`(?i)</?system>`)
	reAbot       = regexp.MustCompile(`(?i)\babot\b`)
	reIntroEN    = regexp.MustCompile(`(?i)^(Okay|Alright|So|I think|I need to answer|From the context|Now I understand|From the table)[,:.]?\s*`)
	reIntroRU    = regexp.MustCompile(`(?i)^(Хорошо|Давайте посмотрим|Итак|Я думаю|мне нужно ответить|Из контекста видно|Теперь я понимаю|Из таблицы видно)[,:.]?\s*`)
	reSourceLine = regexp.MustCompile(`(?im)^\s*(Источник|Source):.*\n?`)
)

func disclaimerForLocale(locale string) string {
	d := strings.TrimSpace(brandingForLocale(locale).Disclaimer)
	if d != "" {
		return d
	}
	return "Reference information from the knowledge base. Not a substitute for official expert advice."
}

func extractNumbersFromText(s string) []float64 {
	s = strings.ReplaceAll(s, ",", ".")
	var out []float64
	for _, m := range reNumberWord.FindAllString(s, -1) {
		v, err := strconv.ParseFloat(m, 64)
		if err == nil {
			out = append(out, v)
		}
	}
	return out
}

func cleanRAGAnswer(text string) string {
	if text == "" {
		return "The answer could not be formatted correctly."
	}
	text = reThink.ReplaceAllString(text, "")
	text = reAnswerTag.ReplaceAllString(text, "")
	text = reSystemTag.ReplaceAllString(text, "")
	text = reAbot.ReplaceAllString(text, "")
	text = reIntroEN.ReplaceAllString(text, "")
	text = reIntroRU.ReplaceAllString(text, "")
	text = strings.TrimSpace(reMultiSpace.ReplaceAllString(text, " "))
	if text == "" {
		return "The answer could not be formatted correctly."
	}
	return text
}

func stripSourceAttribution(answer string) string {
	s := reSourceLine.ReplaceAllString(answer, "")
	return strings.TrimSpace(reMultiSpace.ReplaceAllString(s, " "))
}

func appendRAGDisclaimer(answer, locale string) string {
	disclaimer := disclaimerForLocale(locale)
	body := stripSourceAttribution(answer)
	if body == "" {
		return disclaimer
	}
	if strings.Contains(body, disclaimer) {
		return body
	}
	return body + "\n\n" + disclaimer
}

func answerBodyForVerification(answer, locale string) string {
	s := stripSourceAttribution(answer)
	s = strings.ReplaceAll(s, disclaimerForLocale(locale), "")
	return strings.TrimSpace(s)
}

func verifyRAGAnswer(answer string, fragments []RAGFragment, locale string) (bool, string) {
	if answer == "" {
		return false, "Empty answer"
	}
	var ctxBuilder strings.Builder
	for _, f := range fragments {
		ctxBuilder.WriteString(f.Content)
		ctxBuilder.WriteByte('\n')
	}
	contextText := ctxBuilder.String()
	body := answerBodyForVerification(answer, locale)

	mode := GuardrailsModeLocal
	piiBlock := false
	tenantID := "default"
	if config != nil {
		mode = normalizeGuardrailsMode(config.GuardrailsMode)
		piiBlock = config.GuardrailsPIIBlock
		if config.DefaultTenantID != "" {
			tenantID = config.DefaultTenantID
		}
	}

	switch mode {
	case GuardrailsModeRemote:
		ok, reason, err := verifyViaGuardrails(body, contextText, tenantID, piiBlock)
		if err != nil {
			return false, fmt.Sprintf("Guardrails unavailable: %v", err)
		}
		return ok, reason
	case GuardrailsModeHybrid:
		ok, reason, err := verifyViaGuardrails(body, contextText, tenantID, piiBlock)
		if err == nil {
			return ok, reason
		}
		log.Printf("Guardrails hybrid fallback to local: %v", err)
		return verifyRAGAnswerLocal(body, contextText)
	default:
		return verifyRAGAnswerLocal(body, contextText)
	}
}

// verifyRAGAnswerLocal is the in-process numeric check (Spec v1 behavior).
func verifyRAGAnswerLocal(body, contextText string) (bool, string) {
	numsAns := extractNumbersFromText(body)
	if len(numsAns) == 0 {
		return true, "Verification passed"
	}
	numsCtx := extractNumbersFromText(contextText)
	var missing []float64
	for _, n := range numsAns {
		found := false
		for _, c := range numsCtx {
			if math.Abs(n-c) < 0.01 {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, n)
		}
	}
	if len(missing) > 0 {
		return false, fmt.Sprintf("Number(s) %v not found in sources.", missing)
	}
	return true, "Verification passed"
}
