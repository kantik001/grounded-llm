package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func stripeSecretKey() string {
	return strings.TrimSpace(os.Getenv("STRIPE_SECRET_KEY"))
}

func stripeAPIBase() string {
	if base := strings.TrimSpace(os.Getenv("STRIPE_API_BASE")); base != "" {
		return strings.TrimRight(base, "/")
	}
	return "https://api.stripe.com"
}

type checkoutRequest struct {
	TenantID   string `json:"tenant_id"`
	Plan       string `json:"plan"`
	Email      string `json:"email"`
	SuccessURL string `json:"success_url"`
	CancelURL  string `json:"cancel_url"`
}

type stripeCheckoutSessionResponse struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

func registerStripeCheckoutRoute(g *gin.RouterGroup) {
	g.POST("/billing/stripe/checkout", handleStripeCheckout)
}

// POST /api/v1/billing/stripe/checkout
func handleStripeCheckout(c *gin.Context) {
	if !saasSignupEnabled() {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "checkout requires SaaS signup", "code": "signup_disabled"})
		return
	}
	if stripeSecretKey() == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "stripe not configured"})
		return
	}
	if err := loadPlans(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "plans not loaded"})
		return
	}

	var req checkoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid JSON body"})
		return
	}
	tenantID := normalizeTenantID(req.TenantID)
	planID := strings.ToLower(strings.TrimSpace(req.Plan))
	if tenantID == "" || planID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "tenant_id and plan are required"})
		return
	}
	if _, ok := allowedTenants[tenantID]; !ok {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "unknown tenant"})
		return
	}

	plan, ok := planByID(planID)
	if !ok || plan.ContactSales {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid plan"})
		return
	}
	if !planRequiresCheckout(plan) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "plan does not require checkout"})
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		email = tenantSignupEmail(tenantID)
	}
	successURL := strings.TrimSpace(req.SuccessURL)
	cancelURL := strings.TrimSpace(req.CancelURL)
	if successURL == "" {
		successURL = defaultCheckoutSuccessURL()
	}
	if cancelURL == "" {
		cancelURL = defaultCheckoutCancelURL()
	}

	sessionURL, err := createStripeCheckoutSession(stripeCheckoutParams{
		PriceID:      strings.TrimSpace(plan.StripePriceID),
		CustomerEmail: email,
		SuccessURL:   successURL,
		CancelURL:    cancelURL,
		Metadata: map[string]string{
			"tenant_id": tenantID,
			"plan":      planID,
		},
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"checkout_url": sessionURL,
		"tenant_id":    tenantID,
		"plan":         planID,
	})
}

type stripeCheckoutParams struct {
	PriceID       string
	CustomerEmail string
	SuccessURL    string
	CancelURL     string
	Metadata      map[string]string
}

func createStripeCheckoutSession(p stripeCheckoutParams) (string, error) {
	key := stripeSecretKey()
	if key == "" {
		return "", fmt.Errorf("stripe not configured")
	}
	if p.PriceID == "" {
		return "", fmt.Errorf("stripe price id missing")
	}

	form := url.Values{}
	form.Set("mode", "subscription")
	form.Set("line_items[0][price]", p.PriceID)
	form.Set("line_items[0][quantity]", "1")
	form.Set("success_url", p.SuccessURL)
	form.Set("cancel_url", p.CancelURL)
	if p.CustomerEmail != "" {
		form.Set("customer_email", p.CustomerEmail)
	}
	for k, v := range p.Metadata {
		form.Set("metadata["+k+"]", v)
	}
	form.Set("subscription_data[metadata][tenant_id]", p.Metadata["tenant_id"])
	form.Set("subscription_data[metadata][plan]", p.Metadata["plan"])

	req, err := http.NewRequest(http.MethodPost, stripeAPIBase()+"/v1/checkout/sessions", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("stripe checkout: %s", strings.TrimSpace(string(body)))
	}

	var sess stripeCheckoutSessionResponse
	if err := json.Unmarshal(body, &sess); err != nil {
		return "", err
	}
	if sess.URL == "" {
		return "", fmt.Errorf("stripe checkout: empty url")
	}
	return sess.URL, nil
}

func planRequiresCheckout(p planDefinition) bool {
	if p.ContactSales || p.PriceMonthly == nil || *p.PriceMonthly <= 0 {
		return false
	}
	return stripeSecretKey() != "" && strings.TrimSpace(p.StripePriceID) != ""
}

func defaultCheckoutSuccessURL() string {
	if u := strings.TrimSpace(os.Getenv("STRIPE_CHECKOUT_SUCCESS_URL")); u != "" {
		return u
	}
	return "http://localhost/signup.html?checkout=success"
}

func defaultCheckoutCancelURL() string {
	if u := strings.TrimSpace(os.Getenv("STRIPE_CHECKOUT_CANCEL_URL")); u != "" {
		return u
	}
	return "http://localhost/signup.html?checkout=cancel"
}

func tenantSignupEmail(tenantID string) string {
	tenantRegistryMu.Lock()
	defer tenantRegistryMu.Unlock()
	id := normalizeTenantID(tenantID)
	for _, e := range tenantRegistry {
		if normalizeTenantID(e.TenantID) == id {
			return e.Email
		}
	}
	return ""
}

func maybeCreateCheckoutURL(tenantID, planID, email, successURL, cancelURL string) (string, error) {
	plan, ok := planByID(planID)
	if !ok || !planRequiresCheckout(plan) {
		return "", nil
	}
	return createStripeCheckoutSession(stripeCheckoutParams{
		PriceID:       strings.TrimSpace(plan.StripePriceID),
		CustomerEmail: email,
		SuccessURL:    successURL,
		CancelURL:     cancelURL,
		Metadata: map[string]string{
			"tenant_id": normalizeTenantID(tenantID),
			"plan":      planID,
		},
	})
}
