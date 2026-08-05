package app

import (
	"testing"

	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/domain"
	"grounded_llm_server/internal/locale"
)

func init() {
	domain.BindBundleLocale(locale.BundleLocale)
	domain.BindDefaultLocale(locale.DefaultLocale)
	locale.BindDefaultDomainID(domain.DefaultID)
}

type (
	DomainInfo  = domain.Info
	domainsFile = domain.File
)

func loadDomainCatalog() error {
	return domain.LoadCatalog()
}

func domainsConfigPath() string {
	return domain.ConfigPath()
}

func normalizeDomainID(raw string) (string, error) {
	return domain.NormalizeID(raw)
}

func defaultDomainID() string {
	return domain.DefaultID()
}

func domainInfo(domainID string) (DomainInfo, bool) {
	return domain.Lookup(domainID)
}

func domainDisplayName(d DomainInfo, localeCode string) string {
	return domain.DisplayName(d, localeCode)
}

func domainIDFromQuery(c *gin.Context) string {
	return domain.IDFromQuery(c)
}

func domainIDFromForm(c *gin.Context) string {
	return domain.IDFromForm(c)
}

func requireRAGEnabled(domainID string) error {
	return domain.RequireRAGEnabled(domainID)
}

func domainCatalogLen() int {
	return len(domain.Catalog().Domains)
}

func domainCatalogDefault() string {
	return domain.Catalog().DefaultDomain
}

func domainsConfigForTest(t *testing.T) string {
	return domain.ConfigPathForTest(t)
}
