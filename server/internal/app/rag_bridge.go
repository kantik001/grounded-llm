package app

import (
	"context"
	"strings"

	"grounded_llm_server/internal/rag"
)

func init() {
	rag.BindConfig(func() *Config { return config })
	rag.BindHost(ragHost{})
	rag.OnTraceStep = func(ctx context.Context, name string, fields map[string]any) {
		if tr := pathTraceFrom(ctx); tr != nil {
			tr.Step(name, fields)
		}
	}
	rag.OnTraceElapsedMS = func(ctx context.Context) int64 {
		if tr := pathTraceFrom(ctx); tr != nil {
			return tr.ElapsedMS()
		}
		return 0
	}
	rag.OnTraceRequestID = func(ctx context.Context) string {
		if tr := pathTraceFrom(ctx); tr != nil {
			return tr.RequestID()
		}
		return ""
	}
}

type ragHost struct{}

func (ragHost) NormalizeDomainID(raw string) (string, error) { return normalizeDomainID(raw) }
func (ragHost) RequireRAGEnabled(domainID string) error      { return requireRAGEnabled(domainID) }
func (ragHost) PublicAPIError(err error) string              { return publicAPIError(err) }
func (ragHost) Prompts(domainID, locale string) (string, string) {
	p := promptsForDomainLocale(domainID, locale)
	return p.RAGSystem, p.RAGTaskIntro
}
func (ragHost) RAGConstraints(locale string) string { return ragConstraintsForLocale(locale) }
func (ragHost) VerifyFailHint(locale string) string { return verifyFailHintForLocale(locale) }
func (ragHost) Disclaimer(locale string) string {
	return strings.TrimSpace(brandingForLocale(locale).Disclaimer)
}

type (
	RAGAnswerResult          = rag.AnswerResult
	pythonRAGContextResponse = rag.ContextResponse
)

func fetchRAGContext(ctx context.Context, question, tenantID, domainID, locale string) (*pythonRAGContextResponse, error) {
	return rag.FetchContext(ctx, question, tenantID, domainID, locale)
}

func ragMockEnabled() bool {
	return rag.MockEnabled()
}

func mockRAGContextResponse(question, domainID string) *pythonRAGContextResponse {
	return rag.MockContextResponse(question, domainID)
}

func answerWithRAG(ctx context.Context, q, tenantID, domainID, locale string, history []Message, sessionID string) RAGAnswerResult {
	return rag.Answer(ctx, q, tenantID, domainID, locale, history, sessionID)
}
