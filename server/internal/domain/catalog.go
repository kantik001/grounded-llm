package domain

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	cfgpkg "grounded_llm_server/internal/config"
)

// Info describes one knowledge domain (workspace).
type Info struct {
	ID         string            `json:"-"`
	Name       string            `json:"name"`
	Names      map[string]string `json:"names,omitempty"`
	NameRU     string            `json:"name_ru,omitempty"` // legacy
	Emoji      string            `json:"emoji"`
	RAGEnabled bool              `json:"rag_enabled"`
	UIHidden   bool              `json:"ui_hidden,omitempty"`
}

// File is the on-disk domains.json shape.
type File struct {
	DefaultDomain string          `json:"default_domain"`
	Domains       map[string]Info `json:"domains"`
}

var catalog File

// Catalog returns the loaded domain catalog.
func Catalog() File {
	return catalog
}

// SetCatalog replaces the in-memory catalog (tests and app bootstrap).
func SetCatalog(f File) {
	catalog = f
}

// ResetCatalog clears the in-memory catalog before reload.
func ResetCatalog() {
	catalog = File{}
}

// LoadCatalog reads domains.json from the configured path.
func LoadCatalog() error {
	path := ConfigPath()
	body, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read domains config %s: %w", path, err)
	}
	if err := json.Unmarshal(body, &catalog); err != nil {
		return fmt.Errorf("parse domains config: %w", err)
	}
	for id, d := range catalog.Domains {
		d.ID = id
		if d.Name == "" {
			d.Name = d.NameRU
		}
		catalog.Domains[id] = d
	}
	if catalog.DefaultDomain == "" {
		catalog.DefaultDomain = "default"
	}
	return nil
}

// ConfigPath resolves the domains.json file location.
func ConfigPath() string {
	return cfgpkg.ResolvePath("DOMAINS_CONFIG_PATH", cfgpkg.DefaultCandidates("domains.json")...)
}

// NormalizeID validates and normalizes a domain identifier.
func NormalizeID(raw string) (string, error) {
	id := strings.TrimSpace(strings.ToLower(raw))
	if id == "" {
		id = catalog.DefaultDomain
	}
	if _, ok := catalog.Domains[id]; !ok {
		return "", fmt.Errorf("unknown domain: %s", raw)
	}
	return id, nil
}

// DefaultID returns the configured default domain.
func DefaultID() string {
	if catalog.DefaultDomain != "" {
		return catalog.DefaultDomain
	}
	return "default"
}

// Lookup returns domain metadata by ID.
func Lookup(domainID string) (Info, bool) {
	d, ok := catalog.Domains[domainID]
	return d, ok
}

// DisplayName picks a localized display name for a domain.
func DisplayName(d Info, locale string) string {
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

// ListEntry is one visible domain for public API responses.
type ListEntry struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Emoji      string `json:"emoji"`
	RAGEnabled bool   `json:"rag_enabled"`
}

// VisibleEntries returns non-hidden domains with localized names.
func VisibleEntries(locale string) []ListEntry {
	list := make([]ListEntry, 0, len(catalog.Domains))
	for id, info := range catalog.Domains {
		if info.UIHidden {
			continue
		}
		list = append(list, ListEntry{
			ID:         id,
			Name:       DisplayName(info, locale),
			Emoji:      info.Emoji,
			RAGEnabled: info.RAGEnabled,
		})
	}
	return list
}
