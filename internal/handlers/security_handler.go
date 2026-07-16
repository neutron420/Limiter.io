package handlers

import (
	"net/http"

	"limiter.io/internal/dto"
	"limiter.io/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SecurityHandler struct {
	securityService services.SecurityService
}

func NewSecurityHandler(securityService services.SecurityService) *SecurityHandler {
	return &SecurityHandler{securityService: securityService}
}

func (h *SecurityHandler) SetupMFA(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	userID := uuid.MustParse(userIDStr.(string))
	resp, err := h.securityService.SetupMFA(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *SecurityHandler) VerifyAndEnableMFA(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	var req dto.MFAVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	userID := uuid.MustParse(userIDStr.(string))
	if err := h.securityService.VerifyAndEnableMFA(c.Request.Context(), userID, req.Code); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "MFA enabled successfully"})
}

func (h *SecurityHandler) DisableMFA(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	var req dto.MFAVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	userID := uuid.MustParse(userIDStr.(string))
	if err := h.securityService.DisableMFA(c.Request.Context(), userID, req.Code); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "MFA disabled successfully"})
}

func (h *SecurityHandler) ListSessions(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	userID := uuid.MustParse(userIDStr.(string))
	currentToken, _ := c.Get("Token")
	currentTokenStr, _ := currentToken.(string)
	sessions, err := h.securityService.ListSessions(c.Request.Context(), userID, currentTokenStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, sessions)
}

func (h *SecurityHandler) RevokeSession(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	sessionID, err := uuid.Parse(c.Param("sessionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid session ID format"})
		return
	}
	userID := uuid.MustParse(userIDStr.(string))
	if err := h.securityService.RevokeSession(c.Request.Context(), userID, sessionID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "session revoked successfully"})
}
