package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func Run() {
	config = loadConfig()
	loadAPIKeys(config)
	loadAdminUsers(config)
	loadOIDCSettings(config)
	loadTenantQuotas()
	initTenantConfig(config)
	loadTenantRegistry()
	if err := loadPlans(); err != nil {
		log.Printf("Plans file: %v (SaaS billing uses config/plans.yaml when enabled)", err)
	}
	if err := validateProductionConfig(config); err != nil {
		log.Fatalf("%v", err)
	}
	initRedis()
	initTracing()
	initGuardrailsClient(config)
	defer closeGuardrailsClient()
	logStartup(config)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	pool, err := waitForPostgres(ctx, config.DatabaseURL, 30)
	if err != nil {
		log.Fatalf("PostgreSQL: %v", err)
	}
	migDir, err := findMigrationsDir()
	if err != nil {
		log.Fatalf("Migrations: %v", err)
	}
	if err := runAllMigrations(ctx, pool, migDir); err != nil {
		log.Fatalf("Apply migrations: %v", err)
	}
	pool.Close()

	if err := loadDomainCatalog(); err != nil {
		log.Fatalf("Domains config: %v", err)
	}
	initLocaleConfig(config)

	chatStore, err = newChatStore(context.Background(), config.DatabaseURL, config.UploadDir)
	if err != nil {
		log.Fatalf("ChatStore: %v", err)
	}
	bindDeps(config, chatStore)
	defer chatStore.Close()
	// Re-hydrate SaaS tenants/quotas/admin users now that Postgres is available.
	loadTenantRegistry()
	loadTenantQuotas()
	loadAdminUsers(config)
	log.Printf("PostgreSQL: connected, migrations from %s", migDir)
	log.Printf("Domains loaded: %d, default=%s", domainCatalogLen(), domainCatalogDefault())

	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.Use(requestIDMiddleware())
	router.Use(metricsMiddleware())
	router.Use(corsMiddleware(config.CORSAllowedOrigins))
	router.Use(localeMiddleware(config))
	router.Use(defaultJSONContentTypeMiddleware())

	rl := newRateLimiter(config.RateLimitPerMinute, time.Minute)
	if c := llmRedis(); c != nil {
		rl.WithRedis(c)
	}

	registerPublicRoutes(router)
	registerSaaSRoutes(router, rl)
	registerAdminRoutes(router, config)
	registerProtectedRoutes(router, config, rl)
	startConfigReloadWatcher()
	startRetentionWorker(config)

	serverAddr := fmt.Sprintf(":%s", config.ServerPort)
	srv := &http.Server{
		Addr:              serverAddr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		IdleTimeout:       120 * time.Second,
		// WriteTimeout stays 0: SSE streaming + LLM completions can exceed
		// any fixed cap; per-request deadlines are enforced via contexts.
	}

	go func() {
		log.Printf("Server starting on port %s", config.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("Shutdown signal received (%v), draining…", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()
	stopRetentionWorker()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Graceful shutdown error: %v", err)
	} else {
		log.Printf("Server stopped cleanly")
	}
}
