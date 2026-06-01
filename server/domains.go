package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// DomainInfo описывает один домен знаний (workspace).
type DomainInfo struct {
	ID         string            `json:"-"`
	Name       string            `json:"name"`
	Names      map[string]string `json:"names,omitempty"`
	NameRU     string            `json:"name_ru,omitempty"` // legacy
	Emoji      string            `json:"emoji"`
	RAGEnabled bool              `json:"rag_enabled"`
	UIHidden   bool              `json:"ui_hidden,omitempty"`
}

type domainsFile struct {
	DefaultDomain string                `json:"default_domain"`
	Domains       map[string]DomainInfo `json:"domains"`
}

var domainCatalog domainsFile

func loadDomainCatalog() error {
	path := domainsConfigPath()
	body, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read domains config %s: %w", path, err)
	}
	if err := json.Unmarshal(body, &domainCatalog); err != nil {
		return fmt.Errorf("parse domains config: %w", err)
	}
	for id, d := range domainCatalog.Domains {
		d.ID = id
		if d.Name == "" {
			d.Name = d.NameRU
		}
		domainCatalog.Domains[id] = d
	}
	if domainCatalog.DefaultDomain == "" {
		domainCatalog.DefaultDomain = "default"
	}
	return nil
}

func domainsConfigPath() string {
	return resolveConfigPath("DOMAINS_CONFIG_PATH", defaultConfigCandidates("domains.json")...)
}

func normalizeDomainID(raw string) (string, error) {
	id := strings.TrimSpace(strings.ToLower(raw))
	if id == "" {
		id = domainCatalog.DefaultDomain
	}
	if _, ok := domainCatalog.Domains[id]; !ok {
		return "", fmt.Errorf("unknown domain: %s", raw)
	}
	return id, nil
}

func defaultDomainID() string {
	if domainCatalog.DefaultDomain != "" {
		return domainCatalog.DefaultDomain
	}
	return "default"
}

func domainInfo(domainID string) (DomainInfo, bool) {
	d, ok := domainCatalog.Domains[domainID]
	return d, ok
}

func domainDisplayName(d DomainInfo, locale string) string {
	if d.Names != nil {
		if n := strings.TrimSpace(d.Names[bundleLocale(locale)]); n != "" {
			return n
		}
	}
	if d.Name != "" {
		return d.Name
	}
	return d.NameRU
}

// GET /domains — список доменов.
func handleListDomains(c *gin.Context) {
	locale := ctxLocale(c)
	list := make([]gin.H, 0, len(domainCatalog.Domains))
	for id, info := range domainCatalog.Domains {
		if info.UIHidden {
			continue
		}
		list = append(list, gin.H{
			"id":          id,
			"name":        domainDisplayName(info, locale),
			"emoji":       info.Emoji,
			"rag_enabled": info.RAGEnabled,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"default_domain":  domainCatalog.DefaultDomain,
		"locale":          locale,
		"domains":         list,
	})
}
