package app

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/admin"
	"grounded_llm_server/internal/analytics"
	"grounded_llm_server/internal/audit"
)

func init() {
	admin.BindConfig(func() *Config { return config })
	admin.BindStore(func() *ChatStore { return chatStore })
	audit.BindStore(func() *ChatStore { return chatStore })
	audit.BindRequestID(ctxRequestID)
	analytics.BindStore(func() *ChatStore { return chatStore })
}

func initKBRegistry(cfg *Config) {
	if cfg == nil {
		return
	}
	if err := admin.InitKBBlobStore(cfg.DataDir); err != nil {
		log.Fatalf("KB blob store: %v", err)
	}
}

func registerAdminRoutes(router *gin.Engine, cfg *Config) {
	_ = cfg
	admin.RegisterRoutes(router)
}

func loadAdminUsers(cfg *Config) {
	admin.LoadUsers(cfg)
}

func saasProvisionAdmin() bool {
	return admin.SaaSProvisionAdmin()
}

func provisionSignupAdminUser(tenantID string) (username, password string, err error) {
	return admin.ProvisionSignupAdminUser(tenantID)
}

func recordRAGAnalytics(ctx context.Context, telegramID int64, tenantID, domainID, question string, result RAGAnswerResult) {
	analytics.RecordRAG(ctx, telegramID, tenantID, domainID, question, result)
}
