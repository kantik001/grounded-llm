package llm_test

import (
	"testing"

	"grounded_llm_server/internal/config"
	"grounded_llm_server/internal/llm"
	"grounded_llm_server/internal/metrics"
)

func TestResolveProviderOllama(t *testing.T) {
	t.Setenv("LLM_BASE_URL", "")
	t.Setenv("OLLAMA_MODEL", "llama3.2")
	cfg := &config.Config{LLMProvider: "ollama", LLMModel: "openrouter/free"}
	llm.ResolveProvider(cfg)
	if cfg.LLMBaseURL != "http://ollama:11434" {
		t.Fatalf("base URL: got %q", cfg.LLMBaseURL)
	}
	if cfg.LLMAPIKey != "local" {
		t.Fatalf("api key: got %q", cfg.LLMAPIKey)
	}
	if cfg.LLMModel != "llama3.2" {
		t.Fatalf("model: got %q", cfg.LLMModel)
	}
}

func TestResolveProviderVLLM(t *testing.T) {
	t.Setenv("LLM_BASE_URL", "")
	t.Setenv("VLLM_MODEL", "")
	cfg := &config.Config{LLMProvider: "vllm", LLMModel: "openrouter/free"}
	llm.ResolveProvider(cfg)
	if cfg.LLMBaseURL != "http://vllm:8000" {
		t.Fatalf("base URL: got %q", cfg.LLMBaseURL)
	}
	if cfg.LLMAPIKey != "local" {
		t.Fatalf("api key: got %q", cfg.LLMAPIKey)
	}
}

func TestResolveProviderExplicitBase(t *testing.T) {
	t.Setenv("LLM_BASE_URL", "http://localhost:9999/")
	cfg := &config.Config{LLMProvider: "ollama"}
	llm.ResolveProvider(cfg)
	if cfg.LLMBaseURL != "http://localhost:9999" {
		t.Fatalf("base URL: got %q", cfg.LLMBaseURL)
	}
}

func TestResponseCacheKey(t *testing.T) {
	a := llm.ResponseCacheKey("How many days?", "default", "default", "m1")
	b := llm.ResponseCacheKey("How many days?", "default", "default", "m1")
	c := llm.ResponseCacheKey("How many days?", "other", "default", "m1")
	if a != b {
		t.Fatalf("expected stable key")
	}
	if a == c {
		t.Fatalf("domain must change key")
	}
	if a[:9] != "response:" {
		t.Fatalf("prefix: %q", a)
	}
}

func TestEstimateTokens(t *testing.T) {
	if metrics.EstimateTokens("") != 0 {
		t.Fatal("empty")
	}
	if metrics.EstimateTokens("abcd") != 1 {
		t.Fatalf("got %d", metrics.EstimateTokens("abcd"))
	}
}
