package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type oidcEnvConfig struct {
	Enabled       bool
	Issuer        string
	ClientID      string
	ClientSecret  string
	RedirectURL   string
	Scopes        []string
	SessionSecret string
	SessionTTL    time.Duration
}

var oidcCfg oidcEnvConfig

func loadOIDCSettings(cfg *Config) {
	oidcCfg = oidcEnvConfig{
		Enabled:       strings.EqualFold(strings.TrimSpace(os.Getenv("OIDC_ENABLED")), "true"),
		Issuer:        strings.TrimSpace(os.Getenv("OIDC_ISSUER")),
		ClientID:      strings.TrimSpace(os.Getenv("OIDC_CLIENT_ID")),
		ClientSecret:  strings.TrimSpace(os.Getenv("OIDC_CLIENT_SECRET")),
		RedirectURL:   strings.TrimSpace(os.Getenv("OIDC_REDIRECT_URL")),
		SessionSecret: strings.TrimSpace(os.Getenv("OIDC_SESSION_SECRET")),
	}
	rawScopes := strings.TrimSpace(os.Getenv("OIDC_SCOPES"))
	if rawScopes == "" {
		oidcCfg.Scopes = []string{"openid", "profile", "email"}
	} else {
		for _, s := range strings.FieldsFunc(rawScopes, func(r rune) bool { return r == ' ' || r == ',' }) {
			s = strings.TrimSpace(s)
			if s != "" {
				oidcCfg.Scopes = append(oidcCfg.Scopes, s)
			}
		}
	}
	hours, _ := strconv.Atoi(strings.TrimSpace(os.Getenv("OIDC_SESSION_TTL_HOURS")))
	if hours <= 0 {
		hours = 12
	}
	oidcCfg.SessionTTL = time.Duration(hours) * time.Hour

	if !oidcCfg.Enabled {
		return
	}
	if oidcCfg.Issuer == "" || oidcCfg.ClientID == "" || oidcCfg.ClientSecret == "" || oidcCfg.RedirectURL == "" {
		log.Printf("OIDC_ENABLED but missing OIDC_ISSUER/CLIENT_ID/CLIENT_SECRET/REDIRECT_URL — SSO disabled")
		oidcCfg.Enabled = false
		return
	}
	if strings.TrimSpace(oidcCfg.SessionSecret) == "" && strings.TrimSpace(cfg.AdminSecret) == "" {
		log.Printf("OIDC requires OIDC_SESSION_SECRET or ADMIN_SECRET for session signing")
		oidcCfg.Enabled = false
		return
	}
	loadOIDCRoleMapping()
	resetOIDCProvider()
	log.Printf("OIDC SSO enabled (issuer=%s)", oidcCfg.Issuer)
}

func (c oidcEnvConfig) sessionTTL() time.Duration {
	if c.SessionTTL <= 0 {
		return 12 * time.Hour
	}
	return c.SessionTTL
}

func oidcConfigured() bool {
	return oidcCfg.Enabled
}

type oidcRoleMapping struct {
	DefaultRoles []string            `json:"default_roles"`
	Claim        string              `json:"claim"`
	Groups       map[string][]string `json:"groups"`
	Emails       map[string][]string `json:"emails"`
}

var oidcRoleMap oidcRoleMapping

func loadOIDCRoleMapping() {
	oidcRoleMap = oidcRoleMapping{
		DefaultRoles: []string{RoleKBEditor},
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
	var raw oidcRoleMapping
	if err := json.Unmarshal(body, &raw); err != nil {
		log.Printf("OIDC_ROLE_MAPPING_FILE parse error: %v", err)
		return
	}
	if len(raw.DefaultRoles) > 0 {
		oidcRoleMap.DefaultRoles = normalizeRoles(raw.DefaultRoles)
	}
	if strings.TrimSpace(raw.Claim) != "" {
		oidcRoleMap.Claim = strings.TrimSpace(raw.Claim)
	}
	if raw.Groups != nil {
		oidcRoleMap.Groups = raw.Groups
	}
	if raw.Emails != nil {
		oidcRoleMap.Emails = raw.Emails
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

func resolveOIDCRoles(email string, claims map[string]any) []string {
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
	for k, mapped := range oidcRoleMap.Emails {
		if strings.EqualFold(k, emailKey) {
			add(mapped)
		}
	}

	claimName := oidcRoleMap.Claim
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
		if mapped, ok := oidcRoleMap.Groups[g]; ok {
			add(mapped)
		}
		for name, mapped := range oidcRoleMap.Groups {
			if strings.EqualFold(name, g) {
				add(mapped)
			}
		}
	}

	if len(roles) == 0 {
		add(oidcRoleMap.DefaultRoles)
	}
	return roles
}
