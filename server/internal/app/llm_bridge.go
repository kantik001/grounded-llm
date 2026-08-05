package app

import (
	"context"

	"grounded_llm_server/internal/llm"
)

func init() {
	llm.BindConfig(func() *Config { return config })
}

type (
	LLMRequest  = llm.Request
	LLMResponse = llm.Response
	LLMUsage    = llm.Usage
	Choice      = llm.Choice
)

type cachedLLMAnswer = llm.CachedAnswer

func initRedis() {
	llm.InitRedis()
}

func resolveLLMProvider(cfg *Config) {
	llm.ResolveProvider(cfg)
}

func llmMockEnabled() bool {
	return llm.MockEnabled()
}

func mockLLMCompletion(messages []Message) (string, error) {
	return llm.MockComplete(messages)
}

func callLLMCompletion(messages []Message) (string, error) {
	return llm.Complete(messages)
}

func callLLMCompletionForTenant(tenantID string, messages []Message) (string, error) {
	return llm.CompleteForTenant(tenantID, messages)
}

func callLLMCompletionStreamForTenant(ctx context.Context, tenantID string, messages []Message, onDelta func(string) error) (string, error) {
	return llm.CompleteStreamForTenant(ctx, tenantID, messages, onDelta)
}

func responseCacheKey(query, domainID, tenantID, model string) string {
	return llm.ResponseCacheKey(query, domainID, tenantID, model)
}

func getCachedLLMAnswer(ctx context.Context, query, domainID, tenantID string) (*cachedLLMAnswer, bool) {
	return llm.GetCachedAnswer(ctx, query, domainID, tenantID)
}

func setCachedLLMAnswer(ctx context.Context, query, domainID, tenantID string, answer string, citations []RAGFragment) {
	llm.SetCachedAnswer(ctx, query, domainID, tenantID, answer, citations)
}
