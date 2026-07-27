package main

import (
	"os"
	"strings"
)

// resolveLLMProvider sets OpenAI-compatible base URL / defaults for openai | ollama | vllm.
func resolveLLMProvider(cfg *Config) {
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

func llmAPIKeyConfigured() bool {
	if config == nil {
		return false
	}
	if config.LLMMock {
		return true
	}
	return strings.TrimSpace(config.LLMAPIKey) != ""
}
