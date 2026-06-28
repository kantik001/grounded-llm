package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAuditClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		name string
		set  func(*http.Request)
		want string
	}{
		{
			name: "x-forwarded-for",
			set: func(r *http.Request) {
				r.Header.Set("X-Forwarded-For", "203.0.113.10, 10.0.0.1")
			},
			want: "203.0.113.10",
		},
		{
			name: "x-real-ip",
			set: func(r *http.Request) {
				r.Header.Set("X-Real-IP", "198.51.100.2")
			},
			want: "198.51.100.2",
		},
		{
			name: "remote-addr",
			set: func(r *http.Request) {
				r.RemoteAddr = "192.0.2.5:12345"
			},
			want: "192.0.2.5",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			tc.set(req)
			c.Request = req
			if got := auditClientIP(c); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestIsAdminStatusCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		method string
		path   string
		want   bool
	}{
		{http.MethodGet, "/api/admin/status", true},
		{http.MethodGet, "/admin/status", true},
		{http.MethodPost, "/admin/status", false},
		{http.MethodGet, "/admin/articles", false},
	}
	for _, tc := range cases {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(tc.method, tc.path, nil)
		if got := isAdminStatusCheck(c); got != tc.want {
			t.Fatalf("%s %s: got %v want %v", tc.method, tc.path, got, tc.want)
		}
	}
}

func TestParseAuditLogQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/audit-log?limit=25&offset=5&action=kb_upload", nil)

	limit, offset, action := parseAuditLogQuery(c)
	if limit != 25 || offset != 5 || action != "kb_upload" {
		t.Fatalf("got limit=%d offset=%d action=%q", limit, offset, action)
	}
}

func TestParseAuditLogQueryDefaults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/audit-log", nil)

	limit, offset, action := parseAuditLogQuery(c)
	if limit != auditLogDefaultLimit || offset != 0 || action != "" {
		t.Fatalf("got limit=%d offset=%d action=%q", limit, offset, action)
	}
}
