package rag

import (
	"context"

	"grounded_llm_server/internal/config"
)

var lookupConfig = config.Current

// BindConfig overrides where this package reads process config (app sets app.config).
func BindConfig(fn func() *config.Config) {
	if fn != nil {
		lookupConfig = fn
	}
}

func cfg() *config.Config {
	return lookupConfig()
}

// DomainLocale provides domain guards and locale strings owned by app until those packages extract.
type DomainLocale interface {
	NormalizeDomainID(raw string) (string, error)
	RequireRAGEnabled(domainID string) error
	PublicAPIError(err error) string
	Prompts(domainID, locale string) (system, taskIntro string)
	RAGConstraints(locale string) string
	VerifyFailHint(locale string) string
	Disclaimer(locale string) string
}

var host DomainLocale

// BindHost wires domain/locale helpers from the app layer.
func BindHost(h DomainLocale) {
	host = h
}

// Optional path-trace hooks (wired from app).
var (
	OnTraceStep      func(ctx context.Context, name string, fields map[string]any)
	OnTraceElapsedMS func(ctx context.Context) int64
	OnTraceRequestID func(ctx context.Context) string
)

func traceStep(ctx context.Context, name string, fields map[string]any) {
	if OnTraceStep != nil {
		OnTraceStep(ctx, name, fields)
	}
}

func traceElapsedMS(ctx context.Context) int64 {
	if OnTraceElapsedMS != nil {
		return OnTraceElapsedMS(ctx)
	}
	return 0
}

func traceRequestID(ctx context.Context) string {
	if OnTraceRequestID != nil {
		return OnTraceRequestID(ctx)
	}
	return ""
}
