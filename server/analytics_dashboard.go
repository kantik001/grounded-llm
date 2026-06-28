package main

import (
	"context"
	"fmt"
	"time"
)

type QuestionsPerDayRow struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type DomainQuestionCount struct {
	DomainID string `json:"domain_id"`
	Count    int64  `json:"count"`
}

type KBGapRow struct {
	OccurredAt      string `json:"occurred_at"`
	DomainID        string `json:"domain_id"`
	Kind            string `json:"kind"`
	QuestionPreview string `json:"question_preview"`
}

type RAGStats struct {
	Total          int64   `json:"total"`
	VerifyPass     int64   `json:"verify_pass"`
	VerifyFail     int64   `json:"verify_fail"`
	SoftFail       int64   `json:"soft_fail"`
	VerifyPassRate float64 `json:"verify_pass_rate"`
}

type AnalyticsDashboard struct {
	Days            int                    `json:"days"`
	TenantID        string                 `json:"tenant_id"`
	QuestionsTotal  int64                  `json:"questions_total"`
	QuestionsToday  int64                  `json:"questions_today"`
	QuestionsPerDay []QuestionsPerDayRow   `json:"questions_per_day"`
	RAG             RAGStats               `json:"rag"`
	Feedback        []FeedbackSummaryRow   `json:"feedback"`
	TopDomains      []DomainQuestionCount  `json:"top_domains"`
	KBGaps          []KBGapRow             `json:"kb_gaps"`
}

func (st *ChatStore) AnalyticsDashboard(ctx context.Context, tenantID string, days int) (*AnalyticsDashboard, error) {
	if days < 1 {
		days = 7
	}
	if days > 90 {
		days = 90
	}
	out := &AnalyticsDashboard{
		Days:     days,
		TenantID: tenantID,
	}
	since := time.Now().UTC().AddDate(0, 0, -(days - 1))
	todayStart := time.Now().UTC().Truncate(24 * time.Hour)

	var err error
	out.QuestionsTotal, err = st.countUserQuestions(ctx, tenantID, since)
	if err != nil {
		return nil, err
	}
	out.QuestionsToday, err = st.countUserQuestions(ctx, tenantID, todayStart)
	if err != nil {
		return nil, err
	}
	out.QuestionsPerDay, err = st.questionsPerDay(ctx, tenantID, since)
	if err != nil {
		return nil, err
	}
	out.RAG, err = st.ragStats(ctx, tenantID, since)
	if err != nil {
		return nil, err
	}
	out.Feedback, err = st.FeedbackSummary(ctx)
	if err != nil {
		return nil, err
	}
	out.TopDomains, err = st.topDomainsByQuestions(ctx, tenantID, since)
	if err != nil {
		return nil, err
	}
	out.KBGaps, err = st.recentKBGaps(ctx, tenantID, since, 20)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (st *ChatStore) countUserQuestions(ctx context.Context, tenantID string, since time.Time) (int64, error) {
	var n int64
	err := st.pool.QueryRow(ctx, `
		SELECT COUNT(*)::bigint
		FROM messages m
		JOIN chat_sessions cs ON cs.id = m.session_id
		WHERE m.role = 'user'
		  AND m.created_at >= $1
		  AND ($2 = '' OR cs.tenant_id = $2)`,
		since, tenantID,
	).Scan(&n)
	return n, err
}

func (st *ChatStore) questionsPerDay(ctx context.Context, tenantID string, since time.Time) ([]QuestionsPerDayRow, error) {
	rows, err := st.pool.Query(ctx, `
		SELECT (m.created_at AT TIME ZONE 'UTC')::date AS day, COUNT(*)::bigint
		FROM messages m
		JOIN chat_sessions cs ON cs.id = m.session_id
		WHERE m.role = 'user'
		  AND m.created_at >= $1
		  AND ($2 = '' OR cs.tenant_id = $2)
		GROUP BY day
		ORDER BY day`,
		since, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []QuestionsPerDayRow
	for rows.Next() {
		var day time.Time
		var count int64
		if err := rows.Scan(&day, &count); err != nil {
			return nil, err
		}
		out = append(out, QuestionsPerDayRow{
			Date:  day.Format("2006-01-02"),
			Count: count,
		})
	}
	return out, rows.Err()
}

func (st *ChatStore) ragStats(ctx context.Context, tenantID string, since time.Time) (RAGStats, error) {
	var stats RAGStats
	err := st.pool.QueryRow(ctx, `
		SELECT
			COUNT(*)::bigint AS total,
			COUNT(*) FILTER (
				WHERE COALESCE((payload->>'soft_fail')::boolean, false) = false
				  AND COALESCE((payload->>'verify_pass')::boolean, false) = true
			)::bigint AS verify_pass,
			COUNT(*) FILTER (
				WHERE COALESCE((payload->>'soft_fail')::boolean, false) = false
				  AND COALESCE((payload->>'verify_pass')::boolean, false) = false
			)::bigint AS verify_fail,
			COUNT(*) FILTER (
				WHERE COALESCE((payload->>'soft_fail')::boolean, false) = true
			)::bigint AS soft_fail
		FROM analytics_events
		WHERE event_type = 'rag_answer'
		  AND created_at >= $1
		  AND ($2 = '' OR payload->>'tenant_id' = $2)`,
		since, tenantID,
	).Scan(&stats.Total, &stats.VerifyPass, &stats.VerifyFail, &stats.SoftFail)
	if err != nil {
		return stats, err
	}
	stats.VerifyPassRate = computeVerifyPassRate(stats.VerifyPass, stats.VerifyFail)
	return stats, nil
}

func (st *ChatStore) topDomainsByQuestions(ctx context.Context, tenantID string, since time.Time) ([]DomainQuestionCount, error) {
	rows, err := st.pool.Query(ctx, `
		SELECT cs.domain_id, COUNT(*)::bigint
		FROM messages m
		JOIN chat_sessions cs ON cs.id = m.session_id
		WHERE m.role = 'user'
		  AND m.created_at >= $1
		  AND ($2 = '' OR cs.tenant_id = $2)
		GROUP BY cs.domain_id
		ORDER BY COUNT(*) DESC
		LIMIT 10`,
		since, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []DomainQuestionCount
	for rows.Next() {
		var r DomainQuestionCount
		if err := rows.Scan(&r.DomainID, &r.Count); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (st *ChatStore) recentKBGaps(ctx context.Context, tenantID string, since time.Time, limit int) ([]KBGapRow, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := st.pool.Query(ctx, `
		SELECT created_at,
		       COALESCE(payload->>'domain_id', ''),
		       CASE
		           WHEN COALESCE((payload->>'soft_fail')::boolean, false) THEN 'soft_fail'
		           ELSE 'verify_fail'
		       END,
		       COALESCE(payload->>'question_preview', '')
		FROM analytics_events
		WHERE event_type = 'rag_answer'
		  AND created_at >= $1
		  AND ($2 = '' OR payload->>'tenant_id' = $2)
		  AND (
		      COALESCE((payload->>'soft_fail')::boolean, false) = true
		      OR COALESCE((payload->>'verify_pass')::boolean, false) = false
		  )
		ORDER BY created_at DESC
		LIMIT $3`,
		since, tenantID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []KBGapRow
	for rows.Next() {
		var at time.Time
		var r KBGapRow
		if err := rows.Scan(&at, &r.DomainID, &r.Kind, &r.QuestionPreview); err != nil {
			return nil, err
		}
		r.OccurredAt = at.UTC().Format(time.RFC3339)
		out = append(out, r)
	}
	return out, rows.Err()
}

func computeVerifyPassRate(pass, fail int64) float64 {
	denom := pass + fail
	if denom == 0 {
		return 0
	}
	return float64(pass) / float64(denom) * 100
}

func parseAnalyticsDays(s string) int {
	if s == "" {
		return 7
	}
	var d int
	if _, err := fmt.Sscanf(s, "%d", &d); err != nil || d < 1 {
		return 7
	}
	if d > 90 {
		return 90
	}
	return d
}
