package audit

import (
	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/store"
)

var lookupStore func() *store.ChatStore
var requestIDFrom func(*gin.Context) string

// BindStore overrides where this package persists audit events.
func BindStore(fn func() *store.ChatStore) {
	lookupStore = fn
}

// BindRequestID wires request-id extraction from the HTTP layer.
func BindRequestID(fn func(*gin.Context) string) {
	requestIDFrom = fn
}

func st() *store.ChatStore {
	if lookupStore == nil {
		return nil
	}
	return lookupStore()
}

func ctxRequestID(c *gin.Context) string {
	if requestIDFrom != nil {
		return requestIDFrom(c)
	}
	return ""
}
