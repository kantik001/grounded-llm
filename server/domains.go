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
	ID         string `json:"-"`
	Name       string `json:"name"`
	NameRU     string `json:"name_ru,omitempty"`
	Emoji      string `json:"emoji"`
	RAGEnabled bool   `json:"rag_enabled"`
	UIHidden   bool   `json:"ui_hidden,omitempty"`
}

type domainsFile struct {
	DefaultDomain string                `json:"default_domain"`
	Domains       map[string]DomainInfo `json:"domains"`
}

var domainCatalog domainsFile

type domainPrompts struct {
	RAGSystem    string `json:"rag_system"`
	RAGTaskIntro string `json:"rag_task_intro"`
}

type platformPrompts struct {
	RAGConstraints  string `json:"rag_constraints"`
	VerifyFailHint  string `json:"verify_fail_hint"`
}

var promptCatalog map[string]domainPrompts
var platformPrompt platformPrompts

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
		return "", fmt.Errorf("неизвестный домен: %s", raw)
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

func domainDisplayName(d DomainInfo) string {
	if d.Name != "" {
		return d.Name
	}
	return d.NameRU
}

func loadPromptCatalog() error {
	path := promptsConfigPath()
	body, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read prompts config %s: %w", path, err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return fmt.Errorf("parse prompts config: %w", err)
	}
	promptCatalog = make(map[string]domainPrompts)
	for key, val := range raw {
		if key == "_platform" {
			if err := json.Unmarshal(val, &platformPrompt); err != nil {
				return fmt.Errorf("parse _platform prompts: %w", err)
			}
			continue
		}
		var p domainPrompts
		if err := json.Unmarshal(val, &p); err != nil {
			return fmt.Errorf("parse prompts for %s: %w", key, err)
		}
		promptCatalog[key] = p
	}
	return nil
}

func promptsConfigPath() string {
	return resolveConfigPath("PROMPTS_CONFIG_PATH", defaultConfigCandidates("prompts.json")...)
}

func promptsForDomain(domainID string) domainPrompts {
	if p, ok := promptCatalog[domainID]; ok {
		return p
	}
	if p, ok := promptCatalog[defaultDomainID()]; ok {
		return p
	}
	return domainPrompts{
		RAGSystem:    "Ты — ассистент по документам организации. Отвечай только на основе контекста.",
		RAGTaskIntro: "Отвечай строго на основе контекста.",
	}
}

func ragConstraintsText() string {
	if platformPrompt.RAGConstraints != "" {
		return platformPrompt.RAGConstraints
	}
	return `- НЕ ВЫДУМЫВАЙ. Если ответа нет в контексте — скажи: "В справочных материалах нет информации по вашему вопросу."`
}

func verifyFailHint() string {
	if platformPrompt.VerifyFailHint != "" {
		return platformPrompt.VerifyFailHint
	}
	return "Обратитесь к администратору базы знаний."
}

// GET /domains — список доменов.
func handleListDomains(c *gin.Context) {
	list := make([]gin.H, 0, len(domainCatalog.Domains))
	for id, info := range domainCatalog.Domains {
		if info.UIHidden {
			continue
		}
		list = append(list, gin.H{
			"id":          id,
			"name":        domainDisplayName(info),
			"emoji":       info.Emoji,
			"rag_enabled": info.RAGEnabled,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"default_domain":  domainCatalog.DefaultDomain,
		"domains":         list,
	})
}
