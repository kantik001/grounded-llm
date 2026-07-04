package main

import (
	"context"
	"log"
	"time"
)

// startRetentionWorker runs periodic purge of old messages and idle sessions when configured.
func startRetentionWorker(cfg *Config) {
	if cfg.MessageRetentionDays <= 0 && cfg.SessionRetentionDays <= 0 {
		return
	}
	interval := time.Duration(cfg.RetentionIntervalHours) * time.Hour
	go func() {
		runRetentionOnce(cfg)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			runRetentionOnce(cfg)
		}
	}()
	log.Printf("Retention worker started (every %s)", interval)
}

func runRetentionOnce(cfg *Config) {
	if chatStore == nil || chatStore.pool == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if cfg.MessageRetentionDays > 0 {
		n, err := chatStore.PurgeMessagesOlderThan(ctx, cfg.MessageRetentionDays)
		if err != nil {
			log.Printf("Retention: purge messages: %v", err)
		} else if n > 0 {
			log.Printf("Retention: purged %d message(s) older than %d days", n, cfg.MessageRetentionDays)
		}
	}
	if cfg.SessionRetentionDays > 0 {
		n, err := chatStore.PurgeSessionsOlderThan(ctx, cfg.SessionRetentionDays)
		if err != nil {
			log.Printf("Retention: purge sessions: %v", err)
		} else if n > 0 {
			log.Printf("Retention: purged %d session(s) idle longer than %d days", n, cfg.SessionRetentionDays)
		}
	}
}
