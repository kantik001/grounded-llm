package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type planQuotas struct {
	MessagesPerDay *int `yaml:"messages_per_day" json:"messages_per_day"`
	StorageMB      *int `yaml:"storage_mb" json:"storage_mb"`
	Domains        *int `yaml:"domains" json:"domains"`
}

type planDefinition struct {
	Label         string     `yaml:"label" json:"label"`
	PriceMonthly  *int       `yaml:"price_monthly" json:"price_monthly"`
	ContactSales  bool       `yaml:"contact_sales" json:"contact_sales"`
	Quotas        planQuotas `yaml:"quotas" json:"quotas"`
}

type plansFile struct {
	Version  int                       `yaml:"version"`
	Currency string                    `yaml:"currency"`
	Plans    map[string]planDefinition `yaml:"plans"`
}

var planCatalog plansFile

func plansFilePath() string {
	if p := strings.TrimSpace(os.Getenv("PLANS_FILE")); p != "" {
		return p
	}
	return "config/plans.yaml"
}

func loadPlans() error {
	path := plansFilePath()
	body, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("plans file %s: %w", path, err)
	}
	var f plansFile
	if err := yaml.Unmarshal(body, &f); err != nil {
		return fmt.Errorf("plans parse: %w", err)
	}
	if len(f.Plans) == 0 {
		return fmt.Errorf("plans file has no plans")
	}
	planCatalog = f
	return nil
}

func planQuotasToLimits(q planQuotas) TenantQuotaLimits {
	lim := TenantQuotaLimits{}
	if q.MessagesPerDay != nil {
		lim.MessagesPerDay = *q.MessagesPerDay
	}
	if q.StorageMB != nil {
		lim.StorageMB = *q.StorageMB
	}
	if q.Domains != nil {
		lim.MaxDomains = *q.Domains
	}
	return lim
}

func planByID(planID string) (planDefinition, bool) {
	p, ok := planCatalog.Plans[strings.ToLower(strings.TrimSpace(planID))]
	return p, ok
}

func selfServePlanIDs() []string {
	var ids []string
	for id, p := range planCatalog.Plans {
		if p.ContactSales {
			continue
		}
		ids = append(ids, id)
	}
	return ids
}

func publicPlanList() []map[string]any {
	out := make([]map[string]any, 0, len(planCatalog.Plans))
	for id, p := range planCatalog.Plans {
		item := map[string]any{
			"id":            id,
			"label":         p.Label,
			"contact_sales": p.ContactSales,
			"quotas":        p.Quotas,
		}
		if p.PriceMonthly != nil {
			item["price_monthly"] = *p.PriceMonthly
		}
		if planCatalog.Currency != "" {
			item["currency"] = planCatalog.Currency
		}
		out = append(out, item)
	}
	return out
}
