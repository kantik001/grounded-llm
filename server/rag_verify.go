package main

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// ragAnswerDisclaimer — общий дисклеймер в конце ответа (без названий статей).
// Синхронизировать с rag/verifier.py (RAG_ANSWER_DISCLAIMER) и tests/test_verifier.py.
const ragAnswerDisclaimer = "Справочная информация из базы знаний. Не заменяет официальную консультацию ответственного специалиста."

var (
	reNumberWord = regexp.MustCompile(`\b\d+(?:\.\d+)?\b`)
	reMultiSpace = regexp.MustCompile(`\s+`)
	reThink      = regexp.MustCompile(`(?i)</?think>`)
	reAnswerTag  = regexp.MustCompile(`(?i)</?answer>`)
	reSystemTag  = regexp.MustCompile(`(?i)</?system>`)
	reAbot       = regexp.MustCompile(`(?i)\bаботчик\b`)
	reIntro      = regexp.MustCompile(`(?i)^(Хорошо|Давайте посмотрим|Итак|Я думаю|мне нужно ответить|Из контекста видно|Теперь я понимаю|Из таблицы видно)[,:.]?\s*`)
	reSourceLine = regexp.MustCompile(`(?im)^\s*Источник:.*\n?`)
)

// Извлекает числа из текста для верификации ответа RAG.
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

// Убирает служебные теги и вводные фразы из ответа LLM.
func cleanRAGAnswer(text string) string {
	if text == "" {
		return "Ответ не сформирован корректно."
	}
	text = reThink.ReplaceAllString(text, "")
	text = reAnswerTag.ReplaceAllString(text, "")
	text = reSystemTag.ReplaceAllString(text, "")
	text = reAbot.ReplaceAllString(text, "")
	text = reIntro.ReplaceAllString(text, "")
	text = strings.TrimSpace(reMultiSpace.ReplaceAllString(text, " "))
	if text == "" {
		return "Ответ не сформирован корректно."
	}
	return text
}

// Удаляет строки «Источник:» из ответа перед показом пользователю.
func stripSourceAttribution(answer string) string {
	s := reSourceLine.ReplaceAllString(answer, "")
	return strings.TrimSpace(reMultiSpace.ReplaceAllString(s, " "))
}

// Добавляет дисклеймер в конец ответа RAG.
func appendRAGDisclaimer(answer string) string {
	body := stripSourceAttribution(answer)
	if body == "" {
		return ragAnswerDisclaimer
	}
	if strings.Contains(body, ragAnswerDisclaimer) || strings.Contains(body, "Не заменяет официальную консультацию") {
		return body
	}
	return body + "\n\n" + ragAnswerDisclaimer
}

// Текст ответа без дисклеймера и источников — для проверки чисел.
func answerBodyForVerification(answer string) string {
	s := stripSourceAttribution(answer)
	s = strings.ReplaceAll(s, ragAnswerDisclaimer, "")
	return strings.TrimSpace(s)
}

// Проверяет, что все числа в ответе есть во фрагментах статей.
func verifyRAGAnswer(answer string, fragments []RAGFragment) (bool, string) {
	if answer == "" {
		return false, "Ответ отсутствует"
	}
	var ctx strings.Builder
	for _, f := range fragments {
		ctx.WriteString(f.Content)
		ctx.WriteByte('\n')
	}
	numsAns := extractNumbersFromText(answerBodyForVerification(answer))
	if len(numsAns) == 0 {
		return true, "Верификация пройдена"
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
		return false, fmt.Sprintf("Число(а) %v не найдены в источниках.", missing)
	}
	return true, "Верификация пройдена"
}
