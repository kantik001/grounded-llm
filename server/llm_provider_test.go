package main

import (
	"testing"
)

func TestResolveLLMProviderOllama(t *testing.T) {
	t.Setenv("LLM_BASE_URL", "")
	t.Setenv("OLLAMA_MODEL", "llama3.2")
	cfg := &Config{LLMProvider: "ollama", LLMModel: "openrouter/free"}
	resolveLLMProvider(cfg)
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

func TestResolveLLMProviderVLLM(t *testing.T) {
	t.Setenv("LLM_BASE_URL", "")
	t.Setenv("VLLM_MODEL", "")
	cfg := &Config{LLMProvider: "vllm", LLMModel: "openrouter/free"}
	resolveLLMProvider(cfg)
	if cfg.LLMBaseURL != "http://vllm:8000" {
		t.Fatalf("base URL: got %q", cfg.LLMBaseURL)
	}
	if cfg.LLMAPIKey != "local" {
		t.Fatalf("api key: got %q", cfg.LLMAPIKey)
	}
}

func TestResolveLLMProviderExplicitBase(t *testing.T) {
	t.Setenv("LLM_BASE_URL", "http://localhost:9999/")
	cfg := &Config{LLMProvider: "ollama"}
	resolveLLMProvider(cfg)
	if cfg.LLMBaseURL != "http://localhost:9999" {
		t.Fatalf("base URL: got %q", cfg.LLMBaseURL)
	}
}

func TestResponseCacheKey(t *testing.T) {
	a := responseCacheKey("How many days?", "default", "default", "m1")
	b := responseCacheKey("How many days?", "default", "default", "m1")
	c := responseCacheKey("How many days?", "other", "default", "m1")
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
	if estimateTokens("") != 0 {
		t.Fatal("empty")
	}
	if estimateTokens("abcd") != 1 {
		t.Fatalf("got %d", estimateTokens("abcd"))
	}
}
