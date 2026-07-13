package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func reloadRuntimeConfig() error {
	if err := loadDomainCatalog(); err != nil {
		return err
	}
	if err := loadAllLocaleBundles(); err != nil {
		return err
	}
	loadAPIKeys(config)
	loadAdminUsers(config)
	loadOIDCSettings(config)
	loadTenantQuotas()
	loadTenantRegistry()
	if err := loadPlans(); err != nil {
		log.Printf("plans reload: %v", err)
	}
	resetOIDCProvider()
	log.Printf("Config reloaded: domains=%d api_keys=%d admin_users=%d tenant_quotas=%d",
		len(domainCatalog.Domains), len(apiKeyRegistry), len(adminUserRegistry), len(tenantQuotaRegistry))
	return nil
}

func startConfigReloadWatcher() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP)
	go func() {
		for range sigCh {
			if err := reloadRuntimeConfig(); err != nil {
				log.Printf("SIGHUP config reload failed: %v", err)
			}
		}
	}()

	sec, _ := strconv.Atoi(os.Getenv("CONFIG_RELOAD_INTERVAL_SEC"))
	if sec <= 0 {
		return
	}
	go func() {
		ticker := time.NewTicker(time.Duration(sec) * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := reloadRuntimeConfig(); err != nil {
				log.Printf("periodic config reload failed: %v", err)
			}
		}
	}()
}
