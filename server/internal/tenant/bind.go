package tenant

import (
	"context"

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
