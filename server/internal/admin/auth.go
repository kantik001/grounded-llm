package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/audit"
	"grounded_llm_server/internal/oidc"
	"grounded_llm_server/internal/store"
)

// AuthEnabled reports whether any admin auth mechanism is configured.
func AuthEnabled() bool {
	if oidc.Configured() {
		return true
	}
	if c := cfg(); c != nil && c.AdminPassword != "" {
		return true
	}
	return len(userRegistry) > 0
}

// BasicAuthEnabled reports whether HTTP basic auth is available.
func BasicAuthEnabled() bool {
	if c := cfg(); c != nil && c.AdminPassword != "" {
		return true
	}
	return len(userRegistry) > 0
}

// AuthMiddleware validates OIDC session or HTTP basic credentials.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !AuthEnabled() {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"success": false,
				"error":   "Admin UI disabled: set OIDC, ADMIN_PASSWORD, or ADMIN_USERS_FILE",
			})
			return
		}

		if sess, ok := oidc.AdminSessionFromRequest(c); ok {
			c.Set(CtxKeyAdminActor, oidc.AdminActorLabel(sess))
			c.Set(CtxKeyAdminRoles, sess.Roles)
			c.Next()
			return
		}

		if BasicAuthEnabled() {
			user, pass, ok := c.Request.BasicAuth()
			if ok {
				roles, authed := authenticateUser(user, pass)
				if authed {
					c.Set(CtxKeyAdminActor, user)
					c.Set(CtxKeyAdminRoles, roles)
					if audit.IsAdminStatusCheck(c) {
						audit.Record(c, audit.Opts{
							Action:  store.AuditActionLogin,
							Actor:   user,
							Success: true,
							Details: map[string]any{"auth": "basic"},
						})
					}
					c.Next()
					return
				}
				audit.Record(c, audit.Opts{
					Action:  store.AuditActionLoginFailed,
					Actor:   user,
					Success: false,
					Details: map[string]any{"auth": "basic"},
				})
			} else if !oidc.Configured() {
				audit.Record(c, audit.Opts{Action: store.AuditActionLoginFailed, Success: false})
			}
		}

		if oidc.Configured() {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success":     false,
				"error":       "Authentication required",
				"sso_login":   "/api/admin/auth/login",
				"sso_enabled": true,
			})
			return
		}

		c.Header("WWW-Authenticate", `Basic realm="Grounded LLM Admin"`)
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}

// OIDCHost wires admin RBAC and audit into the OIDC package.
type OIDCHost struct{}

func (OIDCHost) NormalizeRoles(in []string) []string { return NormalizeRoles(in) }

func (OIDCHost) RecordAudit(c *gin.Context, opts oidc.AuditOpts) {
	audit.Record(c, audit.Opts{
		Action:  opts.Action,
		Actor:   opts.Actor,
		Success: opts.Success,
		Details: opts.Details,
	})
}

func (OIDCHost) BasicAuthEnabled() bool { return BasicAuthEnabled() }
