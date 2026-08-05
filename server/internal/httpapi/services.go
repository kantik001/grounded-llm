package httpapi

import (
	"context"

	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/config"
	"grounded_llm_server/internal/rag"
	"grounded_llm_server/internal/store"
)

// Services are app-owned collaborators needed by HTTP handlers.
type Services interface {
	Store() *store.ChatStore
	Config() *config.Config

	ActorUser(c *gin.Context) (*store.TelegramUser, error)
	ResolveTenant(c *gin.Context) (string, error)
	SetTenantID(c *gin.Context, tenantID string)
	TenantID(c *gin.Context) string
	NormalizeTenantID(raw string) string
	NormalizeDomainID(raw string) (string, error)
	Locale(c *gin.Context) string

	CheckMessageQuota(ctx context.Context, tenantID string) error
	QuotaExceeded(c *gin.Context, err error)
	JSONError(c *gin.Context, code int, err error)
	PublicAPIError(err error) string

	DomainIDFromForm(c *gin.Context) string

	LogRequest(c *gin.Context, event string, fields map[string]any)
	RecordRAGAnalytics(ctx context.Context, telegramID int64, tenantID, domainID, question string, result rag.AnswerResult)
}

var svc Services

// Bind wires chat/health services from the app composition root.
func Bind(s Services) {
	svc = s
}

func requireServices() Services {
	if svc == nil {
		panic("httpapi: Bind(Services) was not called")
	}
	return svc
}
