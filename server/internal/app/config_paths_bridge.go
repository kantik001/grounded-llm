package app

import cfgpkg "grounded_llm_server/internal/config"

func resolveConfigPath(envKey string, candidates ...string) string {
	return cfgpkg.ResolvePath(envKey, candidates...)
}

func defaultConfigCandidates(name string) []string {
	return cfgpkg.DefaultCandidates(name)
}
