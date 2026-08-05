package rag

import (
	"fmt"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"

	"grounded_llm_server/internal/guardrails"
	"grounded_llm_server/internal/store"
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
	if host != nil {
		if d := strings.TrimSpace(host.Disclaimer(locale)); d != "" {
			return d
		}
	}
	return "Reference information from the knowledge base. Not a substitute for official expert advice."
}

// ExtractNumbersFromText parses numeric tokens from text (comma as decimal separator).
func ExtractNumbersFromText(s string) []float64 {
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

// CleanAnswer strips think/answer tags and filler intros.
func CleanAnswer(text string) string {
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

// AppendDisclaimer strips Source: lines and appends the locale disclaimer.
func AppendDisclaimer(answer, locale string) string {
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

// VerifyAnswer checks the answer against fragments (local / remote / hybrid).
func VerifyAnswer(answer string, fragments []store.RAGFragment, locale string) (bool, string) {
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

	mode := guardrails.ModeLocal
	piiBlock := false
	tenantID := "default"
	if c := cfg(); c != nil {
		mode = guardrails.NormalizeMode(c.GuardrailsMode)
		piiBlock = c.GuardrailsPIIBlock
		if c.DefaultTenantID != "" {
			tenantID = c.DefaultTenantID
		}
	}

	switch mode {
	case guardrails.ModeRemote:
		ok, reason, err := guardrails.VerifyText(body, contextText, tenantID, piiBlock)
		if err != nil {
			return false, fmt.Sprintf("Guardrails unavailable: %v", err)
		}
		return ok, reason
	case guardrails.ModeHybrid:
		ok, reason, err := guardrails.VerifyText(body, contextText, tenantID, piiBlock)
		if err == nil {
			return ok, reason
		}
		log.Printf("Guardrails hybrid fallback to local: %v", err)
		return VerifyAnswerLocal(body, contextText)
	default:
		return VerifyAnswerLocal(body, contextText)
	}
}

// VerifyAnswerLocal is the in-process numeric check (Spec v1 behavior).
func VerifyAnswerLocal(body, contextText string) (bool, string) {
	numsAns := ExtractNumbersFromText(body)
	if len(numsAns) == 0 {
		return true, "Verification passed"
	}
	numsCtx := ExtractNumbersFromText(contextText)
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
