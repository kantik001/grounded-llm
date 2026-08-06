package store

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
)

// LogAnalyticsEvent records a product analytics event (user_id is optional).
func (st *ChatStore) LogAnalyticsEvent(ctx context.Context, telegramID int64, eventType string, payload map[string]any) error {
	if payload == nil {
		payload = map[string]any{}
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	var userID *int64
	if telegramID > 0 {
		var uid int64
		err = st.Pool.QueryRow(ctx, `SELECT id FROM users WHERE telegram_id = $1`, telegramID).Scan(&uid)
		if err == nil {
			userID = &uid
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
	}
	_, err = st.Pool.Exec(ctx,
		`INSERT INTO analytics_events (user_id, event_type, payload) VALUES ($1, $2, $3)`,
		userID, eventType, raw,
	)
	return err
}

// LogEvent records a product event (no PII in payload).
func (st *ChatStore) LogEvent(ctx context.Context, telegramID int64, eventType string, payload map[string]any) error {
	return st.LogAnalyticsEvent(ctx, telegramID, eventType, payload)
}

// SaveMessageFeedback stores an assistant reply rating (1 or -1) for the tenant.
func (st *ChatStore) SaveMessageFeedback(ctx context.Context, telegramID int64, tenantID string, messageID int64, rating int) error {
	if rating != 1 && rating != -1 {
		return errors.New("rating must be 1 or -1")
	}
	var userID int64
	err := st.Pool.QueryRow(ctx, `SELECT id FROM users WHERE telegram_id = $1`, telegramID).Scan(&userID)
	if err != nil {
		return err
	}
	var ok bool
	err = st.Pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM messages m
			JOIN chat_sessions cs ON cs.id = m.session_id
			WHERE m.id = $1 AND m.role = 'assistant' AND cs.user_id = $2 AND cs.tenant_id = $3
		)`, messageID, userID, tenantID,
	).Scan(&ok)
	if err != nil {
		return err
	}
	if !ok {
		return errSessionNotFound
	}
	_, err = st.Pool.Exec(ctx, `
		INSERT INTO message_feedback (message_id, user_id, rating)
		VALUES ($1, $2, $3)
		ON CONFLICT (message_id, user_id) DO UPDATE SET rating = EXCLUDED.rating, created_at = NOW()`,
		messageID, userID, rating,
	)
	return err
}
