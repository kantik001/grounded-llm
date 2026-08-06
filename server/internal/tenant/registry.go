package tenant

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"grounded_llm_server/internal/store"
)

// RegistryEntry is one SaaS tenant record (Postgres saas_tenants and/or TENANTS_REGISTRY_FILE).
type RegistryEntry struct {
	TenantID         string `json:"tenant_id"`
	OrgName          string `json:"org_name"`
	Email            string `json:"email"`
	Plan             string `json:"plan"`
	CreatedAt        string `json:"created_at"`
	StripeCustomerID string `json:"stripe_customer_id,omitempty"`
}

var (
	registryMu sync.Mutex
	registry   []RegistryEntry
)

// RegistryPath returns TENANTS_REGISTRY_FILE when set.
func RegistryPath() string {
	if p := strings.TrimSpace(os.Getenv("TENANTS_REGISTRY_FILE")); p != "" {
		return p
	}
	return ""
}

// LoadRegistry reads tenants from Postgres (preferred) and/or the JSON file,
// merging ids into the allowlist. Postgres wins on conflict for in-memory cache.
func LoadRegistry() {
	registry = nil
	path := RegistryPath()

	if path != "" {
		body, err := os.ReadFile(path)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Printf("TENANTS_REGISTRY_FILE read error: %v", err)
			}
		} else {
			var entries []RegistryEntry
			if err := json.Unmarshal(body, &entries); err != nil {
				log.Printf("TENANTS_REGISTRY_FILE parse error: %v", err)
			} else {
				registry = entries
			}
		}
	}

	if UsePostgresBackend() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		rows, err := saasStore().ListSaaSTenants(ctx)
		if err != nil {
			log.Printf("saas_tenants load error: %v", err)
		} else if len(rows) > 0 {
			byID := make(map[string]RegistryEntry, len(registry))
			for _, e := range registry {
				byID[NormalizeTenantID(e.TenantID)] = e
			}
			for _, t := range rows {
				id := NormalizeTenantID(t.TenantID)
				byID[id] = RegistryEntry{
					TenantID:         id,
					OrgName:          t.OrgName,
					Email:            t.Email,
					Plan:             t.Plan,
					CreatedAt:        t.CreatedAt.UTC().Format(time.RFC3339),
					StripeCustomerID: t.StripeCustomerID,
				}
			}
			merged := make([]RegistryEntry, 0, len(byID))
			for _, e := range byID {
				merged = append(merged, e)
			}
			registry = merged
			log.Printf("Tenant registry: %d tenant(s) from Postgres (+ optional JSON seed)", len(registry))
		}
	} else if path != "" && len(registry) > 0 {
		log.Printf("Tenant registry: %d tenant(s) from %s", len(registry), path)
	}

	for _, e := range registry {
		id := NormalizeTenantID(e.TenantID)
		if id != "" {
			AllowTenant(id)
		}
	}
}

func registryContains(tenantID string) bool {
	id := NormalizeTenantID(tenantID)
	for _, e := range registry {
		if NormalizeTenantID(e.TenantID) == id {
			return true
		}
	}
	return false
}

func saveRegistryLocked(path string, entries []RegistryEntry) error {
	body, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, body, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func persistTenantPG(entry RegistryEntry) error {
	st := saasStore()
	if st == nil {
		return nil
	}
	createdAt, _ := time.Parse(time.RFC3339, entry.CreatedAt)
	return st.UpsertSaaSTenant(context.Background(), store.SaaSTenant{
		TenantID:         entry.TenantID,
		OrgName:          entry.OrgName,
		Email:            entry.Email,
		Plan:             entry.Plan,
		StripeCustomerID: entry.StripeCustomerID,
		CreatedAt:        createdAt,
	})
}

// RegisterEntry appends a tenant to Postgres (preferred) and optionally the JSON file.
func RegisterEntry(entry RegistryEntry) error {
	path := RegistryPath()
	pg := UsePostgresBackend()
	if path == "" && !pg {
		return fmt.Errorf("TENANTS_REGISTRY_FILE is not configured (or enable Postgres TENANTS_STORE)")
	}

	registryMu.Lock()
	defer registryMu.Unlock()

	id := NormalizeTenantID(entry.TenantID)
	if id == "" {
		return fmt.Errorf("invalid tenant id")
	}
	entry.TenantID = id
	for _, e := range registry {
		if NormalizeTenantID(e.TenantID) == id {
			return fmt.Errorf("tenant already exists: %s", id)
		}
	}
	if IsAllowed(id) && !registryContains(id) {
		return fmt.Errorf("tenant id already reserved: %s", id)
	}

	if pg {
		if err := persistTenantPG(entry); err != nil {
			return err
		}
	}
	registry = append(registry, entry)
	if path != "" {
		if err := saveRegistryLocked(path, registry); err != nil {
			registry = registry[:len(registry)-1]
			return err
		}
	}
	AllowTenant(id)
	return nil
}

// UpdatePlan sets the plan field for a registry tenant.
func UpdatePlan(tenantID, plan string) error {
	path := RegistryPath()
	pg := UsePostgresBackend()
	if path == "" && !pg {
		return fmt.Errorf("TENANTS_REGISTRY_FILE is not configured")
	}
	registryMu.Lock()
	defer registryMu.Unlock()

	id := NormalizeTenantID(tenantID)
	found := false
	for i, e := range registry {
		if NormalizeTenantID(e.TenantID) == id {
			registry[i].Plan = plan
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("tenant not found: %s", id)
	}
	if pg {
		if err := saasStore().UpdateSaaSTenantPlan(context.Background(), id, plan); err != nil {
			return err
		}
	}
	if path != "" {
		return saveRegistryLocked(path, registry)
	}
	return nil
}

// UpdateStripeCustomer stores Stripe customer id for a registry tenant.
func UpdateStripeCustomer(tenantID, customerID string) error {
	path := RegistryPath()
	pg := UsePostgresBackend()
	if path == "" && !pg {
		return fmt.Errorf("TENANTS_REGISTRY_FILE is not configured")
	}
	registryMu.Lock()
	defer registryMu.Unlock()

	id := NormalizeTenantID(tenantID)
	found := false
	for i, e := range registry {
		if NormalizeTenantID(e.TenantID) == id {
			registry[i].StripeCustomerID = customerID
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("tenant not found: %s", id)
	}
	if pg {
		if err := saasStore().UpdateSaaSTenantStripeCustomer(context.Background(), id, customerID); err != nil {
			return err
		}
	}
	if path != "" {
		return saveRegistryLocked(path, registry)
	}
	return nil
}

// NewRegistryEntry builds a registry row for signup.
func NewRegistryEntry(tenantID, orgName, email, plan string) RegistryEntry {
	return RegistryEntry{
		TenantID:  NormalizeTenantID(tenantID),
		OrgName:   orgName,
		Email:     email,
		Plan:      plan,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// SignupEmail returns the signup email for a registry tenant, if known.
func SignupEmail(tenantID string) string {
	registryMu.Lock()
	defer registryMu.Unlock()
	id := NormalizeTenantID(tenantID)
	for _, e := range registry {
		if NormalizeTenantID(e.TenantID) == id {
			return e.Email
		}
	}
	if UsePostgresBackend() {
		email, err := saasStore().GetSaaSTenantEmail(context.Background(), id)
		if err == nil && email != "" {
			return email
		}
	}
	return ""
}

// TenantsRegistryConfigured reports whether signup can persist tenants.
func TenantsRegistryConfigured() bool {
	return RegistryPath() != "" || UsePostgresBackend()
}

// ClaimStripeEvent delegates to the store for webhook idempotency.
func ClaimStripeEvent(eventID, eventType string) (bool, error) {
	st := saasStore()
	if st == nil {
		return true, nil
	}
	return st.ClaimStripeEvent(context.Background(), eventID, eventType)
}
