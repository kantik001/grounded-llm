package store

import (
	"context"
	"log"
	"sync"
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

// RetentionWorker is a stoppable periodic retention loop.
type RetentionWorker struct {
	cancel context.CancelFunc
	done   chan struct{}
	once   sync.Once
}

// StartRetentionWorker runs periodic retention purges when configured.
// Returns nil when retention is disabled.
func StartRetentionWorker(st *ChatStore, cfg *config.Config) *RetentionWorker {
	if cfg.MessageRetentionDays <= 0 && cfg.SessionRetentionDays <= 0 {
		return nil
	}
	interval := time.Duration(cfg.RetentionIntervalHours) * time.Hour
	ctx, cancel := context.WithCancel(context.Background())
	w := &RetentionWorker{cancel: cancel, done: make(chan struct{})}
	go func() {
		defer close(w.done)
		RunRetentionOnce(st, cfg)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				RunRetentionOnce(st, cfg)
			}
		}
	}()
	log.Printf("Retention worker started (every %s)", interval)
	return w
}

// Stop cancels the worker and waits briefly for the loop to exit.
func (w *RetentionWorker) Stop(timeout time.Duration) {
	if w == nil {
		return
	}
	w.once.Do(func() {
		w.cancel()
		select {
		case <-w.done:
		case <-time.After(timeout):
			log.Printf("Retention worker stop timed out after %s", timeout)
		}
	})
}
