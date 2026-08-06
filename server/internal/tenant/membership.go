package tenant

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CtxKeyMemberTenants is the gin key for Telegram membership allowlist ([]string).
const CtxKeyMemberTenants = "tenant_memberships"

// MembershipEnforce reports whether Telegram users must belong to a tenant.
// Default off (backward compatible). Set TENANT_MEMBERSHIP_ENFORCE=1 to enable.
func MembershipEnforce() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("TENANT_MEMBERSHIP_ENFORCE")))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

// MembershipAutoDefault grants the default tenant on first Telegram login when
// enforce is on and the user has no memberships yet (default true when enforce).
func MembershipAutoDefault() bool {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv("TENANT_MEMBERSHIP_AUTO_DEFAULT")))
	if raw == "" {
		return MembershipEnforce()
	}
	return raw == "1" || raw == "true" || raw == "yes" || raw == "on"
}

// BindTelegramMembership loads memberships for telegramID and pins the request.
// When enforce is off and the user has no rows, behavior is unchanged (any allowlisted tenant).
func BindTelegramMembership(c *gin.Context, telegramID int64) error {
	if telegramID == 0 {
		return nil
	}
	st := saasStore()
	if st == nil {
		if MembershipEnforce() {
			return fmt.Errorf("tenant membership store unavailable")
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	members, err := st.ListUserTenantMemberships(ctx, telegramID)
	if err != nil {
		return err
	}

	if len(members) == 0 {
		if !MembershipEnforce() {
			return nil
		}
		if !MembershipAutoDefault() {
			return fmt.Errorf("no tenant membership for this user")
		}
		def := "default"
		if cfg := cfg(); cfg != nil && cfg.DefaultTenantID != "" {
			def = NormalizeTenantID(cfg.DefaultTenantID)
		}
		if err := st.UpsertUserTenantMembership(ctx, telegramID, def, "member"); err != nil {
			return err
		}
		members = []string{def}
		log.Printf("tenant membership: auto-granted %s to telegram_id=%d", def, telegramID)
	}

	normalized := make([]string, 0, len(members))
	for _, m := range members {
		id := NormalizeTenantID(m)
		if id != "" {
			normalized = append(normalized, id)
		}
	}
	c.Set(CtxKeyMemberTenants, normalized)

	if len(normalized) == 1 {
		return BindCredentialTenant(c, normalized[0])
	}
	// Multi-tenant user: allow any of their memberships ("*" with membership check).
	c.Set(CtxKeyBoundTenant, "*")
	if requested, explicit := ExplicitID(c); explicit {
		if !memberContains(normalized, requested) {
			return fmt.Errorf("credentials are not authorized for tenant %q", requested)
		}
		c.Set(CtxKeyTenantID, requested)
	} else {
		// Prefer default if member, else first membership.
		def := "default"
		if cfg := cfg(); cfg != nil && cfg.DefaultTenantID != "" {
			def = NormalizeTenantID(cfg.DefaultTenantID)
		}
		if memberContains(normalized, def) {
			c.Set(CtxKeyTenantID, def)
		} else {
			c.Set(CtxKeyTenantID, normalized[0])
		}
	}
	return nil
}

func memberContains(members []string, id string) bool {
	id = NormalizeTenantID(id)
	for _, m := range members {
		if m == id {
			return true
		}
	}
	return false
}

// MemberTenants returns memberships stored on the request, if any.
func MemberTenants(c *gin.Context) ([]string, bool) {
	if v, ok := c.Get(CtxKeyMemberTenants); ok {
		if s, ok := v.([]string); ok && len(s) > 0 {
			return s, true
		}
	}
	return nil, false
}

// RequestOverride applies a client-requested tenant override, enforcing
// allowlist, credential binding, and Telegram memberships when present.
func RequestOverride(c *gin.Context, raw string) error {
	id := NormalizeTenantID(raw)
	if id == "" {
		return nil
	}
	if len(allowedTenants) > 0 {
		if _, ok := allowedTenants[id]; !ok {
			return fmt.Errorf("unknown tenant: %s", id)
		}
	}
	if bound, ok := BoundTenant(c); ok && bound != "*" && bound != id {
		return fmt.Errorf("credentials are not authorized for tenant %q", id)
	}
	if members, ok := MemberTenants(c); ok && !memberContains(members, id) {
		return fmt.Errorf("credentials are not authorized for tenant %q", id)
	}
	c.Set(CtxKeyTenantID, id)
	c.Set(CtxKeyTenantExplicit, true)
	return nil
}

// GrantMembership is a helper for admin/signup flows.
func GrantMembership(telegramID int64, tenantID string) error {
	st := saasStore()
	if st == nil {
		return fmt.Errorf("store not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return st.UpsertUserTenantMembership(ctx, telegramID, NormalizeTenantID(tenantID), "member")
}
