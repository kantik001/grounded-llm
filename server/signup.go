package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func saasSignupEnabled() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("SAAS_SIGNUP_ENABLED")), "true")
}

func registerSaaSRoutes(router *gin.Engine, rl *rateLimiter) {
	lim := rateLimitMiddleware(rl)
	g := router.Group("/api/v1")
	g.Use(lim)
	g.GET("/plans", handleListPlans)
	g.POST("/signup", handleSignup)
	registerStripeCheckoutRoute(g)
	g.POST("/billing/stripe/webhook", handleStripeWebhook)
}

// GET /api/v1/plans
func handleListPlans(c *gin.Context) {
	if err := loadPlans(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "plans not loaded"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"currency": planCatalog.Currency,
		"plans":    publicPlanList(),
	})
}

type signupRequest struct {
	OrgName string `json:"org_name"`
	Email   string `json:"email"`
	Plan    string `json:"plan"`
}

// POST /api/v1/signup
func handleSignup(c *gin.Context) {
	if !saasSignupEnabled() {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "signup is disabled", "code": "signup_disabled"})
		return
	}
	if tenantsRegistryPath() == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "tenant registry not configured"})
		return
	}
	if err := loadPlans(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "plans not loaded"})
		return
	}

	var req signupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid JSON body"})
		return
	}
	orgName := strings.TrimSpace(req.OrgName)
	email := strings.TrimSpace(strings.ToLower(req.Email))
	planID := strings.ToLower(strings.TrimSpace(req.Plan))
	if orgName == "" || email == "" || planID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "org_name, email, and plan are required"})
		return
	}
	if !strings.Contains(email, "@") {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid email"})
		return
	}

	plan, ok := planByID(planID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "unknown plan"})
		return
	}
	if plan.ContactSales {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "plan requires sales contact", "code": "contact_sales"})
		return
	}

	tenantID, err := allocateTenantID(orgName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	entry := newTenantRegistryEntry(tenantID, orgName, email, planID)
	if err := registerTenantEntry(entry); err != nil {
		c.JSON(http.StatusConflict, gin.H{"success": false, "error": err.Error()})
		return
	}
	effectivePlan := planID
	if planRequiresCheckout(plan) {
		effectivePlan = "starter"
	}
	if err := applyPlanQuotas(tenantID, effectivePlan); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if err := provisionTenantDataDir(tenantID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	resp := gin.H{
		"success":   true,
		"tenant_id": tenantID,
		"plan":      planID,
		"chat_url":  fmt.Sprintf("/?tenant_id=%s", tenantID),
		"admin_url": fmt.Sprintf("/admin.html?tenant_id=%s", tenantID),
	}

	if saasProvisionAdmin() {
		username, password, err := provisionSignupAdminUser(tenantID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
			return
		}
		if username != "" {
			resp["admin_username"] = username
			resp["admin_password"] = password
		}
	}

	if planRequiresCheckout(plan) {
		checkoutURL, err := maybeCreateCheckoutURL(tenantID, planID, email, defaultCheckoutSuccessURL(), defaultCheckoutCancelURL())
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": err.Error()})
			return
		}
		if checkoutURL != "" {
			resp["checkout_url"] = checkoutURL
			resp["payment_pending"] = true
		}
	}

	c.JSON(http.StatusCreated, resp)
}

func allocateTenantID(orgName string) (string, error) {
	base := slugRe.ReplaceAllString(strings.ToLower(orgName), "-")
	base = strings.Trim(base, "-")
	if base == "" {
		base = "org"
	}
	if len(base) > 24 {
		base = base[:24]
	}
	for i := 0; i < 8; i++ {
		suffix, err := randomHex(3)
		if err != nil {
			return "", err
		}
		candidate := normalizeTenantID(base + "-" + suffix)
		if _, ok := allowedTenants[candidate]; !ok {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not allocate tenant id")
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func provisionTenantDataDir(tenantID string) error {
	if config == nil {
		return fmt.Errorf("server not configured")
	}
	domainID := defaultDomainID()
	dir := filepath.Join(config.DataDir, normalizeTenantID(tenantID), domainID)
	return os.MkdirAll(dir, 0o755)
}
