package admin

import (
	"testing"

	cfgpkg "grounded_llm_server/internal/config"
	"grounded_llm_server/internal/oidc"
)

func TestSafeFilename(t *testing.T) {
	cases := []struct {
		name string
		ok   bool
	}{
		{"article1.txt", true},
		{"my-article_v2.txt", true},
		{"policy_vacation.pdf", true},
		{"handbook.docx", true},
		{"../etc/passwd", false},
		{"article.txt.exe", false},
		{"кириллица.txt", false},
		{"report.doc", false},
	}
	for _, tc := range cases {
		got := safeFilename.MatchString(tc.name)
		if got != tc.ok {
			t.Errorf("%q: got %v want %v", tc.name, got, tc.ok)
		}
	}
}

func TestAuthEnabledOIDC(t *testing.T) {
	cfg := &cfgpkg.Config{AdminSecret: "fallback-secret-key-32bytes-min!!"}
	BindConfig(func() *cfgpkg.Config { return cfg })
	t.Setenv("OIDC_ENABLED", "true")
	t.Setenv("OIDC_ISSUER", "https://issuer.example")
	t.Setenv("OIDC_CLIENT_ID", "client")
	t.Setenv("OIDC_CLIENT_SECRET", "secret")
	t.Setenv("OIDC_REDIRECT_URL", "http://localhost/callback")
	t.Setenv("OIDC_SESSION_SECRET", "test-session-secret-key-min-32-chars!!")
	oidc.LoadSettings(cfg)
	LoadUsers(cfg)

	if !AuthEnabled() {
		t.Fatal("expected admin enabled when OIDC on")
	}

	t.Setenv("OIDC_ENABLED", "false")
	oidc.LoadSettings(cfg)
}
