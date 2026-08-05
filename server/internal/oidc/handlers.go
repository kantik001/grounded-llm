package oidc

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/store"
)

const (
	auditActionLogin       = store.AuditActionLogin
	auditActionLoginFailed = store.AuditActionLoginFailed
	auditActionLogout      = store.AuditActionLogout
)

// RegisterAuthRoutes mounts OIDC SSO login/callback/logout routes.
func RegisterAuthRoutes(g *gin.RouterGroup) {
	g.GET("/config", handleAuthConfig)
	g.GET("/login", handleLogin)
	g.GET("/callback", handleCallback)
	g.POST("/logout", handleLogout)
}

func handleAuthConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"sso_enabled": Configured(),
		"login_path":  "/api/admin/auth/login",
		"logout_path": "/api/admin/auth/logout",
		"basic_auth":  basicAuthEnabled(),
	})
}

func handleLogin(c *gin.Context) {
	if !Configured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "OIDC SSO is not enabled"})
		return
	}
	_, oauth2Config, _, err := ensureProvider(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": err.Error()})
		return
	}
	state := newOAuthState()
	setOAuthStateCookie(c, state)
	c.Redirect(http.StatusFound, oauth2Config.AuthCodeURL(state))
}

func handleCallback(c *gin.Context) {
	if !Configured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "OIDC SSO is not enabled"})
		return
	}
	if errMsg := strings.TrimSpace(c.Query("error")); errMsg != "" {
		desc := strings.TrimSpace(c.Query("error_description"))
		recordAudit(c, AuditOpts{
			Action:  auditActionLoginFailed,
			Success: false,
			Details: map[string]any{"oidc_error": errMsg, "description": desc},
		})
		c.Redirect(http.StatusFound, "/admin.html?sso_error="+errMsg)
		return
	}
	state, ok := popOAuthStateCookie(c)
	if !ok || state != strings.TrimSpace(c.Query("state")) {
		recordAudit(c, AuditOpts{Action: auditActionLoginFailed, Success: false, Details: map[string]any{"reason": "invalid_oauth_state"}})
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid OAuth state"})
		return
	}
	code := strings.TrimSpace(c.Query("code"))
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Missing authorization code"})
		return
	}
	_, oauth2Config, verifier, err := ensureProvider(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": err.Error()})
		return
	}
	token, err := oauth2Config.Exchange(c.Request.Context(), code)
	if err != nil {
		recordAudit(c, AuditOpts{Action: auditActionLoginFailed, Success: false, Details: map[string]any{"reason": "token_exchange", "error": err.Error()}})
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
		recordAudit(c, AuditOpts{Action: auditActionLoginFailed, Success: false, Details: map[string]any{"reason": "id_token_verify"}})
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
	roles := ResolveRoles(email, claims)
	if len(roles) == 0 {
		recordAudit(c, AuditOpts{Action: auditActionLoginFailed, Actor: email, Success: false, Details: map[string]any{"reason": "no_roles"}})
		c.Redirect(http.StatusFound, "/admin.html?sso_error=forbidden")
		return
	}
	payload := SessionPayload{
		Subject: idToken.Subject,
		Email:   email,
		Name:    name,
		Roles:   roles,
		Exp:     timeNow().Add(envCfg.sessionTTL()).Unix(),
		Auth:    SessionAuthOIDC,
	}
	if err := setAdminSessionCookie(c, payload); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	recordAudit(c, AuditOpts{
		Action:  auditActionLogin,
		Actor:   AdminActorLabel(payload),
		Success: true,
		Details: map[string]any{"auth": "oidc", "roles": roles},
	})
	c.Redirect(http.StatusFound, "/admin.html")
}

func handleLogout(c *gin.Context) {
	if sess, ok := AdminSessionFromRequest(c); ok {
		recordAudit(c, AuditOpts{
			Action:  auditActionLogout,
			Actor:   AdminActorLabel(sess),
			Success: true,
			Details: map[string]any{"auth": sess.Auth},
		})
	}
	ClearAdminSessionCookie(c)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// timeNow is overridden in tests.
var timeNow = func() time.Time { return time.Now() }
