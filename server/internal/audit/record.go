package audit

import (
	"context"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/store"
)

const ctxKeyAdminActor = "admin_actor"

type Opts struct {
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

func clientIP(c *gin.Context) string {
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

// IsAdminStatusCheck reports GET …/status probes used for basic-auth login audit.
func IsAdminStatusCheck(c *gin.Context) bool {
	return c.Request.Method == http.MethodGet && strings.HasSuffix(c.Request.URL.Path, "/status")
}

// Record persists an admin audit event from a Gin request context.
func Record(c *gin.Context, opts Opts) {
	if st() == nil || strings.TrimSpace(opts.Action) == "" {
		return
	}
	actor := strings.TrimSpace(opts.Actor)
	if actor == "" {
		actor = adminActorFromContext(c)
	}
	rec := store.AuditRecord{
		Action:    opts.Action,
		Actor:     actor,
		TenantID:  opts.TenantID,
		DomainID:  opts.DomainID,
		Resource:  opts.Resource,
		ClientIP:  clientIP(c),
		RequestID: ctxRequestID(c),
		Success:   opts.Success,
		Details:   opts.Details,
	}
	if err := st().RecordAudit(c.Request.Context(), rec); err != nil {
		log.Printf("audit %s: %v", opts.Action, err)
	}
}

// RecordBackground persists an audit event without a Gin context (e.g. async workers).
func RecordBackground(rec store.AuditRecord) {
	if st() == nil || strings.TrimSpace(rec.Action) == "" {
		return
	}
	if err := st().RecordAudit(context.Background(), rec); err != nil {
		log.Printf("audit %s: %v", rec.Action, err)
	}
}

// ParseLogQuery reads limit/offset/action from GET /admin/audit-log query params.
func ParseLogQuery(c *gin.Context) (limit, offset int, action string) {
	limit = store.AuditLogDefaultLimit
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

// ActorFromContext returns the admin actor label from Gin context.
func ActorFromContext(c *gin.Context) string {
	return adminActorFromContext(c)
}

// ClientIP returns the best-effort client IP for audit logging.
func ClientIP(c *gin.Context) string {
	return clientIP(c)
}
