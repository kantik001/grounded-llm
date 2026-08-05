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
	return id, nil
}

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
