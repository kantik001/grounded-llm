package analytics

import (
	"grounded_llm_server/internal/store"
)

var lookupStore func() *store.ChatStore

// BindStore overrides where this package logs analytics events.
func BindStore(fn func() *store.ChatStore) {
	lookupStore = fn
}

func st() *store.ChatStore {
	if lookupStore == nil {
		return nil
	}
	return lookupStore()
}
