package oidc

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
	adminSessionCookie = "grounded_admin_session"
	oauthStateCookie   = "grounded_oauth_state"

	// SessionAuthOIDC marks sessions established via OIDC SSO.
	SessionAuthOIDC = "oidc"
)

// SessionPayload is the signed admin session cookie payload.
type SessionPayload struct {
	Subject string   `json:"sub"`
	Email   string   `json:"email"`
	Name    string   `json:"name"`
	Roles   []string `json:"roles"`
	Exp     int64    `json:"exp"`
	Auth    string   `json:"auth"`
}

func sessionSecret() string {
	if s := strings.TrimSpace(envCfg.SessionSecret); s != "" {
		return s
	}
	if c := cfg(); c != nil && c.AdminSecret != "" {
		return c.AdminSecret
	}
	return ""
}

// SignSession returns a signed session token for tests and internal use.
func SignSession(payload SessionPayload) (string, error) {
	secret := sessionSecret()
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

// VerifySession validates a signed session token.
func VerifySession(token string) (SessionPayload, error) {
	var empty SessionPayload
	secret := sessionSecret()
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
	var payload SessionPayload
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

func setAdminSessionCookie(c *gin.Context, payload SessionPayload) error {
	token, err := SignSession(payload)
	if err != nil {
		return err
	}
	maxAge := int(envCfg.sessionTTL().Seconds())
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

// ClearAdminSessionCookie removes the admin session cookie.
func ClearAdminSessionCookie(c *gin.Context) {
	secure := strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") || c.Request.TLS != nil
	c.SetCookie(adminSessionCookie, "", -1, "/", "", secure, true)
}

// AdminSessionFromRequest returns a valid session from the request cookie.
func AdminSessionFromRequest(c *gin.Context) (SessionPayload, bool) {
	token, err := c.Cookie(adminSessionCookie)
	if err != nil || strings.TrimSpace(token) == "" {
		return SessionPayload{}, false
	}
	payload, err := VerifySession(token)
	if err != nil {
		return SessionPayload{}, false
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

// AdminActorLabel returns a human-readable actor label for audit logs.
func AdminActorLabel(payload SessionPayload) string {
	if payload.Email != "" {
		return payload.Email
	}
	if payload.Name != "" {
		return payload.Name
	}
	return payload.Subject
}
