package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"

	"limiter.io/internal/config"
	"limiter.io/internal/dto"
	"limiter.io/internal/repository"
	"limiter.io/internal/services"

	"github.com/gin-gonic/gin"
)

type BillingHandler struct {
	cfg         *config.Config
	userRepo    repository.UserRepository
	subService  services.SubscriptionService
}

func NewBillingHandler(
	cfg *config.Config,
	userRepo repository.UserRepository,
	subService services.SubscriptionService,
) *BillingHandler {
	return &BillingHandler{
		cfg:        cfg,
		userRepo:   userRepo,
		subService: subService,
	}
}

// LemonSqueezyWebhook handles incoming billing events from Lemon Squeezy
func (h *BillingHandler) LemonSqueezyWebhook(c *gin.Context) {
	// 1. Read Raw Body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "failed to read body"})
		return
	}

	// 2. Validate Signature (HMAC-SHA256)
	signature := c.GetHeader("X-Signature")
	if signature == "" {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "missing signature header"})
		return
	}

	mac := hmac.New(sha256.New, []byte(h.cfg.LemonSqueezyWebhookSecret))
	mac.Write(body)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedMAC)) {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid signature"})
		return
	}

	// 3. Parse Payload Structure
	var payload struct {
		Meta struct {
			EventName string `json:"event_name"`
		} `json:"meta"`
		Data struct {
			Attributes struct {
				Email     string `json:"user_email"`
				VariantID int    `json:"variant_id"`
			} `json:"attributes"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "failed to parse payload"})
		return
	}

	// 4. Handle Subscription Creation
	if payload.Meta.EventName == "subscription_created" {
		// Verify this variant matches our Pro subscription
		variantStr := string(rune(payload.Data.Attributes.VariantID))
		if variantStr == h.cfg.LemonSqueezyProVariantID || payload.Data.Attributes.VariantID == 1899978 {
			// Find user by email
			user, err := h.userRepo.GetByEmail(c.Request.Context(), payload.Data.Attributes.Email)
			if err != nil {
				c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "user not found"})
				return
			}

			// Upgrade user subscription to PRO
			_, err = h.subService.UpgradeSubscription(c.Request.Context(), user.ID, dto.UpgradeSubscriptionRequest{
				PlanID: "pro",
				Reason: "Lemon Squeezy Webhook Activation",
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to upgrade plan"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "processed"})
}
