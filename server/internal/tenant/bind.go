package tenant

import (
	"context"
	"os"
	"strings"

	"grounded_llm_server/internal/config"
	"grounded_llm_server/internal/store"
)

var lookupConfig = config.Current

// BindConfig overrides where this package reads process config (app sets app.config).
func BindConfig(fn func() *config.Config) {
	if fn != nil {
		lookupConfig = fn
	}
}

func cfg() *config.Config {
	return lookupConfig()
}

type messageCounter interface {
	CountTenantUserMessagesToday(ctx context.Context, tenantID string) (int64, error)
}

var storeGetter func() *store.ChatStore

// BindStore wires chat persistence for quota usage checks.
func BindStore(fn func() *store.ChatStore) {
	storeGetter = fn
}

func chatStore() messageCounter {
	if storeGetter != nil {
		return storeGetter()
	}
	return nil
}

func saasStore() *store.ChatStore {
	if storeGetter != nil {
		return storeGetter()
	}
	return nil
}

// UsePostgresBackend reports whether SaaS tenant rows should prefer Postgres.
// Default on when a ChatStore is wired; set TENANTS_STORE=file to force JSON-only.
func UsePostgresBackend() bool {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("TENANTS_STORE")))
	if mode == "file" || mode == "json" {
		return false
	}
	return saasStore() != nil
}
