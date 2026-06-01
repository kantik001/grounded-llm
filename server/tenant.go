package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

const ctxKeyTenantID = "tenant_id"

var allowedTenants map[string]struct{}

func initTenantConfig(cfg *Config) {
	allowedTenants = make(map[string]struct{})
	raw := strings.TrimSpace(os.Getenv("ALLOWED_TENANTS"))
	if raw == "" {
		allowedTenants[cfg.DefaultTenantID] = struct{}{}
		return
	}
	for _, part := range strings.Split(raw, ",") {
		id := normalizeTenantID(part)
		if id != "" {
			allowedTenants[id] = struct{}{}
		}
	}
	if len(allowedTenants) == 0 {
		allowedTenants[cfg.DefaultTenantID] = struct{}{}
	}
}

func normalizeTenantID(raw string) string {
	return strings.TrimSpace(strings.ToLower(raw))
}

func resolveTenantID(c *gin.Context, cfg *Config) (string, error) {
	if v, ok := c.Get(ctxKeyTenantID); ok {
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
	id := normalizeTenantID(raw)
	if id == "" {
		id = cfg.DefaultTenantID
	}
	if len(allowedTenants) > 0 {
		if _, ok := allowedTenants[id]; !ok {
			return "", fmt.Errorf("unknown tenant: %s", raw)
		}
	}
	c.Set(ctxKeyTenantID, id)
	return id, nil
}

func tenantMiddleware(cfg *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, err := resolveTenantID(c, cfg); err != nil {
			c.AbortWithStatusJSON(400, gin.H{"success": false, "error": err.Error()})
			return
		}
		c.Next()
	}
}

func ctxTenantID(c *gin.Context) string {
	if v, ok := c.Get(ctxKeyTenantID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	if config != nil {
		return config.DefaultTenantID
	}
	return "default"
}

func kbDataDir(tenantID, domainID string) string {
	tenantID = normalizeTenantID(tenantID)
	if tenantID == "" {
		tenantID = "default"
	}
	nested := filepath.Join(config.DataDir, tenantID, domainID)
	if tenantID == config.DefaultTenantID {
		legacy := filepath.Join(config.DataDir, domainID)
		if hasKnowledgeFiles(legacy) {
			return legacy
		}
	}
	return nested
}

func adminTenantID(c *gin.Context) string {
	raw := strings.TrimSpace(c.Query("tenant_id"))
	if raw == "" {
		return config.DefaultTenantID
	}
	return normalizeTenantID(raw)
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
