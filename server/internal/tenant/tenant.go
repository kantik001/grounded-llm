package tenant

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/config"
)

// CtxKeyTenantID is the gin context key for the resolved tenant id.
const CtxKeyTenantID = "tenant_id"

// CtxKeyTenantExplicit marks that the client explicitly requested a tenant (header/query).
const CtxKeyTenantExplicit = "tenant_id_explicit"

// CtxKeyBoundTenant is the tenant the presented credentials are bound to ("*" = any).
const CtxKeyBoundTenant = "tenant_id_bound"

var allowedTenants map[string]struct{}

// InitAllowlist builds the tenant allowlist from config and ALLOWED_TENANTS.
func InitAllowlist(cfg *config.Config) {
	allowedTenants = make(map[string]struct{})
	raw := strings.TrimSpace(os.Getenv("ALLOWED_TENANTS"))
	if raw == "" {
		allowedTenants[cfg.DefaultTenantID] = struct{}{}
		return
	}
	for _, part := range strings.Split(raw, ",") {
		id := NormalizeTenantID(part)
		if id != "" {
			allowedTenants[id] = struct{}{}
		}
	}
	if len(allowedTenants) == 0 {
		allowedTenants[cfg.DefaultTenantID] = struct{}{}
	}
}

// NormalizeTenantID lowercases and trims a tenant id.
func NormalizeTenantID(raw string) string {
	return strings.TrimSpace(strings.ToLower(raw))
}

// AllowTenant adds a tenant id to the in-memory allowlist.
func AllowTenant(id string) {
	if allowedTenants == nil {
		allowedTenants = make(map[string]struct{})
	}
	id = NormalizeTenantID(id)
	if id != "" {
		allowedTenants[id] = struct{}{}
	}
}

// IsAllowed reports whether tenantID is on the allowlist (always true when allowlist is empty).
func IsAllowed(tenantID string) bool {
	id := NormalizeTenantID(tenantID)
	if len(allowedTenants) == 0 {
		return true
	}
	_, ok := allowedTenants[id]
	return ok
}

// OnAllowlist reports whether tenantID is explicitly present on the allowlist map.
func OnAllowlist(tenantID string) bool {
	if allowedTenants == nil {
		return false
	}
	_, ok := allowedTenants[NormalizeTenantID(tenantID)]
	return ok
}

// ResolveID resolves tenant id from context, headers, query, or config default.
func ResolveID(c *gin.Context, cfg *config.Config) (string, error) {
	if v, ok := c.Get(CtxKeyTenantID); ok {
		if s, ok := v.(string); ok && s != "" {
			return s, nil
		}
	}
	raw := strings.TrimSpace(c.GetHeader("X-Tenant-ID"))
	if raw == "" {
		raw = strings.TrimSpace(c.Query("tenant_id"))
	}
	explicit := raw != ""
	if raw == "" {
		raw = cfg.DefaultTenantID
	}
	id := NormalizeTenantID(raw)
	if id == "" {
		id = cfg.DefaultTenantID
	}
	if len(allowedTenants) > 0 {
		if _, ok := allowedTenants[id]; !ok {
			return "", fmt.Errorf("unknown tenant: %s", raw)
		}
	}
	c.Set(CtxKeyTenantID, id)
	if explicit {
		c.Set(CtxKeyTenantExplicit, true)
	}
	if members, ok := MemberTenants(c); ok && !memberContains(members, id) {
		return "", fmt.Errorf("credentials are not authorized for tenant %q", id)
	}
	return id, nil
}

// ExplicitID returns the resolved tenant and whether the client requested it explicitly.
func ExplicitID(c *gin.Context) (string, bool) {
	explicit := false
	if v, ok := c.Get(CtxKeyTenantExplicit); ok {
		explicit, _ = v.(bool)
	}
	return CtxID(c), explicit
}

// BoundTenant returns the tenant the request credentials are bound to, if any.
func BoundTenant(c *gin.Context) (string, bool) {
	if v, ok := c.Get(CtxKeyBoundTenant); ok {
		if s, ok := v.(string); ok && s != "" {
			return s, true
		}
	}
	return "", false
}

// BindCredentialTenant pins the request to the tenant its credentials belong to.
// A bound tenant of "*" permits any allowlisted tenant (integration gateways).
// Returns an error when the client explicitly requested a different tenant.
func BindCredentialTenant(c *gin.Context, boundTenant string) error {
	bound := NormalizeTenantID(boundTenant)
	if bound == "" || bound == "*" {
		if bound == "*" {
			c.Set(CtxKeyBoundTenant, "*")
		}
		return nil
	}
	if requested, explicit := ExplicitID(c); explicit && requested != bound {
		return fmt.Errorf("credentials are not authorized for tenant %q", requested)
	}
	if len(allowedTenants) > 0 {
		if _, ok := allowedTenants[bound]; !ok {
			return fmt.Errorf("credential tenant %q is not on the allowlist", bound)
		}
	}
	c.Set(CtxKeyTenantID, bound)
	c.Set(CtxKeyBoundTenant, bound)
	return nil
}

// RequestOverride is defined in membership.go (also enforces Telegram memberships).

// Middleware resolves tenant id on each request.
func Middleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, err := ResolveID(c, cfg); err != nil {
			c.AbortWithStatusJSON(400, gin.H{"success": false, "error": err.Error()})
			return
		}
		c.Next()
	}
}

// CtxID returns the tenant id stored on the gin context, or the process default.
func CtxID(c *gin.Context) string {
	if v, ok := c.Get(CtxKeyTenantID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	if c := cfg(); c != nil {
		return c.DefaultTenantID
	}
	return "default"
}

// KBDataDir returns the knowledge-base directory for tenant and domain.
func KBDataDir(tenantID, domainID string) string {
	c := cfg()
	tenantID = NormalizeTenantID(tenantID)
	if tenantID == "" {
		tenantID = "default"
	}
	nested := filepath.Join(c.DataDir, tenantID, domainID)
	if tenantID == c.DefaultTenantID {
		legacy := filepath.Join(c.DataDir, domainID)
		if hasKnowledgeFiles(legacy) {
			return legacy
		}
	}
	return nested
}

// AdminID reads optional tenant_id query param for admin routes.
func AdminID(c *gin.Context) string {
	raw := strings.TrimSpace(c.Query("tenant_id"))
	if raw == "" {
		return cfg().DefaultTenantID
	}
	return NormalizeTenantID(raw)
}

func hasKnowledgeFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && isKnowledgeFile(e.Name()) {
			return true
		}
	}
	return false
}

func isKnowledgeFile(name string) bool {
	n := strings.ToLower(name)
	return strings.HasSuffix(n, ".txt") || strings.HasSuffix(n, ".pdf") || strings.HasSuffix(n, ".docx")
}
