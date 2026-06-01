package main

import (
	"fmt"
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
	var ctx strings.Builder
	for _, f := range fragments {
		ctx.WriteString(f.Content)
		ctx.WriteByte('\n')
	}
	numsAns := extractNumbersFromText(answerBodyForVerification(answer, locale))
	if len(numsAns) == 0 {
		return true, "Verification passed"
	}
	numsCtx := extractNumbersFromText(ctx.String())
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
