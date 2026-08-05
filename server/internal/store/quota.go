package store

import (
	"context"
	"fmt"
)

// CountTenantUserMessagesToday counts user-role messages since UTC midnight.
func (st *ChatStore) CountTenantUserMessagesToday(ctx context.Context, tenantID string) (int64, error) {
	if st == nil || st.Pool == nil {
		return 0, fmt.Errorf("chat store not initialized")
	}
	var n int64
	err := st.Pool.QueryRow(ctx, `
		SELECT COUNT(*)::bigint FROM messages m
		JOIN chat_sessions cs ON cs.id = m.session_id
		WHERE cs.tenant_id = $1 AND m.role = 'user'
		  AND m.created_at >= date_trunc('day', NOW() AT TIME ZONE 'UTC')`,
		NormalizeTenantID(tenantID),
	).Scan(&n)
	return n, err
}
