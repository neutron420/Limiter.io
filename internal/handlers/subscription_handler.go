package handlers

import (
	"net/http"

	"limiter.io/internal/dto"
	"limiter.io/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SubscriptionHandler struct {
	subService services.SubscriptionService
}

func NewSubscriptionHandler(subService services.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{subService: subService}
}

func (h *SubscriptionHandler) Get(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	userID := uuid.MustParse(userIDStr.(string))
	sub, err := h.subService.GetSubscription(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.SubscriptionResponse{
		UserID:    sub.UserID,
		PlanID:    sub.PlanID,
		Status:    sub.Status,
		StartsAt:  sub.StartsAt,
		ExpiresAt: sub.ExpiresAt,
	})
}

func (h *SubscriptionHandler) Upgrade(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	var req dto.UpgradeSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	userID := uuid.MustParse(userIDStr.(string))
	sub, err := h.subService.UpgradeSubscription(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.SubscriptionResponse{
		UserID:    sub.UserID,
		PlanID:    sub.PlanID,
		Status:    sub.Status,
		StartsAt:  sub.StartsAt,
		ExpiresAt: sub.ExpiresAt,
	})
}
