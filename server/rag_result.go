package main

import "strings"

// RAGAnswerResult — ответ RAG-пайплайна с источниками для UI.
type RAGAnswerResult struct {
	Answer    string
	Citations []RAGFragment
	OK        bool
	ErrMsg    string
	SoftFail  bool
}

func publicCitations(fragments []RAGFragment) []RAGFragment {
	if len(fragments) == 0 {
		return nil
	}
	out := make([]RAGFragment, len(fragments))
	for i, f := range fragments {
		out[i] = RAGFragment{
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
