package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

const ctxKeyRequestID = "request_id"

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			b := make([]byte, 8)
			_, _ = rand.Read(b)
			id = hex.EncodeToString(b)
		}
		c.Set(ctxKeyRequestID, id)
		c.Writer.Header().Set("X-Request-ID", id)
		c.Next()
	}
}

func ctxRequestID(c *gin.Context) string {
	if v, ok := c.Get(ctxKeyRequestID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func logRequest(c *gin.Context, event string, fields map[string]any) {
	msg := "req=" + ctxRequestID(c) + " event=" + event
	if tid := ctxTenantID(c); tid != "" {
		msg += " tenant=" + tid
	}
	for k, v := range fields {
		msg += " " + k + "=" + formatLogValue(v)
	}
	log.Print(msg)
}

func formatLogValue(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case bool:
		return strconv.FormatBool(x)
	default:
		return "?"
	}
}
