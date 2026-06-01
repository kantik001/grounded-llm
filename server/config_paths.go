package main

import (
	"os"
	"path/filepath"
	"strings"
)

// resolveConfigPath picks the first existing file from env or candidate paths.
func resolveConfigPath(envKey string, candidates ...string) string {
	if p := strings.TrimSpace(os.Getenv(envKey)); p != "" {
		return p
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	if len(candidates) > 0 {
		return candidates[len(candidates)-1]
	}
	return ""
}

func defaultConfigCandidates(name string) []string {
	return []string{
		filepath.Join("/config", name),
		filepath.Join("..", "config", name),
		filepath.Join("config", name),
	}
}
