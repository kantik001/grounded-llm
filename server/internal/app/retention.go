package app

import "grounded_llm_server/internal/store"

// startRetentionWorker runs periodic purge of old messages and idle sessions when configured.
func startRetentionWorker(cfg *Config) {
	store.StartRetentionWorker(chatStore, cfg)
}
