package handlers

import (
	"net/http"

	"limiter.io/internal/dto"
	"limiter.io/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PolicyHandler struct {
	policyService services.PolicyService
}

func NewPolicyHandler(policyService services.PolicyService) *PolicyHandler {
	return &PolicyHandler{policyService: policyService}
}

func (h *PolicyHandler) Create(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid project ID format"})
		return
	}

	var req dto.CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	userID := uuid.MustParse(userIDStr.(string))
	rule, err := h.policyService.CreateRule(c.Request.Context(), userID, projectID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.RuleResponse{
		ID:           rule.ID,
		ProjectID:    rule.ProjectID,
		Name:         rule.Name,
		RoutePattern: rule.RoutePattern,
		Algorithm:    rule.Algorithm,
		Limit:        rule.Limit,
		Period:       rule.Period,
		Burst:        rule.Burst,
		IsActive:     rule.IsActive,
		CreatedAt:    rule.CreatedAt,
		UpdatedAt:    rule.UpdatedAt,
	})
}

func (h *PolicyHandler) Get(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid project ID format"})
		return
	}

	ruleIDStr := c.Param("ruleId")
	ruleID, err := uuid.Parse(ruleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid rule ID format"})
		return
	}

	userID := uuid.MustParse(userIDStr.(string))
	rule, err := h.policyService.GetRule(c.Request.Context(), userID, projectID, ruleID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.RuleResponse{
		ID:           rule.ID,
		ProjectID:    rule.ProjectID,
		Name:         rule.Name,
		RoutePattern: rule.RoutePattern,
		Algorithm:    rule.Algorithm,
		Limit:        rule.Limit,
		Period:       rule.Period,
		Burst:        rule.Burst,
		IsActive:     rule.IsActive,
		CreatedAt:    rule.CreatedAt,
		UpdatedAt:    rule.UpdatedAt,
	})
}

func (h *PolicyHandler) List(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid project ID format"})
		return
	}

	userID := uuid.MustParse(userIDStr.(string))
	rules, err := h.policyService.ListRules(c.Request.Context(), userID, projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	resp := make([]dto.RuleResponse, len(rules))
	for i, rule := range rules {
		resp[i] = dto.RuleResponse{
			ID:           rule.ID,
			ProjectID:    rule.ProjectID,
			Name:         rule.Name,
			RoutePattern: rule.RoutePattern,
			Algorithm:    rule.Algorithm,
			Limit:        rule.Limit,
			Period:       rule.Period,
			Burst:        rule.Burst,
			IsActive:     rule.IsActive,
			CreatedAt:    rule.CreatedAt,
			UpdatedAt:    rule.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, resp)
}

func (h *PolicyHandler) Update(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid project ID format"})
		return
	}

	ruleIDStr := c.Param("ruleId")
	ruleID, err := uuid.Parse(ruleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid rule ID format"})
		return
	}

	var req dto.UpdateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	userID := uuid.MustParse(userIDStr.(string))
	rule, err := h.policyService.UpdateRule(c.Request.Context(), userID, projectID, ruleID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.RuleResponse{
		ID:           rule.ID,
		ProjectID:    rule.ProjectID,
		Name:         rule.Name,
		RoutePattern: rule.RoutePattern,
		Algorithm:    rule.Algorithm,
		Limit:        rule.Limit,
		Period:       rule.Period,
		Burst:        rule.Burst,
		IsActive:     rule.IsActive,
		CreatedAt:    rule.CreatedAt,
		UpdatedAt:    rule.UpdatedAt,
	})
}

func (h *PolicyHandler) Delete(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid project ID format"})
		return
	}

	ruleIDStr := c.Param("ruleId")
	ruleID, err := uuid.Parse(ruleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid rule ID format"})
		return
	}

	userID := uuid.MustParse(userIDStr.(string))
	err = h.policyService.DeleteRule(c.Request.Context(), userID, projectID, ruleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "rate limit rule deleted successfully"})
}
