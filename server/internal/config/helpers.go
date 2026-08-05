package config

import (
	"os"
	"strings"
)

func parseAllowedOrigins(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		p := strings.TrimSpace(part)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// IsTruthyEnv reports whether env key is a common truthy value.
func IsTruthyEnv(key string) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func isTruthyEnv(key string) bool { return IsTruthyEnv(key) }

// ResolveLLMProvider sets OpenAI-compatible base URL / defaults for openai | ollama | vllm.
func ResolveLLMProvider(cfg *Config) {
	if cfg == nil {
		return
	}
	provider := strings.ToLower(strings.TrimSpace(cfg.LLMProvider))
	if provider == "" {
		provider = "openai"
	}
	cfg.LLMProvider = provider

	explicitBase := strings.TrimSpace(os.Getenv("LLM_BASE_URL"))
	if explicitBase != "" {
		cfg.LLMBaseURL = strings.TrimRight(explicitBase, "/")
	} else {
		switch provider {
		case "ollama":
			cfg.LLMBaseURL = "http://ollama:11434"
		case "vllm":
			cfg.LLMBaseURL = "http://vllm:8000"
		default:
			cfg.LLMBaseURL = "https://openrouter.ai/api"
		}
	}

	if (provider == "ollama" || provider == "vllm") && strings.TrimSpace(cfg.LLMAPIKey) == "" {
		cfg.LLMAPIKey = "local"
	}
	if provider == "ollama" {
		if m := strings.TrimSpace(os.Getenv("OLLAMA_MODEL")); m != "" {
			cfg.LLMModel = m
		} else if cfg.LLMModel == "" || cfg.LLMModel == "openrouter/free" {
			cfg.LLMModel = "llama3.2"
		}
	}
	if provider == "vllm" {
		if m := strings.TrimSpace(os.Getenv("VLLM_MODEL")); m != "" {
			cfg.LLMModel = m
		} else if cfg.LLMModel == "" || cfg.LLMModel == "openrouter/free" {
			cfg.LLMModel = "meta-llama/Meta-Llama-3.1-8B-Instruct"
		}
	}
}

func resolveLLMProvider(cfg *Config) { ResolveLLMProvider(cfg) }

func normalizeGuardrailsMode(s string) string {
	m := strings.ToLower(strings.TrimSpace(s))
	switch m {
	case "remote", "hybrid", "local":
		return m
	default:
		return "local"
	}
}
