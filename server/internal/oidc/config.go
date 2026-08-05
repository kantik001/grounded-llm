package oidc

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"grounded_llm_server/internal/config"
)

const (
	defaultEditorRole = "kb_editor"
	defaultAdminRole  = "admin"
)

type envConfig struct {
	Enabled       bool
	Issuer        string
	ClientID      string
	ClientSecret  string
	RedirectURL   string
	Scopes        []string
	SessionSecret string
	SessionTTL    time.Duration
}

var envCfg envConfig

// LoadSettings reads OIDC env vars and validates against process config.
func LoadSettings(cfg *config.Config) {
	envCfg = envConfig{
		Enabled:       strings.EqualFold(strings.TrimSpace(os.Getenv("OIDC_ENABLED")), "true"),
		Issuer:        strings.TrimSpace(os.Getenv("OIDC_ISSUER")),
		ClientID:      strings.TrimSpace(os.Getenv("OIDC_CLIENT_ID")),
		ClientSecret:  strings.TrimSpace(os.Getenv("OIDC_CLIENT_SECRET")),
		RedirectURL:   strings.TrimSpace(os.Getenv("OIDC_REDIRECT_URL")),
		SessionSecret: strings.TrimSpace(os.Getenv("OIDC_SESSION_SECRET")),
	}
	rawScopes := strings.TrimSpace(os.Getenv("OIDC_SCOPES"))
	if rawScopes == "" {
		envCfg.Scopes = []string{"openid", "profile", "email"}
	} else {
		for _, s := range strings.FieldsFunc(rawScopes, func(r rune) bool { return r == ' ' || r == ',' }) {
			s = strings.TrimSpace(s)
			if s != "" {
				envCfg.Scopes = append(envCfg.Scopes, s)
			}
		}
	}
	hours, _ := strconv.Atoi(strings.TrimSpace(os.Getenv("OIDC_SESSION_TTL_HOURS")))
	if hours <= 0 {
		hours = 12
	}
	envCfg.SessionTTL = time.Duration(hours) * time.Hour

	if !envCfg.Enabled {
		return
	}
	if envCfg.Issuer == "" || envCfg.ClientID == "" || envCfg.ClientSecret == "" || envCfg.RedirectURL == "" {
		log.Printf("OIDC_ENABLED but missing OIDC_ISSUER/CLIENT_ID/CLIENT_SECRET/REDIRECT_URL — SSO disabled")
		envCfg.Enabled = false
		return
	}
	if strings.TrimSpace(envCfg.SessionSecret) == "" && cfg != nil && strings.TrimSpace(cfg.AdminSecret) == "" {
		log.Printf("OIDC requires OIDC_SESSION_SECRET or ADMIN_SECRET for session signing")
		envCfg.Enabled = false
		return
	}
	loadRoleMapping()
	ResetProvider()
	log.Printf("OIDC SSO enabled (issuer=%s)", envCfg.Issuer)
}

func (c envConfig) sessionTTL() time.Duration {
	if c.SessionTTL <= 0 {
		return 12 * time.Hour
	}
	return c.SessionTTL
}

// Configured reports whether OIDC SSO is enabled and valid.
func Configured() bool {
	return envCfg.Enabled
}

// Issuer returns the configured OIDC issuer URL (empty when disabled).
func Issuer() string {
	return envCfg.Issuer
}

type roleMapping struct {
	DefaultRoles []string            `json:"default_roles"`
	Claim        string              `json:"claim"`
	Groups       map[string][]string `json:"groups"`
	Emails       map[string][]string `json:"emails"`
}

var roleMap roleMapping

func loadRoleMapping() {
	roleMap = roleMapping{
		DefaultRoles: []string{defaultEditorRole},
		Claim:        "groups",
		Groups:       map[string][]string{},
		Emails:       map[string][]string{},
	}
	path := strings.TrimSpace(os.Getenv("OIDC_ROLE_MAPPING_FILE"))
	if path == "" {
		return
	}
	body, err := os.ReadFile(path)
	if err != nil {
		log.Printf("OIDC_ROLE_MAPPING_FILE read error: %v", err)
		return
	}
	var raw roleMapping
	if err := json.Unmarshal(body, &raw); err != nil {
		log.Printf("OIDC_ROLE_MAPPING_FILE parse error: %v", err)
		return
	}
	if len(raw.DefaultRoles) > 0 {
		roleMap.DefaultRoles = normalizeRoles(raw.DefaultRoles)
	}
	if strings.TrimSpace(raw.Claim) != "" {
		roleMap.Claim = strings.TrimSpace(raw.Claim)
	}
	if raw.Groups != nil {
		roleMap.Groups = raw.Groups
	}
	if raw.Emails != nil {
		roleMap.Emails = raw.Emails
	}
}

func stringSliceClaim(v any) []string {
	switch x := v.(type) {
	case []string:
		return x
	case []any:
		var out []string
		for _, item := range x {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	case string:
		if x != "" {
			return []string{x}
		}
	}
	return nil
}

// ResolveRoles maps OIDC claims to admin RBAC roles.
func ResolveRoles(email string, claims map[string]any) []string {
	seen := make(map[string]struct{})
	var roles []string
	add := func(list []string) {
		for _, r := range normalizeRoles(list) {
			if _, ok := seen[r]; ok {
				continue
			}
			seen[r] = struct{}{}
			roles = append(roles, r)
		}
	}

	emailKey := strings.ToLower(strings.TrimSpace(email))
	for k, mapped := range roleMap.Emails {
		if strings.EqualFold(k, emailKey) {
			add(mapped)
		}
	}

	claimName := roleMap.Claim
	if claimName == "" {
		claimName = "groups"
	}
	groupValues := stringSliceClaim(claims[claimName])
	if len(groupValues) == 0 && claimName != "groups" {
		groupValues = stringSliceClaim(claims["groups"])
	}
	if len(groupValues) == 0 {
		groupValues = stringSliceClaim(claims["roles"])
	}
	for _, g := range groupValues {
		if mapped, ok := roleMap.Groups[g]; ok {
			add(mapped)
		}
		for name, mapped := range roleMap.Groups {
			if strings.EqualFold(name, g) {
				add(mapped)
			}
		}
	}

	if len(roles) == 0 {
		add(roleMap.DefaultRoles)
	}
	return roles
}
