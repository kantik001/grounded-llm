package rag

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Faithfulness check: every substantive answer sentence must be lexically
// supported by the retrieved fragments. This extends the numeric-only Spec v1
// verifier to catch fabricated entities and claims that contain no digits.
//
// Matching is stem-based (first stemPrefixLen runes of each content word) so
// Russian/English inflection ("сотрудника"/"сотрудников") still counts as
// support. It is intentionally lexical — an NLI model can replace scoreSupport
// later without changing the wiring.

const (
	// FaithfulnessOff disables the check entirely.
	FaithfulnessOff = "off"
	// FaithfulnessWarn logs unsupported sentences but passes verification.
	FaithfulnessWarn = "warn"
	// FaithfulnessEnforce fails verification on unsupported sentences.
	FaithfulnessEnforce = "enforce"

	stemPrefixLen = 5
	// Sentences with fewer content tokens carry too little signal to judge.
	minSentenceTokens = 4
)

var (
	reSentenceSplit = regexp.MustCompile(`[.!?…]+\s+|[.!?…]+$|\n+`)
	reContentWord   = regexp.MustCompile(`[\p{L}\p{N}]+`)
)

// FaithfulnessMode reads VERIFY_FAITHFULNESS (off|warn|enforce, default warn).
func FaithfulnessMode() string {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("VERIFY_FAITHFULNESS"))) {
	case FaithfulnessEnforce:
		return FaithfulnessEnforce
	case FaithfulnessOff:
		return FaithfulnessOff
	default:
		return FaithfulnessWarn
	}
}

// FaithfulnessMinSupport reads VERIFY_FAITHFULNESS_MIN_SUPPORT (0..1, default 0.5).
func FaithfulnessMinSupport() float64 {
	raw := strings.TrimSpace(os.Getenv("VERIFY_FAITHFULNESS_MIN_SUPPORT"))
	if raw == "" {
		return 0.5
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil || v <= 0 || v > 1 {
		return 0.5
	}
	return v
}

func stemToken(word string) string {
	runes := []rune(strings.ToLower(word))
	if len(runes) > stemPrefixLen {
		runes = runes[:stemPrefixLen]
	}
	return string(runes)
}

// contentStems returns stems of substantive tokens (words of 4+ runes and numbers).
func contentStems(text string) []string {
	var out []string
	for _, w := range reContentWord.FindAllString(text, -1) {
		runes := []rune(w)
		isNumber := strings.IndexFunc(w, func(r rune) bool { return r < '0' || r > '9' }) == -1
		if !isNumber && len(runes) < 4 {
			continue
		}
		out = append(out, stemToken(w))
	}
	return out
}

func splitSentences(text string) []string {
	var out []string
	for _, s := range reSentenceSplit.Split(text, -1) {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// scoreSupport is the fraction of sentence stems present in the context stem set.
func scoreSupport(sentenceStems []string, contextStems map[string]struct{}) float64 {
	if len(sentenceStems) == 0 {
		return 1
	}
	found := 0
	for _, s := range sentenceStems {
		if _, ok := contextStems[s]; ok {
			found++
		}
	}
	return float64(found) / float64(len(sentenceStems))
}

// UnsupportedSentences returns answer sentences whose lexical support in the
// context falls below minSupport. Short sentences are skipped.
func UnsupportedSentences(body, contextText string, minSupport float64) []string {
	contextStems := make(map[string]struct{})
	for _, s := range contentStems(contextText) {
		contextStems[s] = struct{}{}
	}
	var unsupported []string
	for _, sentence := range splitSentences(body) {
		stems := contentStems(sentence)
		if len(stems) < minSentenceTokens {
			continue
		}
		if scoreSupport(stems, contextStems) < minSupport {
			unsupported = append(unsupported, sentence)
		}
	}
	return unsupported
}
