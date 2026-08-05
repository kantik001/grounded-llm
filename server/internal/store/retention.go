package store

import (
	"context"
	"log"
	"time"

	"grounded_llm_server/internal/config"
)

// RunRetentionOnce purges old messages and idle sessions when retention days are configured.
func RunRetentionOnce(st *ChatStore, cfg *config.Config) {
	if st == nil || st.Pool == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if cfg.MessageRetentionDays > 0 {
		n, err := st.PurgeMessagesOlderThan(ctx, cfg.MessageRetentionDays)
		if err != nil {
			log.Printf("Retention: purge messages: %v", err)
		} else if n > 0 {
			log.Printf("Retention: purged %d message(s) older than %d days", n, cfg.MessageRetentionDays)
		}
	}
	if cfg.SessionRetentionDays > 0 {
		n, err := st.PurgeSessionsOlderThan(ctx, cfg.SessionRetentionDays)
		if err != nil {
			log.Printf("Retention: purge sessions: %v", err)
		} else if n > 0 {
			log.Printf("Retention: purged %d session(s) idle longer than %d days", n, cfg.SessionRetentionDays)
		}
	}
}

// StartRetentionWorker runs periodic retention purges when configured.
func StartRetentionWorker(st *ChatStore, cfg *config.Config) {
	if cfg.MessageRetentionDays <= 0 && cfg.SessionRetentionDays <= 0 {
		return
	}
	interval := time.Duration(cfg.RetentionIntervalHours) * time.Hour
	go func() {
		RunRetentionOnce(st, cfg)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			RunRetentionOnce(st, cfg)
		}
	}()
	log.Printf("Retention worker started (every %s)", interval)
}
