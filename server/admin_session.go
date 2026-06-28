package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	adminSessionCookie   = "grounded_admin_session"
	oauthStateCookie     = "grounded_oauth_state"
	adminSessionAuthOIDC = "oidc"
)

type adminSessionPayload struct {
	Subject string   `json:"sub"`
	Email   string   `json:"email"`
	Name    string   `json:"name"`
	Roles   []string `json:"roles"`
	Exp     int64    `json:"exp"`
	Auth    string   `json:"auth"`
}

func adminSessionSecret() string {
	if s := strings.TrimSpace(oidcCfg.SessionSecret); s != "" {
		return s
	}
	if config != nil && config.AdminSecret != "" {
		return config.AdminSecret
	}
	return ""
}

func signAdminSession(payload adminSessionPayload) (string, error) {
	secret := adminSessionSecret()
	if secret == "" {
		return "", errors.New("session secret not configured")
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	body := base64.RawURLEncoding.EncodeToString(raw)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return body + "." + sig, nil
}

func verifyAdminSession(token string) (adminSessionPayload, error) {
	var empty adminSessionPayload
	secret := adminSessionSecret()
	if secret == "" {
		return empty, errors.New("session secret not configured")
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return empty, errors.New("invalid session token")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(parts[0]))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[1])) {
		return empty, errors.New("invalid session signature")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return empty, err
	}
	var payload adminSessionPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return empty, err
	}
	if payload.Exp > 0 && time.Now().Unix() > payload.Exp {
		return empty, errors.New("session expired")
	}
	if len(payload.Roles) == 0 {
		return empty, errors.New("session has no roles")
	}
	return payload, nil
}

func setAdminSessionCookie(c *gin.Context, payload adminSessionPayload) error {
	token, err := signAdminSession(payload)
	if err != nil {
		return err
	}
	maxAge := int(oidcCfg.sessionTTL().Seconds())
	if payload.Exp > 0 {
		maxAge = int(time.Until(time.Unix(payload.Exp, 0)).Seconds())
		if maxAge < 0 {
			maxAge = 0
		}
	}
	c.SetSameSite(http.SameSiteLaxMode)
	secure := strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") || c.Request.TLS != nil
	c.SetCookie(adminSessionCookie, token, maxAge, "/", "", secure, true)
	return nil
}

func clearAdminSessionCookie(c *gin.Context) {
	secure := strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") || c.Request.TLS != nil
	c.SetCookie(adminSessionCookie, "", -1, "/", "", secure, true)
}

func adminSessionFromRequest(c *gin.Context) (adminSessionPayload, bool) {
	token, err := c.Cookie(adminSessionCookie)
	if err != nil || strings.TrimSpace(token) == "" {
		return adminSessionPayload{}, false
	}
	payload, err := verifyAdminSession(token)
	if err != nil {
		return adminSessionPayload{}, false
	}
	return payload, true
}

func newOAuthState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func setOAuthStateCookie(c *gin.Context, state string) {
	secure := strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") || c.Request.TLS != nil
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(oauthStateCookie, state, 600, "/", "", secure, true)
}

func popOAuthStateCookie(c *gin.Context) (string, bool) {
	state, err := c.Cookie(oauthStateCookie)
	secure := strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") || c.Request.TLS != nil
	c.SetCookie(oauthStateCookie, "", -1, "/", "", secure, true)
	if err != nil || strings.TrimSpace(state) == "" {
		return "", false
	}
	return state, true
}

func adminActorLabel(payload adminSessionPayload) string {
	if payload.Email != "" {
		return payload.Email
	}
	if payload.Name != "" {
		return payload.Name
	}
	return payload.Subject
}
