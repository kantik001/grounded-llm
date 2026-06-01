package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNormalizeLocale(t *testing.T) {
	cases := map[string]string{
		"ru": "ru", "ru-RU": "ru", "en": "en", "en-US": "en", "de": "", "": "",
	}
	for in, want := range cases {
		if got := normalizeLocale(in); got != want {
			t.Errorf("normalizeLocale(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestResolveLocaleHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initLocaleConfig(&Config{DefaultLocale: "ru"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/branding", nil)
	c.Request.Header.Set(headerLocale, "en")
	got := resolveLocale(c, &Config{DefaultLocale: "ru"})
	if got != "en" {
		t.Fatalf("resolveLocale header: got %q", got)
	}
}
