package main

import (
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const ctxKeyAdminActor = "admin_actor"

type auditOpts struct {
	Action   string
	Actor    string
	TenantID string
	DomainID string
	Resource string
	Success  bool
	Details  map[string]any
}

func adminActorFromContext(c *gin.Context) string {
	if v, ok := c.Get(ctxKeyAdminActor); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	user, _, ok := c.Request.BasicAuth()
	if ok {
		return user
	}
	return ""
}

func auditClientIP(c *gin.Context) string {
	if fwd := strings.TrimSpace(c.GetHeader("X-Forwarded-For")); fwd != "" {
		parts := strings.Split(fwd, ",")
		return strings.TrimSpace(parts[0])
	}
	if rip := strings.TrimSpace(c.GetHeader("X-Real-IP")); rip != "" {
		return rip
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	return strings.TrimSpace(c.Request.RemoteAddr)
}

func isAdminStatusCheck(c *gin.Context) bool {
	return c.Request.Method == http.MethodGet && strings.HasSuffix(c.Request.URL.Path, "/status")
}

func recordAdminAudit(c *gin.Context, opts auditOpts) {
	if chatStore == nil || strings.TrimSpace(opts.Action) == "" {
		return
	}
	actor := strings.TrimSpace(opts.Actor)
	if actor == "" {
		actor = adminActorFromContext(c)
	}
	rec := auditRecord{
		Action:    opts.Action,
		Actor:     actor,
		TenantID:  opts.TenantID,
		DomainID:  opts.DomainID,
		Resource:  opts.Resource,
		ClientIP:  auditClientIP(c),
		RequestID: ctxRequestID(c),
		Success:   opts.Success,
		Details:   opts.Details,
	}
	if err := chatStore.RecordAudit(c.Request.Context(), rec); err != nil {
		log.Printf("audit %s: %v", opts.Action, err)
	}
}

func parseAuditLogQuery(c *gin.Context) (limit, offset int, action string) {
	limit = auditLogDefaultLimit
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			limit = n
		}
	}
	if raw := strings.TrimSpace(c.Query("offset")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n >= 0 {
			offset = n
		}
	}
	return limit, offset, strings.TrimSpace(c.Query("action"))
}

// GET /admin/audit-log
func handleAdminAuditLog(c *gin.Context) {
	if chatStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "Store not ready"})
		return
	}
	limit, offset, action := parseAuditLogQuery(c)
	entries, err := chatStore.ListAuditLog(c.Request.Context(), limit, offset, action)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if entries == nil {
		entries = []AuditEntry{}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"entries": entries,
		"limit":   limit,
		"offset":  offset,
	})
}
