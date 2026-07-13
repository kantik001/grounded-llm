package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type tenantRegistryEntry struct {
	TenantID         string `json:"tenant_id"`
	OrgName          string `json:"org_name"`
	Email            string `json:"email"`
	Plan             string `json:"plan"`
	CreatedAt        string `json:"created_at"`
	StripeCustomerID string `json:"stripe_customer_id,omitempty"`
}

var (
	tenantRegistryMu sync.Mutex
	tenantRegistry   []tenantRegistryEntry
)

func tenantsRegistryPath() string {
	if p := strings.TrimSpace(os.Getenv("TENANTS_REGISTRY_FILE")); p != "" {
		return p
	}
	return ""
}

func loadTenantRegistry() {
	path := tenantsRegistryPath()
	if path == "" {
		tenantRegistry = nil
		return
	}
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			tenantRegistry = nil
			return
		}
		log.Printf("TENANTS_REGISTRY_FILE read error: %v", err)
		return
	}
	var entries []tenantRegistryEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		log.Printf("TENANTS_REGISTRY_FILE parse error: %v", err)
		return
	}
	tenantRegistry = entries
	for _, e := range entries {
		id := normalizeTenantID(e.TenantID)
		if id != "" {
			allowedTenants[id] = struct{}{}
		}
	}
	log.Printf("Tenant registry: %d tenant(s) from %s", len(entries), path)
}

func tenantRegistryContains(tenantID string) bool {
	id := normalizeTenantID(tenantID)
	for _, e := range tenantRegistry {
		if normalizeTenantID(e.TenantID) == id {
			return true
		}
	}
	return false
}

func saveTenantRegistryLocked(path string, entries []tenantRegistryEntry) error {
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

func registerTenantEntry(entry tenantRegistryEntry) error {
	path := tenantsRegistryPath()
	if path == "" {
		return fmt.Errorf("TENANTS_REGISTRY_FILE is not configured")
	}
	tenantRegistryMu.Lock()
	defer tenantRegistryMu.Unlock()

	id := normalizeTenantID(entry.TenantID)
	if id == "" {
		return fmt.Errorf("invalid tenant id")
	}
	for _, e := range tenantRegistry {
		if normalizeTenantID(e.TenantID) == id {
			return fmt.Errorf("tenant already exists: %s", id)
		}
	}
	if _, ok := allowedTenants[id]; ok && !tenantRegistryContains(id) {
		return fmt.Errorf("tenant id already reserved: %s", id)
	}

	tenantRegistry = append(tenantRegistry, entry)
	if err := saveTenantRegistryLocked(path, tenantRegistry); err != nil {
		tenantRegistry = tenantRegistry[:len(tenantRegistry)-1]
		return err
	}
	allowedTenants[id] = struct{}{}
	return nil
}

func updateTenantPlan(tenantID, plan string) error {
	path := tenantsRegistryPath()
	if path == "" {
		return fmt.Errorf("TENANTS_REGISTRY_FILE is not configured")
	}
	tenantRegistryMu.Lock()
	defer tenantRegistryMu.Unlock()

	id := normalizeTenantID(tenantID)
	found := false
	for i, e := range tenantRegistry {
		if normalizeTenantID(e.TenantID) == id {
			tenantRegistry[i].Plan = plan
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("tenant not found: %s", id)
	}
	return saveTenantRegistryLocked(path, tenantRegistry)
}

func updateTenantStripeCustomer(tenantID, customerID string) error {
	path := tenantsRegistryPath()
	if path == "" {
		return fmt.Errorf("TENANTS_REGISTRY_FILE is not configured")
	}
	tenantRegistryMu.Lock()
	defer tenantRegistryMu.Unlock()

	id := normalizeTenantID(tenantID)
	found := false
	for i, e := range tenantRegistry {
		if normalizeTenantID(e.TenantID) == id {
			tenantRegistry[i].StripeCustomerID = customerID
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("tenant not found: %s", id)
	}
	return saveTenantRegistryLocked(path, tenantRegistry)
}

func newTenantRegistryEntry(tenantID, orgName, email, plan string) tenantRegistryEntry {
	return tenantRegistryEntry{
		TenantID:  normalizeTenantID(tenantID),
		OrgName:   orgName,
		Email:     email,
		Plan:      plan,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}
