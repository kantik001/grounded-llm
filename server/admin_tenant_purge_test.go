package main

import "testing"

func TestValidateTenantPurgeTarget(t *testing.T) {
	config = &Config{DefaultTenantID: "default"}

	if err := validateTenantPurgeTarget("acme", true, false); err != nil {
		t.Fatalf("acme: %v", err)
	}
	if err := validateTenantPurgeTarget("acme", false, false); err == nil {
		t.Fatal("expected confirm error")
	}
	if err := validateTenantPurgeTarget("default", true, false); err == nil {
		t.Fatal("expected default tenant guard")
	}
	if err := validateTenantPurgeTarget("default", true, true); err != nil {
		t.Fatalf("default with purge_default: %v", err)
	}
	if err := validateTenantPurgeTarget("../bad", true, false); err == nil {
		t.Fatal("expected invalid tenant_id")
	}
}
