package llm

import cfgpkg "grounded_llm_server/internal/config"

// ResolveProvider applies Ollama/vLLM defaults onto cfg.
func ResolveProvider(cfg *cfgpkg.Config) {
	cfgpkg.ResolveLLMProvider(cfg)
}
