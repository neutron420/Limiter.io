package dto

import (
	"time"

	"github.com/google/uuid"
)

// Auth DTOs
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	UserEmail    string    `json:"email"`
	UserID       uuid.UUID `json:"user_id"`
}

// Project DTOs
type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=100"`
	Description string `json:"description" binding:"max=500"`
}

type ProjectResponse struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// API Key DTOs
type CreateAPIKeyRequest struct {
	Name      string     `json:"name" binding:"required,min=3,max=100"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type APIKeyResponse struct {
	ID         uuid.UUID  `json:"id"`
	ProjectID  uuid.UUID  `json:"project_id"`
	Name       string     `json:"name"`
	Prefix     string     `json:"prefix"`
	PlainKey   string     `json:"key,omitempty"` // only visible on creation
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// Rate Limit Rule DTOs
type CreateRuleRequest struct {
	Name         string `json:"name" binding:"required,min=3,max=100"`
	RoutePattern string `json:"route_pattern" binding:"required"` // e.g. /users/*
	Algorithm    string `json:"algorithm" binding:"required,oneof=token_bucket fixed_window sliding_window_counter sliding_window_log leaky_bucket"`
	Limit        int    `json:"limit" binding:"required,gt=0"`
	Period       int    `json:"period" binding:"required,gt=0"` // in seconds
	Burst        int    `json:"burst" binding:"omitempty,gte=0"`
}

type RuleResponse struct {
	ID           uuid.UUID `json:"id"`
	ProjectID    uuid.UUID `json:"project_id"`
	Name         string    `json:"name"`
	RoutePattern string    `json:"route_pattern"`
	Algorithm    string    `json:"algorithm"`
	Limit        int       `json:"limit"`
	Period       int       `json:"period"`
	Burst        int       `json:"burst"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UpdateRuleRequest struct {
	Name         *string `json:"name" binding:"omitempty,min=3,max=100"`
	RoutePattern *string `json:"route_pattern" binding:"omitempty"`
	Algorithm    *string `json:"algorithm" binding:"omitempty,oneof=token_bucket fixed_window sliding_window_counter sliding_window_log leaky_bucket"`
	Limit        *int    `json:"limit" binding:"omitempty,gt=0"`
	Period       *int    `json:"period" binding:"omitempty,gt=0"`
	Burst        *int    `json:"burst" binding:"omitempty,gte=0"`
	IsActive     *bool   `json:"is_active" binding:"omitempty"`
}

// Subscription DTOs
type SubscriptionResponse struct {
	UserID    uuid.UUID  `json:"user_id"`
	PlanID    string     `json:"plan_id"`
	Status    string     `json:"status"`
	StartsAt  time.Time  `json:"starts_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type UpgradeSubscriptionRequest struct {
	PlanID string `json:"plan_id" binding:"required,oneof=free pro enterprise"`
	Reason string `json:"reason" binding:"max=200"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
