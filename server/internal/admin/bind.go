package admin

import (
	cfgpkg "grounded_llm_server/internal/config"
	"grounded_llm_server/internal/store"
)

var lookupConfig = cfgpkg.Current
var lookupStore func() *store.ChatStore

// BindConfig overrides where this package reads process config (app sets app.config).
func BindConfig(fn func() *cfgpkg.Config) {
	if fn != nil {
		lookupConfig = fn
	}
}

// BindStore overrides where this package reads the chat store (app sets app.chatStore).
func BindStore(fn func() *store.ChatStore) {
	lookupStore = fn
}

func cfg() *cfgpkg.Config {
	return lookupConfig()
}

func st() *store.ChatStore {
	if lookupStore == nil {
		return nil
	}
	return lookupStore()
}
