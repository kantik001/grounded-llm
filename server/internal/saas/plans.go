package saas

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

// PlanDefinition describes a SaaS pricing tier.
type PlanDefinition struct {
	Label         string     `yaml:"label" json:"label"`
	PriceMonthly  *int       `yaml:"price_monthly" json:"price_monthly"`
	StripePriceID string     `yaml:"stripe_price_id" json:"stripe_price_id,omitempty"`
	ContactSales  bool       `yaml:"contact_sales" json:"contact_sales"`
	Quotas        planQuotas `yaml:"quotas" json:"quotas"`
}

type plansFile struct {
	Version  int                       `yaml:"version"`
	Currency string                    `yaml:"currency"`
	Plans    map[string]PlanDefinition `yaml:"plans"`
}

var planCatalog plansFile

func plansFilePath() string {
	if p := strings.TrimSpace(os.Getenv("PLANS_FILE")); p != "" {
		return p
	}
	return "config/plans.yaml"
}

// LoadPlans reads the plans catalog from PLANS_FILE.
func LoadPlans() error {
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

func quotasToLimits(q planQuotas) QuotaLimits {
	lim := QuotaLimits{}
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

// PlanByID returns a plan definition by id.
func PlanByID(planID string) (PlanDefinition, bool) {
	p, ok := planCatalog.Plans[strings.ToLower(strings.TrimSpace(planID))]
	return p, ok
}

// QuotaLimitsForPlan resolves quota limits for a plan id.
func QuotaLimitsForPlan(planID string) (QuotaLimits, error) {
	plan, ok := PlanByID(planID)
	if !ok {
		return QuotaLimits{}, fmt.Errorf("unknown plan: %s", planID)
	}
	return quotasToLimits(plan.Quotas), nil
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

// SignupEnabled reports whether public self-service signup is enabled.
func SignupEnabled() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("SAAS_SIGNUP_ENABLED")), "true")
}

// StripeSecretKey returns STRIPE_SECRET_KEY when set.
func StripeSecretKey() string {
	return strings.TrimSpace(os.Getenv("STRIPE_SECRET_KEY"))
}

// StripeWebhookSecret returns STRIPE_WEBHOOK_SECRET when set.
func StripeWebhookSecret() string {
	return strings.TrimSpace(os.Getenv("STRIPE_WEBHOOK_SECRET"))
}
