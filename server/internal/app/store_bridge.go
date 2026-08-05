package app

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"grounded_llm_server/internal/store"
)

// Type aliases — store owns persistence models.
type (
	ChatStore           = store.ChatStore
	TelegramUser        = store.TelegramUser
	ChatMessage         = store.ChatMessage
	RAGFragment         = store.RAGFragment
	Message             = store.Message
	TenantPurgeStats    = store.TenantPurgeStats
	FeedbackSummaryRow  = store.FeedbackSummaryRow
	AnalyticsDashboard  = store.AnalyticsDashboard
	QuestionsPerDayRow    = store.QuestionsPerDayRow
	DomainQuestionCount = store.DomainQuestionCount
	KBGapRow            = store.KBGapRow
	RAGStats            = store.RAGStats
	ReindexJob          = store.ReindexJob
)

func parseAnalyticsDays(s string) int {
	return store.ParseAnalyticsDays(s)
}

func computeVerifyPassRate(pass, fail int64) float64 {
	return store.ComputeVerifyPassRate(pass, fail)
}

func newChatStore(ctx context.Context, databaseURL, uploadDir string) (*ChatStore, error) {
	return store.New(ctx, databaseURL, uploadDir)
}

func waitForPostgres(ctx context.Context, databaseURL string, attempts int) (*pgxpool.Pool, error) {
	return store.WaitForPostgres(ctx, databaseURL, attempts)
}

func runAllMigrations(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	return store.RunAllMigrations(ctx, pool, dir)
}

func findMigrationsDir() (string, error) {
	return store.FindMigrationsDir()
}
