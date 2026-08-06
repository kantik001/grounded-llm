package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// SaaSTenant is a row in saas_tenants.
type SaaSTenant struct {
	TenantID         string
	OrgName          string
	Email            string
	Plan             string
	StripeCustomerID string
	CreatedAt        time.Time
}

// TenantQuotaLimits mirrors tenant.QuotaLimits for persistence.
type TenantQuotaLimits struct {
	MessagesPerDay int
	StorageMB      int
	MaxDomains     int
}

// ClaimStripeEvent inserts event_id; returns true if this process should handle it
// (first claim), false if the event was already processed (idempotent replay).
func (st *ChatStore) ClaimStripeEvent(ctx context.Context, eventID, eventType string) (bool, error) {
	if st == nil || st.Pool == nil {
		return true, nil // no DB → caller proceeds without persistence
	}
	eventID = strings.TrimSpace(eventID)
	if eventID == "" {
		return true, nil
	}
	_, err := st.Pool.Exec(ctx, `
		INSERT INTO stripe_webhook_events (event_id, event_type)
		VALUES ($1, $2)`, eventID, eventType)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return false, nil
		}
		return false, fmt.Errorf("claim stripe event: %w", err)
	}
	return true, nil
}

// UpsertSaaSTenant inserts or updates a SaaS tenant row.
func (st *ChatStore) UpsertSaaSTenant(ctx context.Context, t SaaSTenant) error {
	if st == nil || st.Pool == nil {
		return fmt.Errorf("store not configured")
	}
	id := trimID(t.TenantID)
	if id == "" {
		return fmt.Errorf("empty tenant_id")
	}
	created := t.CreatedAt
	if created.IsZero() {
		created = time.Now().UTC()
	}
	_, err := st.Pool.Exec(ctx, `
		INSERT INTO saas_tenants (tenant_id, org_name, email, plan, stripe_customer_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (tenant_id) DO UPDATE SET
			org_name = EXCLUDED.org_name,
			email = EXCLUDED.email,
			plan = EXCLUDED.plan,
			stripe_customer_id = CASE
				WHEN EXCLUDED.stripe_customer_id <> '' THEN EXCLUDED.stripe_customer_id
				ELSE saas_tenants.stripe_customer_id
			END,
			updated_at = NOW()`,
		id, t.OrgName, t.Email, t.Plan, t.StripeCustomerID, created)
	return err
}

// UpdateSaaSTenantPlan updates plan for an existing tenant.
func (st *ChatStore) UpdateSaaSTenantPlan(ctx context.Context, tenantID, plan string) error {
	if st == nil || st.Pool == nil {
		return fmt.Errorf("store not configured")
	}
	tag, err := st.Pool.Exec(ctx, `
		UPDATE saas_tenants SET plan = $2, updated_at = NOW() WHERE tenant_id = $1`,
		trimID(tenantID), plan)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("tenant not found: %s", tenantID)
	}
	return nil
}

// UpdateSaaSTenantStripeCustomer stores the Stripe customer id.
func (st *ChatStore) UpdateSaaSTenantStripeCustomer(ctx context.Context, tenantID, customerID string) error {
	if st == nil || st.Pool == nil {
		return fmt.Errorf("store not configured")
	}
	tag, err := st.Pool.Exec(ctx, `
		UPDATE saas_tenants SET stripe_customer_id = $2, updated_at = NOW() WHERE tenant_id = $1`,
		trimID(tenantID), customerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("tenant not found: %s", tenantID)
	}
	return nil
}

// GetSaaSTenantEmail returns the signup email for a tenant.
func (st *ChatStore) GetSaaSTenantEmail(ctx context.Context, tenantID string) (string, error) {
	if st == nil || st.Pool == nil {
		return "", nil
	}
	var email string
	err := st.Pool.QueryRow(ctx,
		`SELECT email FROM saas_tenants WHERE tenant_id = $1`, trimID(tenantID),
	).Scan(&email)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return email, err
}

// ListSaaSTenants returns all SaaS tenants.
func (st *ChatStore) ListSaaSTenants(ctx context.Context) ([]SaaSTenant, error) {
	if st == nil || st.Pool == nil {
		return nil, nil
	}
	rows, err := st.Pool.Query(ctx, `
		SELECT tenant_id, org_name, email, plan, COALESCE(stripe_customer_id, ''), created_at
		FROM saas_tenants ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []SaaSTenant
	for rows.Next() {
		var t SaaSTenant
		if err := rows.Scan(&t.TenantID, &t.OrgName, &t.Email, &t.Plan, &t.StripeCustomerID, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// UpsertTenantQuota writes quota limits for a tenant (tenant row must exist).
func (st *ChatStore) UpsertTenantQuota(ctx context.Context, tenantID string, lim TenantQuotaLimits) error {
	if st == nil || st.Pool == nil {
		return fmt.Errorf("store not configured")
	}
	_, err := st.Pool.Exec(ctx, `
		INSERT INTO tenant_quotas (tenant_id, messages_per_day, storage_mb, max_domains, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (tenant_id) DO UPDATE SET
			messages_per_day = EXCLUDED.messages_per_day,
			storage_mb = EXCLUDED.storage_mb,
			max_domains = EXCLUDED.max_domains,
			updated_at = NOW()`,
		trimID(tenantID), lim.MessagesPerDay, lim.StorageMB, lim.MaxDomains)
	return err
}

// ListTenantQuotas returns all persisted quota rows.
func (st *ChatStore) ListTenantQuotas(ctx context.Context) (map[string]TenantQuotaLimits, error) {
	out := make(map[string]TenantQuotaLimits)
	if st == nil || st.Pool == nil {
		return out, nil
	}
	rows, err := st.Pool.Query(ctx, `
		SELECT tenant_id, messages_per_day, storage_mb, max_domains FROM tenant_quotas`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var lim TenantQuotaLimits
		if err := rows.Scan(&id, &lim.MessagesPerDay, &lim.StorageMB, &lim.MaxDomains); err != nil {
			return nil, err
		}
		out[id] = lim
	}
	return out, rows.Err()
}

func trimID(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}
