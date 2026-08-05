package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"

	"github.com/gin-gonic/gin"
)

const CtxKeyRequestID = "request_id"

// RequestID sets/propagates X-Request-ID.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			b := make([]byte, 8)
			_, _ = rand.Read(b)
			id = hex.EncodeToString(b)
		}
		c.Set(CtxKeyRequestID, id)
		c.Writer.Header().Set("X-Request-ID", id)
		c.Next()
	}
}

// CtxRequestID returns the request id from gin context.
func CtxRequestID(c *gin.Context) string {
	if v, ok := c.Get(CtxKeyRequestID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// FormatLogValue formats a field for structured request logs.
func FormatLogValue(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(x)
	default:
		return "?"
	}
}
