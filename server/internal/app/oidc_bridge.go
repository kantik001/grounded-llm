package app

import (
	"grounded_llm_server/internal/admin"
	"grounded_llm_server/internal/oidc"
)

func init() {
	oidc.BindConfig(func() *Config { return config })
	oidc.BindHost(admin.OIDCHost{})
}

func loadOIDCSettings(cfg *Config) {
	oidc.LoadSettings(cfg)
}

func oidcConfigured() bool {
	return oidc.Configured()
}

func resetOIDCProvider() {
	oidc.ResetProvider()
}
