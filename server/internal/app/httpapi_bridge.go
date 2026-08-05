package app

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"

	cfgpkg "grounded_llm_server/internal/config"
	"grounded_llm_server/internal/httpapi"
	"grounded_llm_server/internal/rag"
	"grounded_llm_server/internal/store"
)

func init() {
	httpapi.Bind(chatHTTPServices{})
}

type chatHTTPServices struct{}

func (chatHTTPServices) Store() *store.ChatStore { return chatStore }
func (chatHTTPServices) Config() *cfgpkg.Config  { return config }
func (chatHTTPServices) ActorUser(c *gin.Context) (*store.TelegramUser, error) {
	return ctxActorUser(c)
}
func (chatHTTPServices) ResolveTenant(c *gin.Context) (string, error) {
	return resolveTenantID(c, config)
}
func (chatHTTPServices) SetTenantID(c *gin.Context, tenantID string) {
	c.Set(ctxKeyTenantID, tenantID)
}
func (chatHTTPServices) TenantID(c *gin.Context) string { return ctxTenantID(c) }
func (chatHTTPServices) NormalizeTenantID(raw string) string {
	return normalizeTenantID(raw)
}
func (chatHTTPServices) NormalizeDomainID(raw string) (string, error) {
	return normalizeDomainID(raw)
}
func (chatHTTPServices) Locale(c *gin.Context) string { return ctxLocale(c) }
func (chatHTTPServices) CheckMessageQuota(ctx context.Context, tenantID string) error {
	return checkMessageQuota(ctx, tenantID)
}
func (chatHTTPServices) QuotaExceeded(c *gin.Context, err error) { quotaErrorResponse(c, err) }
func (chatHTTPServices) JSONError(c *gin.Context, code int, err error) {
	httpapi.JSONError(c, code, err)
}
func (chatHTTPServices) PublicAPIError(err error) string        { return httpapi.PublicAPIError(err) }
func (chatHTTPServices) DomainIDFromForm(c *gin.Context) string { return domainIDFromForm(c) }
func (chatHTTPServices) LogRequest(c *gin.Context, event string, fields map[string]any) {
	logRequest(c, event, fields)
}
func (chatHTTPServices) RecordRAGAnalytics(ctx context.Context, telegramID int64, tenantID, domainID, question string, result rag.AnswerResult) {
	recordRAGAnalytics(ctx, telegramID, tenantID, domainID, question, result)
}

func requestIDMiddleware() gin.HandlerFunc { return httpapi.RequestID() }
func corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return httpapi.CORS(allowedOrigins)
}
func defaultJSONContentTypeMiddleware() gin.HandlerFunc {
	return httpapi.DefaultJSONContentType()
}
func metricsMiddleware() gin.HandlerFunc { return httpapi.MetricsMiddleware() }

func ctxRequestID(c *gin.Context) string { return httpapi.CtxRequestID(c) }
func formatLogValue(v any) string        { return httpapi.FormatLogValue(v) }

func logRequest(c *gin.Context, event string, fields map[string]any) {
	msg := "req=" + ctxRequestID(c) + " event=" + event
	if tid := ctxTenantID(c); tid != "" {
		msg += " tenant=" + tid
	}
	for k, v := range fields {
		msg += " " + k + "=" + formatLogValue(v)
	}
	log.Print(msg)
}

func publicAPIError(err error) string { return httpapi.PublicAPIError(err) }

func handleHealthCheck(c *gin.Context) { httpapi.Health(c) }
func handleReadiness(c *gin.Context)   { httpapi.Ready(c) }

func pathTraceFrom(ctx context.Context) *httpapi.PathTrace {
	return httpapi.PathTraceFrom(ctx)
}

type rateLimiter = httpapi.RateLimiter

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return httpapi.NewRateLimiter(limit, window)
}

func rateLimitMiddleware(rl *rateLimiter) gin.HandlerFunc {
	return rl.Middleware(rateLimitKey)
}

func registerPublicRoutes(router *gin.Engine) {
	httpapi.RegisterPublic(router, httpapi.PublicHandlers{
		Health:     httpapi.Health,
		Ready:      httpapi.Ready,
		Metrics:    httpapi.Metrics,
		Domains:    httpapi.Domains,
		Onboarding: httpapi.Onboarding,
		Branding:   httpapi.Branding,
	})
}

func protectedHandlers() httpapi.ProtectedHandlers {
	return httpapi.ProtectedHandlers{
		Session:  httpapi.NewSession,
		History:  httpapi.History,
		Message:  httpapi.Message,
		Feedback: httpapi.Feedback,
		Media:    httpapi.Media,
		OpenAPI:  httpapi.OpenAPI,
	}
}

func registerProtectedRoutes(router *gin.Engine, cfg *Config, rl *rateLimiter) {
	httpapi.RegisterProtected(router, httpapi.Stack{
		Tenant:    tenantMiddleware(cfg),
		Auth:      combinedAuthMiddleware(cfg),
		RateLimit: rateLimitMiddleware(rl),
	}, protectedHandlers())
}
