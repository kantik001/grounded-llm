package httpapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ParseAllowedOrigins splits CORS_ALLOWED_ORIGINS (comma-separated).
func ParseAllowedOrigins(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		p := strings.TrimSpace(part)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// CORS allows listed origins. An empty list allows no cross-origin callers
// (same-origin only); an explicit "*" reflects any Origin without credentials.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowAll := false
	originSet := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		if o == "*" {
			allowAll = true
			continue
		}
		originSet[o] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			if _, ok := originSet[origin]; ok {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			} else if allowAll {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			}
			c.Writer.Header().Set("Vary", "Origin")
		}
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Requested-With, X-Telegram-Init-Data, Authorization, X-API-Key, X-Tenant-ID, X-Locale, X-Request-ID, Accept-Language")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "X-Cache, X-Request-ID")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
