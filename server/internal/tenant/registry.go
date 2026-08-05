package tenant

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// RegistryEntry is one SaaS tenant record persisted in TENANTS_REGISTRY_FILE.
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

// LoadRegistry reads tenant registry from disk and merges ids into the allowlist.
func LoadRegistry() {
	path := RegistryPath()
	if path == "" {
		registry = nil
		return
	}
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			registry = nil
			return
		}
		log.Printf("TENANTS_REGISTRY_FILE read error: %v", err)
		return
	}
	var entries []RegistryEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		log.Printf("TENANTS_REGISTRY_FILE parse error: %v", err)
		return
	}
	registry = entries
	for _, e := range entries {
		id := NormalizeTenantID(e.TenantID)
		if id != "" {
			AllowTenant(id)
		}
	}
	log.Printf("Tenant registry: %d tenant(s) from %s", len(entries), path)
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

// RegisterEntry appends a tenant to the registry file and allowlist.
func RegisterEntry(entry RegistryEntry) error {
	path := RegistryPath()
	if path == "" {
		return fmt.Errorf("TENANTS_REGISTRY_FILE is not configured")
	}
	registryMu.Lock()
	defer registryMu.Unlock()

	id := NormalizeTenantID(entry.TenantID)
	if id == "" {
		return fmt.Errorf("invalid tenant id")
	}
	for _, e := range registry {
		if NormalizeTenantID(e.TenantID) == id {
			return fmt.Errorf("tenant already exists: %s", id)
		}
	}
	if IsAllowed(id) && !registryContains(id) {
		return fmt.Errorf("tenant id already reserved: %s", id)
	}

	registry = append(registry, entry)
	if err := saveRegistryLocked(path, registry); err != nil {
		registry = registry[:len(registry)-1]
		return err
	}
	AllowTenant(id)
	return nil
}

// UpdatePlan sets the plan field for a registry tenant.
func UpdatePlan(tenantID, plan string) error {
	path := RegistryPath()
	if path == "" {
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
	return saveRegistryLocked(path, registry)
}

// UpdateStripeCustomer stores Stripe customer id for a registry tenant.
func UpdateStripeCustomer(tenantID, customerID string) error {
	path := RegistryPath()
	if path == "" {
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
	return saveRegistryLocked(path, registry)
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
	return ""
}
