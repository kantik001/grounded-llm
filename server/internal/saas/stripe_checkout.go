package saas

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// stripeHTTPClient bounds outbound Stripe calls (http.DefaultClient has no timeout).
var stripeHTTPClient = &http.Client{Timeout: 15 * time.Second}

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

func handleStripeCheckout(c *gin.Context) {
	if !SignupEnabled() {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "checkout requires SaaS signup", "code": "signup_disabled"})
		return
	}
	if StripeSecretKey() == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "stripe not configured"})
		return
	}
	if err := LoadPlans(); err != nil {
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
	if !tenantRegistered(tenantID) {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "unknown tenant"})
		return
	}
	// Ownership check: the caller must present the email the tenant signed up
	// with. Knowing a tenant id alone must not be enough to start billing flows.
	registeredEmail := strings.TrimSpace(strings.ToLower(tenantEmail(tenantID)))
	requestEmail := strings.TrimSpace(strings.ToLower(req.Email))
	if registeredEmail == "" || requestEmail != registeredEmail {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "email does not match tenant registration"})
		return
	}

	plan, ok := PlanByID(planID)
	if !ok || plan.ContactSales {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid plan"})
		return
	}
	if !planRequiresCheckout(plan) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "plan does not require checkout"})
		return
	}

	email := registeredEmail
	successURL := strings.TrimSpace(req.SuccessURL)
	cancelURL := strings.TrimSpace(req.CancelURL)
	if successURL == "" {
		successURL = defaultCheckoutSuccessURL()
	}
	if cancelURL == "" {
		cancelURL = defaultCheckoutCancelURL()
	}

	sessionURL, err := createStripeCheckoutSession(stripeCheckoutParams{
		PriceID:       strings.TrimSpace(plan.StripePriceID),
		CustomerEmail: email,
		SuccessURL:    successURL,
		CancelURL:     cancelURL,
		Metadata: map[string]string{
			"tenant_id": tenantID,
			"plan":      planID,
		},
	})
	if err != nil {
		log.Printf("stripe checkout for tenant %s: %v", tenantID, err)
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": "payment provider error"})
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
	key := StripeSecretKey()
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

	resp, err := stripeHTTPClient.Do(req)
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

func planRequiresCheckout(p PlanDefinition) bool {
	if p.ContactSales || p.PriceMonthly == nil || *p.PriceMonthly <= 0 {
		return false
	}
	return StripeSecretKey() != "" && strings.TrimSpace(p.StripePriceID) != ""
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

func maybeCreateCheckoutURL(tenantID, planID, email, successURL, cancelURL string) (string, error) {
	plan, ok := PlanByID(planID)
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
