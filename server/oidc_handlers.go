package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func adminAuthEnabled(cfg *Config) bool {
	if oidcConfigured() {
		return true
	}
	if cfg.AdminPassword != "" {
		return true
	}
	return len(adminUserRegistry) > 0
}

func adminBasicAuthEnabled() bool {
	if config != nil && config.AdminPassword != "" {
		return true
	}
	return len(adminUserRegistry) > 0
}

func adminAuthMiddleware(cfg *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !adminAuthEnabled(cfg) {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"success": false,
				"error":   "Admin UI disabled: set OIDC, ADMIN_PASSWORD, or ADMIN_USERS_FILE",
			})
			return
		}

		if sess, ok := adminSessionFromRequest(c); ok {
			c.Set(ctxKeyAdminActor, adminActorLabel(sess))
			c.Set(ctxKeyAdminRoles, sess.Roles)
			c.Next()
			return
		}

		if adminBasicAuthEnabled() {
			user, pass, ok := c.Request.BasicAuth()
			if ok {
				roles, authed := authenticateAdminUser(user, pass)
				if authed {
					c.Set(ctxKeyAdminActor, user)
					c.Set(ctxKeyAdminRoles, roles)
					if isAdminStatusCheck(c) {
						recordAdminAudit(c, auditOpts{
							Action:  auditActionLogin,
							Actor:   user,
							Success: true,
							Details: map[string]any{"auth": "basic"},
						})
					}
					c.Next()
					return
				}
				recordAdminAudit(c, auditOpts{
					Action:  auditActionLoginFailed,
					Actor:   user,
					Success: false,
					Details: map[string]any{"auth": "basic"},
				})
			} else if !oidcConfigured() {
				recordAdminAudit(c, auditOpts{Action: auditActionLoginFailed, Success: false})
			}
		}

		if oidcConfigured() {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success":    false,
				"error":      "Authentication required",
				"sso_login":  "/api/admin/auth/login",
				"sso_enabled": true,
			})
			return
		}

		c.Header("WWW-Authenticate", `Basic realm="Grounded LLM Admin"`)
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}

func registerOIDCAuthRoutes(g *gin.RouterGroup) {
	g.GET("/config", handleOIDCAuthConfig)
	g.GET("/login", handleOIDCLogin)
	g.GET("/callback", handleOIDCCallback)
	g.POST("/logout", handleOIDCLogout)
}

func handleOIDCAuthConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"sso_enabled": oidcConfigured(),
		"login_path":  "/api/admin/auth/login",
		"logout_path": "/api/admin/auth/logout",
		"basic_auth":  adminBasicAuthEnabled(),
	})
}

func handleOIDCLogin(c *gin.Context) {
	if !oidcConfigured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "OIDC SSO is not enabled"})
		return
	}
	_, oauth2Config, _, err := ensureOIDCProvider(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": err.Error()})
		return
	}
	state := newOAuthState()
	setOAuthStateCookie(c, state)
	c.Redirect(http.StatusFound, oauth2Config.AuthCodeURL(state))
}

func handleOIDCCallback(c *gin.Context) {
	if !oidcConfigured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "OIDC SSO is not enabled"})
		return
	}
	if errMsg := strings.TrimSpace(c.Query("error")); errMsg != "" {
		desc := strings.TrimSpace(c.Query("error_description"))
		recordAdminAudit(c, auditOpts{
			Action:  auditActionLoginFailed,
			Success: false,
			Details: map[string]any{"oidc_error": errMsg, "description": desc},
		})
		c.Redirect(http.StatusFound, "/admin.html?sso_error="+errMsg)
		return
	}
	state, ok := popOAuthStateCookie(c)
	if !ok || state != strings.TrimSpace(c.Query("state")) {
		recordAdminAudit(c, auditOpts{Action: auditActionLoginFailed, Success: false, Details: map[string]any{"reason": "invalid_oauth_state"}})
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid OAuth state"})
		return
	}
	code := strings.TrimSpace(c.Query("code"))
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Missing authorization code"})
		return
	}
	_, oauth2Config, verifier, err := ensureOIDCProvider(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": err.Error()})
		return
	}
	token, err := oauth2Config.Exchange(c.Request.Context(), code)
	if err != nil {
		recordAdminAudit(c, auditOpts{Action: auditActionLoginFailed, Success: false, Details: map[string]any{"reason": "token_exchange", "error": err.Error()}})
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": "OIDC token exchange failed"})
		return
	}
	rawID, ok := token.Extra("id_token").(string)
	if !ok || rawID == "" {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": "OIDC id_token missing"})
		return
	}
	idToken, err := verifier.Verify(c.Request.Context(), rawID)
	if err != nil {
		recordAdminAudit(c, auditOpts{Action: auditActionLoginFailed, Success: false, Details: map[string]any{"reason": "id_token_verify"}})
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Invalid ID token"})
		return
	}
	var claims map[string]any
	if err := idToken.Claims(&claims); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": "Failed to parse ID token claims"})
		return
	}
	email, _ := claims["email"].(string)
	name, _ := claims["name"].(string)
	roles := resolveOIDCRoles(email, claims)
	if len(roles) == 0 {
		recordAdminAudit(c, auditOpts{Action: auditActionLoginFailed, Actor: email, Success: false, Details: map[string]any{"reason": "no_roles"}})
		c.Redirect(http.StatusFound, "/admin.html?sso_error=forbidden")
		return
	}
	payload := adminSessionPayload{
		Subject: idToken.Subject,
		Email:   email,
		Name:    name,
		Roles:   roles,
		Exp:     timeNow().Add(oidcCfg.sessionTTL()).Unix(),
		Auth:    adminSessionAuthOIDC,
	}
	if err := setAdminSessionCookie(c, payload); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	recordAdminAudit(c, auditOpts{
		Action:  auditActionLogin,
		Actor:   adminActorLabel(payload),
		Success: true,
		Details: map[string]any{"auth": "oidc", "roles": roles},
	})
	c.Redirect(http.StatusFound, "/admin.html")
}

func handleOIDCLogout(c *gin.Context) {
	if sess, ok := adminSessionFromRequest(c); ok {
		recordAdminAudit(c, auditOpts{
			Action:  auditActionLogout,
			Actor:   adminActorLabel(sess),
			Success: true,
			Details: map[string]any{"auth": sess.Auth},
		})
	}
	clearAdminSessionCookie(c)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// timeNow is overridden in tests.
var timeNow = func() time.Time { return time.Now() }
