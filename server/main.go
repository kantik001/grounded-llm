package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
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
	defer chatStore.Close()
	log.Printf("PostgreSQL: connected, migrations from %s", migDir)
	log.Printf("Domains loaded: %d, default=%s", len(domainCatalog.Domains), domainCatalog.DefaultDomain)

	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.Use(requestIDMiddleware())
	router.Use(metricsMiddleware())
	router.Use(corsMiddleware(config.CORSAllowedOrigins))
	router.Use(localeMiddleware(config))
	router.Use(func(c *gin.Context) {
		if strings.Contains(c.Request.URL.Path, "/media/") {
			c.Next()
			return
		}
		if c.Query("stream") == "1" || c.Query("stream") == "true" {
			c.Next()
			return
		}
		if strings.Contains(c.GetHeader("Accept"), "text/event-stream") {
			c.Next()
			return
		}
		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		c.Next()
	})

	rl := newRateLimiter(config.RateLimitPerMinute, time.Minute)

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
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Graceful shutdown error: %v", err)
	} else {
		log.Printf("Server stopped cleanly")
	}
}
