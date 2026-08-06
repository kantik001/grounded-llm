package app

import (
	"time"

	"grounded_llm_server/internal/store"
)

var retentionWorker *store.RetentionWorker

// startRetentionWorker runs periodic purge of old messages and idle sessions when configured.
func startRetentionWorker(cfg *Config) {
	retentionWorker = store.StartRetentionWorker(chatStore, cfg)
}

func stopRetentionWorker() {
	if retentionWorker != nil {
		retentionWorker.Stop(3 * time.Second)
		retentionWorker = nil
	}
}
