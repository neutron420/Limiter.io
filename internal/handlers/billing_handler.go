package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"limiter.io/internal/config"
	"limiter.io/internal/dto"
	"limiter.io/internal/models"
	"limiter.io/internal/repository"
	"limiter.io/internal/services"

	"github.com/gin-gonic/gin"
)

type BillingHandler struct {
	cfg         *config.Config
	userRepo    repository.UserRepository
	subService  services.SubscriptionService
	webhookRepo repository.WebhookEventRepository
}

func NewBillingHandler(
	cfg *config.Config,
	userRepo repository.UserRepository,
	subService services.SubscriptionService,
	webhookRepo repository.WebhookEventRepository,
) *BillingHandler {
	return &BillingHandler{
		cfg:         cfg,
		userRepo:    userRepo,
		subService:  subService,
		webhookRepo: webhookRepo,
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
	mac := hmac.New(sha256.New, []byte(h.cfg.LemonSqueezyWebhookSecret))
	mac.Write(body)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	verified := hmac.Equal([]byte(signature), []byte(expectedMAC))

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

	_ = json.Unmarshal(body, &payload)

	status := "processed"
	detail := string(body)

	if !verified {
		status = "failed_signature"
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid signature"})
		h.saveWebhookEvent(c, "lemon_squeezy", payload.Meta.EventName, payload.Data.Attributes.Email, false, status, detail)
		return
	}

	if payload.Meta.EventName == "subscription_created" {
		variantStr := strconv.Itoa(payload.Data.Attributes.VariantID)
		if variantStr == h.cfg.LemonSqueezyProVariantID {
			user, err := h.userRepo.GetByEmail(c.Request.Context(), payload.Data.Attributes.Email)
			if err != nil {
				status = "error_user_not_found"
				c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "user not found"})
				h.saveWebhookEvent(c, "lemon_squeezy", payload.Meta.EventName, payload.Data.Attributes.Email, true, status, detail)
				return
			}

			_, err = h.subService.UpgradeSubscription(c.Request.Context(), user.ID, dto.UpgradeSubscriptionRequest{
				PlanID: "pro",
				Reason: "Lemon Squeezy Webhook Activation",
			})
			if err != nil {
				status = "error_upgrade_failed"
				c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to upgrade plan"})
				h.saveWebhookEvent(c, "lemon_squeezy", payload.Meta.EventName, payload.Data.Attributes.Email, true, status, detail)
				return
			}
		} else {
			status = "ignored_variant"
		}
	} else {
		status = "ignored_event"
	}

	h.saveWebhookEvent(c, "lemon_squeezy", payload.Meta.EventName, payload.Data.Attributes.Email, true, status, detail)
	c.JSON(http.StatusOK, gin.H{"status": "processed"})
}

func (h *BillingHandler) saveWebhookEvent(c *gin.Context, source, event, email string, verified bool, status, detail string) {
	evt := &models.WebhookEvent{
		Source:     source,
		EventName:  event,
		Email:      email,
		Verified:   verified,
		Status:     status,
		Detail:     detail,
		ReceivedAt: time.Now(),
	}
	_ = h.webhookRepo.Create(c.Request.Context(), evt)
}

func (h *BillingHandler) ListWebhooks(c *gin.Context) {
	email, exists := c.Get("Email")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	events, err := h.webhookRepo.ListByEmail(c.Request.Context(), email.(string), 25)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, events)
}
