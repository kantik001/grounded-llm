package config

import (
	"os"
	"path/filepath"
	"strings"
)

// ResolvePath picks the first existing file from env or candidate paths.
func ResolvePath(envKey string, candidates ...string) string {
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

func resolveConfigPath(envKey string, candidates ...string) string {
	return ResolvePath(envKey, candidates...)
}

// DefaultCandidates lists typical config file locations for different CWDs.
func DefaultCandidates(name string) []string {
	// CWD may be repo root, server/, or server/internal/app/ (go test).
	return []string{
		filepath.Join("/config", name),
		filepath.Join("config", name),
		filepath.Join("..", "config", name),
		filepath.Join("..", "..", "config", name),
		filepath.Join("..", "..", "..", "config", name),
	}
}

func defaultConfigCandidates(name string) []string {
	return DefaultCandidates(name)
}
