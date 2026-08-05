package httpapi

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// DefaultJSONContentType sets application/json unless the route is media or SSE.
func DefaultJSONContentType() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.Request.URL.Path, "/media/") {
			c.Next()
			return
		}
		if c.Query("stream") == "1" || c.Query("stream") == "true" {
			c.Next()
			return
		}
		if strings.Contains(c.GetHeader("Accept"), "text/event-stream") {
			c.Next()
			return
		}
		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		c.Next()
	}
}
