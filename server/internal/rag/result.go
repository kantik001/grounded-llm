package rag

import (
	"strings"

	"grounded_llm_server/internal/store"
)

// AnswerResult — ответ RAG-пайплайна с источниками для UI.
type AnswerResult struct {
	Answer        string
	Citations     []store.RAGFragment
	OK            bool
	ErrMsg        string
	SoftFail      bool
	VerifyPass    bool
	FragmentCount int
	CacheHit      bool // semantic LLM response cache
}

// PublicCitations returns UI-safe citation excerpts (no full chunk bodies).
func PublicCitations(fragments []store.RAGFragment) []store.RAGFragment {
	if len(fragments) == 0 {
		return nil
	}
	out := make([]store.RAGFragment, len(fragments))
	for i, f := range fragments {
		out[i] = store.RAGFragment{
			Filename: f.Filename,
			Page:     f.Page,
			Excerpt:  excerptForUI(f.Content),
		}
	}
	return out
}

func excerptForUI(content string) string {
	const maxLen = 280
	s := strings.TrimSpace(content)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "…"
}
