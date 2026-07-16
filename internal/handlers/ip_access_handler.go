package handlers

import (
	"net/http"

	"limiter.io/internal/dto"
	"limiter.io/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type IPAccessHandler struct {
	ipAccessService services.IPAccessService
}

func NewIPAccessHandler(ipAccessService services.IPAccessService) *IPAccessHandler {
	return &IPAccessHandler{ipAccessService: ipAccessService}
}

func (h *IPAccessHandler) List(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid project ID format"})
		return
	}
	userID := uuid.MustParse(userIDStr.(string))
	rules, err := h.ipAccessService.List(c.Request.Context(), userID, projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, rules)
}

func (h *IPAccessHandler) Create(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid project ID format"})
		return
	}
	var req dto.CreateIPRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	userID := uuid.MustParse(userIDStr.(string))
	rule, err := h.ipAccessService.Create(c.Request.Context(), userID, projectID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, rule)
}

func (h *IPAccessHandler) Delete(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid project ID format"})
		return
	}
	ruleID, err := uuid.Parse(c.Param("ruleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid rule ID format"})
		return
	}
	userID := uuid.MustParse(userIDStr.(string))
	if err := h.ipAccessService.Delete(c.Request.Context(), userID, projectID, ruleID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "IP access rule deleted successfully"})
}
