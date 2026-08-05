package locale

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	cfgpkg "grounded_llm_server/internal/config"
)

func TestNormalizeLocale(t *testing.T) {
	cases := map[string]string{
		"ru": "ru", "ru-RU": "ru", "en": "en", "en-US": "en", "de": "", "": "",
	}
	for in, want := range cases {
		if got := NormalizeLocale(in); got != want {
			t.Errorf("NormalizeLocale(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestResolveLocaleHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	Init(&cfgpkg.Config{DefaultLocale: "ru"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/branding", nil)
	c.Request.Header.Set(HeaderLocale, "en")
	got := ResolveLocale(c)
	if got != "en" {
		t.Fatalf("ResolveLocale header: got %q", got)
	}
}
