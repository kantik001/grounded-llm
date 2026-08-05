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
	ragPrepared              = rag.Prepared
)

func fetchRAGContext(ctx context.Context, question, tenantID, domainID, locale string) (*pythonRAGContextResponse, error) {
	return rag.FetchContext(ctx, question, tenantID, domainID, locale)
}

func buildRAGUserPrompt(question, contextText, fewShot, taskIntro, constraints string) string {
	return rag.BuildUserPrompt(question, contextText, fewShot, taskIntro, constraints)
}

func ragMockEnabled() bool {
	return rag.MockEnabled()
}

func mockRAGContextResponse(question, domainID string) *pythonRAGContextResponse {
	return rag.MockContextResponse(question, domainID)
}

func publicCitations(fragments []RAGFragment) []RAGFragment {
	return rag.PublicCitations(fragments)
}

func extractNumbersFromText(s string) []float64 {
	return rag.ExtractNumbersFromText(s)
}

func cleanRAGAnswer(text string) string {
	return rag.CleanAnswer(text)
}

func appendRAGDisclaimer(answer, locale string) string {
	return rag.AppendDisclaimer(answer, locale)
}

func verifyRAGAnswer(answer string, fragments []RAGFragment, locale string) (bool, string) {
	return rag.VerifyAnswer(answer, fragments, locale)
}

func verifyRAGAnswerLocal(body, contextText string) (bool, string) {
	return rag.VerifyAnswerLocal(body, contextText)
}

func logRAGOutcome(ctx context.Context, domainID, question string, fragmentCount int, verifyPass bool, verifyReason, sessionID string, softFail bool) {
	rag.LogOutcome(ctx, domainID, question, fragmentCount, verifyPass, verifyReason, sessionID, softFail)
}

func prepareRAGMessages(ctx context.Context, q, domainID, tenantID, locale string, history []Message, sessionID string) (ragPrepared, error) {
	return rag.PrepareMessages(ctx, q, domainID, tenantID, locale, history, sessionID)
}

func finalizeRAGAnswer(ctx context.Context, raw string, p ragPrepared) RAGAnswerResult {
	return rag.FinalizeAnswer(ctx, raw, p)
}

func answerWithRAG(ctx context.Context, q, tenantID, domainID, locale string, history []Message, sessionID string) RAGAnswerResult {
	return rag.Answer(ctx, q, tenantID, domainID, locale, history, sessionID)
}
