package oidc

import (
	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/config"
)

var lookupConfig = config.Current

// BindConfig overrides where this package reads process config (app sets app.config).
func BindConfig(fn func() *config.Config) {
	if fn != nil {
		lookupConfig = fn
	}
}

func cfg() *config.Config {
	return lookupConfig()
}

// AuditOpts is forwarded to the app audit recorder.
type AuditOpts struct {
	Action  string
	Actor   string
	Success bool
	Details map[string]any
}

// Host wires admin auth hooks owned by app until those packages extract.
type Host interface {
	NormalizeRoles(in []string) []string
	RecordAudit(c *gin.Context, opts AuditOpts)
	BasicAuthEnabled() bool
}

var host Host

// BindHost wires session audit and RBAC helpers from the app layer.
func BindHost(h Host) {
	host = h
}

func normalizeRoles(in []string) []string {
	if host != nil {
		return host.NormalizeRoles(in)
	}
	return in
}

func recordAudit(c *gin.Context, opts AuditOpts) {
	if host != nil {
		host.RecordAudit(c, opts)
	}
}

func basicAuthEnabled() bool {
	if host != nil {
		return host.BasicAuthEnabled()
	}
	return false
}
