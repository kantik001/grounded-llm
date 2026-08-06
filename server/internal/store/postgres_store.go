package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const maxSessionMessages = 80

// ChatStore is persistent chat storage (PostgreSQL + on-disk files).
type ChatStore struct {
	Pool      *pgxpool.Pool
	uploadDir string
}

// New connects to Postgres and returns a ChatStore with the upload directory.
func New(ctx context.Context, databaseURL, uploadDir string) (*ChatStore, error) {
	if strings.TrimSpace(databaseURL) == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return nil, fmt.Errorf("upload dir: %w", err)
	}
	pool, err := openPool(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("postgres connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres ping: %w", err)
	}
	return &ChatStore{Pool: pool, uploadDir: uploadDir}, nil
}

// Close shuts down the PostgreSQL connection pool.
func (st *ChatStore) Close() {
	if st != nil && st.Pool != nil {
		st.Pool.Close()
	}
}

// runMigrations applies a single SQL migration file inside a transaction together
// with the schema_migrations insert, so a partial apply cannot be marked done.
func runMigrations(ctx context.Context, pool *pgxpool.Pool, sqlPath, filename string) error {
	body, err := os.ReadFile(sqlPath)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", sqlPath, err)
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migration tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err = tx.Exec(ctx, string(body)); err != nil {
		return fmt.Errorf("apply migration: %w", err)
	}
	if _, err = tx.Exec(ctx,
		`INSERT INTO schema_migrations (filename) VALUES ($1) ON CONFLICT DO NOTHING`, filename,
	); err != nil {
		return fmt.Errorf("record migration: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit migration: %w", err)
	}
	return nil
}

func ensureSchemaMigrationsTable(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename   TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`)
	return err
}

func migrationApplied(ctx context.Context, pool *pgxpool.Pool, filename string) (bool, error) {
	var n int
	err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM schema_migrations WHERE filename = $1`, filename,
	).Scan(&n)
	return n > 0, err
}

func markMigrationApplied(ctx context.Context, pool *pgxpool.Pool, filename string) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO schema_migrations (filename) VALUES ($1) ON CONFLICT DO NOTHING`, filename,
	)
	return err
}

// RunAllMigrations applies all .sql files from dir in name order (tracks schema_migrations).
// Each file is applied and recorded in a single transaction.
func RunAllMigrations(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	if err := ensureSchemaMigrationsTable(ctx, pool); err != nil {
		return fmt.Errorf("schema_migrations table: %w", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir %s: %w", dir, err)
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".sql") {
			files = append(files, filepath.Join(dir, name))
		}
	}
	sort.Strings(files)
	if len(files) == 0 {
		return fmt.Errorf("no .sql migrations in %s", dir)
	}
	for _, f := range files {
		base := filepath.Base(f)
		applied, err := migrationApplied(ctx, pool, base)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", base, err)
		}
		if applied {
			log.Printf("Skip migration (already applied): %s", base)
			continue
		}
		if err := runMigrations(ctx, pool, f, base); err != nil {
			return fmt.Errorf("%s: %w", f, err)
		}
		log.Printf("Applied migration: %s", base)
	}
	return nil
}

// FindMigrationsDir locates the migrations directory (MIGRATIONS_DIR or common paths).
func FindMigrationsDir() (string, error) {
	if p := os.Getenv("MIGRATIONS_DIR"); p != "" {
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			return p, nil
		}
	}
	for _, candidate := range []string{
		"/migrations",
		filepath.Join("..", "migrations"),
		filepath.Join("migrations"),
	} {
		if st, err := os.Stat(candidate); err == nil && st.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("migrations directory not found")
}

// newSessionID generates a random chat session id (hex).
func newSessionID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// newImageToken generates a token for an uploaded image URL.
func newImageToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// UpsertUser creates or updates a user by telegram_id.
func (st *ChatStore) UpsertUser(ctx context.Context, u *TelegramUser) (int64, error) {
	var id int64
	err := st.Pool.QueryRow(ctx, `
		INSERT INTO users (telegram_id, username, first_name, last_name, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (telegram_id) DO UPDATE SET
			username = EXCLUDED.username,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			updated_at = NOW()
		RETURNING id`,
		u.ID, nullIfEmpty(u.Username), nullIfEmpty(u.FirstName), nullIfEmpty(u.LastName),
	).Scan(&id)
	return id, err
}

// nullIfEmpty returns SQL NULL for an empty string, otherwise a pointer to the value.
func nullIfEmpty(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

// CreateSession creates a new chat session for a user.
func (st *ChatStore) CreateSession(ctx context.Context, userID int64, tenantID, domainID string) (string, error) {
	sid := newSessionID()
	_, err := st.Pool.Exec(ctx,
		`INSERT INTO chat_sessions (id, user_id, domain_id, tenant_id) VALUES ($1, $2, $3, $4)`,
		sid, userID, domainID, tenantID,
	)
	return sid, err
}

// SessionDomainID returns the session domain_id (with ownership check).
func (st *ChatStore) SessionDomainID(ctx context.Context, sessionID string, telegramID int64, tenantID string) (string, error) {
	var domainID string
	err := st.Pool.QueryRow(ctx, `
		SELECT cs.domain_id FROM chat_sessions cs
		JOIN users u ON u.id = cs.user_id
		WHERE cs.id = $1 AND u.telegram_id = $2 AND cs.tenant_id = $3`,
		sessionID, telegramID, tenantID,
	).Scan(&domainID)
	if err != nil {
		return "", errSessionNotFound
	}
	return domainID, nil
}

// sessionOwned checks that the session belongs to the Telegram user.
func (st *ChatStore) sessionOwned(ctx context.Context, sessionID string, telegramID int64, tenantID string) (bool, error) {
	var ok bool
	err := st.Pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM chat_sessions cs
			JOIN users u ON u.id = cs.user_id
			WHERE cs.id = $1 AND u.telegram_id = $2 AND cs.tenant_id = $3
		)`, sessionID, telegramID, tenantID,
	).Scan(&ok)
	return ok, err
}

// GetOrCreateSession returns an existing session or creates a new one.
func (st *ChatStore) GetOrCreateSession(ctx context.Context, sessionID string, u *TelegramUser, tenantID, domainID string) (string, string, error) {
	userID, err := st.UpsertUser(ctx, u)
	if err != nil {
		return "", "", err
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID != "" {
		owned, err := st.sessionOwned(ctx, sessionID, u.ID, tenantID)
		if err != nil {
			return "", "", err
		}
		if owned {
			domain, err := st.SessionDomainID(ctx, sessionID, u.ID, tenantID)
			if err != nil {
				return "", "", err
			}
			return sessionID, domain, nil
		}
	}
	sid, err := st.CreateSession(ctx, userID, tenantID, domainID)
	return sid, domainID, err
}

// ListMessages returns session history for the UI.
func (st *ChatStore) ListMessages(ctx context.Context, sessionID string, telegramID int64, tenantID string) ([]ChatMessage, error) {
	owned, err := st.sessionOwned(ctx, sessionID, telegramID, tenantID)
	if err != nil {
		return nil, err
	}
	if !owned {
		return nil, errSessionNotFound
	}
	rows, err := st.Pool.Query(ctx, `
		SELECT m.id, m.role, m.content, m.kind, m.image_token, m.class_prediction, m.class_confidence,
		       m.citations, mf.rating
		FROM messages m
		LEFT JOIN users u ON u.telegram_id = $2
		LEFT JOIN message_feedback mf ON mf.message_id = m.id AND mf.user_id = u.id
		WHERE m.session_id = $1
		ORDER BY m.created_at ASC, m.id ASC`, sessionID, telegramID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ChatMessage
	for rows.Next() {
		var m ChatMessage
		var imageToken *string
		var classPred *string
		var classConf *float64
		var citationsJSON []byte
		var fbRating *int16
		if err := rows.Scan(&m.ID, &m.Role, &m.Content, &m.Kind, &imageToken, &classPred, &classConf, &citationsJSON, &fbRating); err != nil {
			return nil, err
		}
		if len(citationsJSON) > 0 {
			_ = json.Unmarshal(citationsJSON, &m.Citations)
		}
		if imageToken != nil && *imageToken != "" {
			m.ImageURL = mediaURL(*imageToken)
		}
		if classPred != nil {
			m.ClassPrediction = *classPred
		}
		if classConf != nil {
			m.ClassConfidence = *classConf
		}
		if fbRating != nil {
			r := int(*fbRating)
			m.FeedbackRating = &r
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// mediaURL builds the public media URL for a token.
func mediaURL(token string) string {
	return "/api/media/" + token
}

// ErrSessionNotFound is returned when a session is missing or not owned by the user.
var ErrSessionNotFound = fmt.Errorf("session not found")

var errSessionNotFound = ErrSessionNotFound

func citationsJSONValue(c []RAGFragment) ([]byte, error) {
	if len(c) == 0 {
		return nil, nil
	}
	return json.Marshal(c)
}

// AppendMessage persists a message and trims history to maxSessionMessages.
func (st *ChatStore) AppendMessage(ctx context.Context, sessionID string, m ChatMessage) (int64, error) {
	citJSON, err := citationsJSONValue(m.Citations)
	if err != nil {
		return 0, fmt.Errorf("citations json: %w", err)
	}
	var id int64
	err = st.Pool.QueryRow(ctx, `
		INSERT INTO messages (session_id, role, content, kind, image_token, class_prediction, class_confidence, citations)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`,
		sessionID, m.Role, m.Content, m.Kind,
		nullToken(m.ImageToken), nullIfEmpty(m.ClassPrediction), nullConfidence(m.ClassConfidence), citJSON,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	_, err = st.Pool.Exec(ctx, `
		DELETE FROM messages
		WHERE session_id = $1
		  AND id NOT IN (
			SELECT id FROM messages
			WHERE session_id = $1
			ORDER BY created_at DESC, id DESC
			LIMIT $2
		  )`, sessionID, maxSessionMessages,
	)
	return id, err
}

// nullToken returns SQL NULL for an empty image_token on INSERT.
func nullToken(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

// nullConfidence returns SQL NULL for zero classification confidence.
func nullConfidence(v float64) *float64 {
	if v <= 0 {
		return nil
	}
	return &v
}

// HistoryForLLM returns recent session messages formatted for the LLM.
func (st *ChatStore) HistoryForLLM(ctx context.Context, sessionID string, telegramID int64, tenantID string, excludeLastN int) ([]Message, error) {
	msgs, err := st.ListMessages(ctx, sessionID, telegramID, tenantID)
	if err != nil {
		return nil, err
	}
	n := len(msgs) - excludeLastN
	if n < 0 {
		n = 0
	}
	var out []Message
	for _, m := range msgs[:n] {
		if msg, ok := m.ToLLMMessage(); ok {
			out = append(out, msg)
		}
	}
	return TrimHistoryMessages(out, 24), nil
}

// SaveImage writes image bytes to disk and returns a URL token.
func (st *ChatStore) SaveImage(data []byte) (string, error) {
	token := newImageToken()
	path := filepath.Join(st.uploadDir, token+".bin")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", err
	}
	return token, nil
}

// UserCanAccessImage checks that the image belongs to the user's message.
func (st *ChatStore) UserCanAccessImage(ctx context.Context, token string, telegramID int64) (bool, error) {
	var ok bool
	err := st.Pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM messages m
			JOIN chat_sessions cs ON cs.id = m.session_id
			JOIN users u ON u.id = cs.user_id
			WHERE m.image_token = $1 AND u.telegram_id = $2
		)`, token, telegramID,
	).Scan(&ok)
	return ok, err
}

// ReadImage returns file bytes for a token.
func (st *ChatStore) ReadImage(token string) ([]byte, error) {
	token = strings.TrimSpace(token)
	if token == "" || strings.Contains(token, "..") || strings.Contains(token, "/") {
		return nil, fmt.Errorf("invalid token")
	}
	return os.ReadFile(filepath.Join(st.uploadDir, token+".bin"))
}

type FeedbackSummaryRow struct {
	Rating int   `json:"rating"`
	Count  int64 `json:"count"`
}

// FeedbackSummary aggregates message feedback ratings.
func (st *ChatStore) FeedbackSummary(ctx context.Context) ([]FeedbackSummaryRow, error) {
	rows, err := st.Pool.Query(ctx, `
		SELECT rating, COUNT(*)::bigint
		FROM message_feedback
		GROUP BY rating
		ORDER BY rating DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []FeedbackSummaryRow
	for rows.Next() {
		var r FeedbackSummaryRow
		if err := rows.Scan(&r.Rating, &r.Count); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// PurgeMessagesOlderThan deletes messages older than the given number of days.
func (st *ChatStore) PurgeMessagesOlderThan(ctx context.Context, days int) (int64, error) {
	if days <= 0 {
		return 0, nil
	}
	tag, err := st.Pool.Exec(ctx, `
		DELETE FROM messages
		WHERE created_at < NOW() - ($1 || ' days')::interval`, days)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// PurgeSessionsOlderThan deletes idle chat sessions (and cascaded messages) older than days.
func (st *ChatStore) PurgeSessionsOlderThan(ctx context.Context, days int) (int64, error) {
	if days <= 0 {
		return 0, nil
	}
	tag, err := st.Pool.Exec(ctx, `
		DELETE FROM chat_sessions
		WHERE updated_at < NOW() - ($1 || ' days')::interval`, days)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// WaitForPostgres waits for Postgres to become ready at startup (e.g. docker compose).
func WaitForPostgres(ctx context.Context, databaseURL string, attempts int) (*pgxpool.Pool, error) {
	var lastErr error
	for i := 0; i < attempts; i++ {
		pool, err := openPool(ctx, databaseURL)
		if err == nil {
			if err = pool.Ping(ctx); err == nil {
				return pool, nil
			}
			pool.Close()
			lastErr = err
		} else {
			lastErr = err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	return nil, fmt.Errorf("postgres not ready after %d attempts: %v", attempts, lastErr)
}
