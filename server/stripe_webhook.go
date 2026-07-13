package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func stripeWebhookSecret() string {
	return strings.TrimSpace(os.Getenv("STRIPE_WEBHOOK_SECRET"))
}

type stripeEvent struct {
	Type string          `json:"type"`
	Data stripeEventData `json:"data"`
}

type stripeEventData struct {
	Object json.RawMessage `json:"object"`
}

type stripeCheckoutSession struct {
	Customer string            `json:"customer"`
	Metadata map[string]string `json:"metadata"`
}

type stripeSubscription struct {
	Customer string            `json:"customer"`
	Metadata map[string]string `json:"metadata"`
}

// POST /api/v1/billing/stripe/webhook
func handleStripeWebhook(c *gin.Context) {
	secret := stripeWebhookSecret()
	if secret == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "stripe webhook not configured"})
		return
	}

	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "read body"})
		return
	}
	sig := c.GetHeader("Stripe-Signature")
	if err := verifyStripeSignature(payload, sig, secret, 5*time.Minute); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	var ev stripeEvent
	if err := json.Unmarshal(payload, &ev); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid event JSON"})
		return
	}

	if err := loadPlans(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "plans not loaded"})
		return
	}

	switch ev.Type {
	case "checkout.session.completed":
		if err := handleCheckoutCompleted(ev.Data.Object); err != nil {
			log.Printf("stripe checkout.session.completed: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}
	case "customer.subscription.deleted":
		if err := handleSubscriptionDeleted(ev.Data.Object); err != nil {
			log.Printf("stripe customer.subscription.deleted: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}
	case "invoice.payment_failed":
		log.Printf("stripe invoice.payment_failed received")
	default:
		log.Printf("stripe webhook ignored event: %s", ev.Type)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "received": true})
}

func handleCheckoutCompleted(raw json.RawMessage) error {
	var sess stripeCheckoutSession
	if err := json.Unmarshal(raw, &sess); err != nil {
		return err
	}
	tenantID := strings.TrimSpace(sess.Metadata["tenant_id"])
	planID := strings.ToLower(strings.TrimSpace(sess.Metadata["plan"]))
	if tenantID == "" || planID == "" {
		return errStripeMetadata
	}
	if _, ok := planByID(planID); !ok {
		return errStripeUnknownPlan
	}
	if err := updateTenantPlan(tenantID, planID); err != nil {
		return err
	}
	if sess.Customer != "" {
		_ = updateTenantStripeCustomer(tenantID, sess.Customer)
	}
	return applyPlanQuotas(tenantID, planID)
}

func handleSubscriptionDeleted(raw json.RawMessage) error {
	var sub stripeSubscription
	if err := json.Unmarshal(raw, &sub); err != nil {
		return err
	}
	tenantID := strings.TrimSpace(sub.Metadata["tenant_id"])
	if tenantID == "" {
		return errStripeMetadata
	}
	planID := "starter"
	if err := updateTenantPlan(tenantID, planID); err != nil {
		return err
	}
	return applyPlanQuotas(tenantID, planID)
}

var (
	errStripeMetadata    = stripeWebhookError("missing tenant_id or plan metadata")
	errStripeUnknownPlan = stripeWebhookError("unknown plan in metadata")
)

type stripeWebhookError string

func (e stripeWebhookError) Error() string { return string(e) }

func verifyStripeSignature(payload []byte, header, secret string, tolerance time.Duration) error {
	if header == "" {
		return stripeWebhookError("missing Stripe-Signature header")
	}
	var timestamp int64
	signatures := make(map[string]struct{})
	for _, part := range strings.Split(header, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			ts, err := strconv.ParseInt(kv[1], 10, 64)
			if err != nil {
				return err
			}
			timestamp = ts
		case "v1":
			signatures[kv[1]] = struct{}{}
		}
	}
	if timestamp == 0 || len(signatures) == 0 {
		return stripeWebhookError("invalid Stripe-Signature header")
	}
	if tolerance > 0 {
		age := time.Since(time.Unix(timestamp, 0))
		if age > tolerance || age < -tolerance {
			return stripeWebhookError("timestamp outside tolerance")
		}
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(strconv.FormatInt(timestamp, 10)))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	if _, ok := signatures[expected]; !ok {
		return stripeWebhookError("signature mismatch")
	}
	return nil
}
