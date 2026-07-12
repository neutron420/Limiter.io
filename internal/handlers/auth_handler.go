package handlers

import (
	"net/http"

	"limiter.io/internal/dto"
	"limiter.io/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authService services.AuthService
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
// @Summary Register a user
// @Description Registers a new developer user and allocates a default free subscription
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.RegisterRequest true "User Registration Details"
// @Success 201 {object} map[string]interface{} "User registered successfully"
// @Failure 400 {object} dto.ErrorResponse "Bad Request"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	user, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "user registered successfully",
		"id":      user.ID,
		"email":   user.Email,
	})
}

// Login godoc
// @Summary Authenticate user
// @Description Logs in a developer user and issues JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.LoginRequest true "User Login Details"
// @Success 200 {object} dto.AuthResponse "Successful authentication"
// @Failure 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Refresh godoc
// @Summary Refresh session token
// @Description Generates a new JWT access token using a valid refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.RefreshTokenRequest true "Refresh Token Details"
// @Success 200 {object} dto.AuthResponse "Successful refresh"
// @Failure 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	resp, err := h.authService.Refresh(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Logout godoc
// @Summary Log out user
// @Description Revokes all active refresh tokens for the authenticated user session
// @Tags auth
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Logged out successfully"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	userID := uuid.MustParse(userIDStr.(string))
	if err := h.authService.Logout(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

// ChangePassword godoc
// @Summary Change account password
// @Description Updates the authenticated user's password
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body dto.ChangePasswordRequest true "Password Update Details"
// @Success 200 {object} map[string]interface{} "Password updated successfully"
// @Failure 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	userID := uuid.MustParse(userIDStr.(string))
	if err := h.authService.ChangePassword(c.Request.Context(), userID, req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password updated successfully"})
}

// ForgotPassword godoc
// @Summary Request password reset
// @Description Initiates a password reset flow (returns HTTP 200 for security to prevent user enumeration)
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.ForgotPasswordRequest true "Password Reset Email"
// @Success 200 {object} map[string]interface{} "Reset link dispatched"
// @Failure 400 {object} dto.ErrorResponse "Bad Request"
// @Router /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	// This is a secure architectural endpoint:
	// Always returns 200 to prevent email enumeration,
	// but schedules email reset flow in the background.
	_ = h.authService.ForgotPassword(c.Request.Context(), req)

	c.JSON(http.StatusOK, gin.H{"message": "if the email exists, a password reset instructions has been sent"})
}
