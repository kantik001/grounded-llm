package store

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// AdminUser is a row in admin_users.
type AdminUser struct {
	Username       string
	PasswordBcrypt string
	Roles          []string
	TenantID       string
	CreatedAt      time.Time
}

// UpsertAdminUser inserts or updates an admin user.
func (st *ChatStore) UpsertAdminUser(ctx context.Context, u AdminUser) error {
	if st == nil || st.Pool == nil {
		return fmt.Errorf("store not configured")
	}
	username := strings.TrimSpace(u.Username)
	if username == "" {
		return fmt.Errorf("empty username")
	}
	roles := u.Roles
	if roles == nil {
		roles = []string{}
	}
	_, err := st.Pool.Exec(ctx, `
		INSERT INTO admin_users (username, password_bcrypt, roles, tenant_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (username) DO UPDATE SET
			password_bcrypt = EXCLUDED.password_bcrypt,
			roles = EXCLUDED.roles,
			tenant_id = CASE
				WHEN EXCLUDED.tenant_id <> '' THEN EXCLUDED.tenant_id
				ELSE admin_users.tenant_id
			END,
			updated_at = NOW()`,
		username, u.PasswordBcrypt, roles, strings.TrimSpace(u.TenantID))
	return err
}

// ListAdminUsers returns all admin users.
func (st *ChatStore) ListAdminUsers(ctx context.Context) ([]AdminUser, error) {
	if st == nil || st.Pool == nil {
		return nil, nil
	}
	rows, err := st.Pool.Query(ctx, `
		SELECT username, password_bcrypt, roles, COALESCE(tenant_id, ''), created_at
		FROM admin_users ORDER BY username`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AdminUser
	for rows.Next() {
		var u AdminUser
		if err := rows.Scan(&u.Username, &u.PasswordBcrypt, &u.Roles, &u.TenantID, &u.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// AdminUserExists reports whether username is present.
func (st *ChatStore) AdminUserExists(ctx context.Context, username string) (bool, error) {
	if st == nil || st.Pool == nil {
		return false, nil
	}
	var n int
	err := st.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM admin_users WHERE username = $1`, strings.TrimSpace(username),
	).Scan(&n)
	return n > 0, err
}

// UpsertUserTenantMembership grants a Telegram user access to a tenant.
func (st *ChatStore) UpsertUserTenantMembership(ctx context.Context, telegramID int64, tenantID, role string) error {
	if st == nil || st.Pool == nil {
		return fmt.Errorf("store not configured")
	}
	tenantID = strings.TrimSpace(strings.ToLower(tenantID))
	if telegramID == 0 || tenantID == "" {
		return fmt.Errorf("telegram_id and tenant_id required")
	}
	if role == "" {
		role = "member"
	}
	_, err := st.Pool.Exec(ctx, `
		INSERT INTO user_tenant_memberships (telegram_id, tenant_id, role, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (telegram_id, tenant_id) DO UPDATE SET role = EXCLUDED.role`,
		telegramID, tenantID, role)
	return err
}

// ListUserTenantMemberships returns tenant ids for a Telegram user.
func (st *ChatStore) ListUserTenantMemberships(ctx context.Context, telegramID int64) ([]string, error) {
	if st == nil || st.Pool == nil {
		return nil, nil
	}
	rows, err := st.Pool.Query(ctx, `
		SELECT tenant_id FROM user_tenant_memberships
		WHERE telegram_id = $1 ORDER BY tenant_id`, telegramID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// UserHasTenantMembership reports whether the user may access tenantID.
func (st *ChatStore) UserHasTenantMembership(ctx context.Context, telegramID int64, tenantID string) (bool, error) {
	if st == nil || st.Pool == nil {
		return true, nil
	}
	var n int
	err := st.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM user_tenant_memberships
		WHERE telegram_id = $1 AND tenant_id = $2`,
		telegramID, strings.TrimSpace(strings.ToLower(tenantID)),
	).Scan(&n)
	return n > 0, err
}
