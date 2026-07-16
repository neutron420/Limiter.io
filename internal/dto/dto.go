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

type GoogleLoginRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	UserEmail    string    `json:"email"`
	UserID       uuid.UUID `json:"user_id"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	// MFARequired signals the client to collect a TOTP code and call /auth/login/mfa.
	MFARequired bool `json:"mfa_required,omitempty"`
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
	Scope     string     `json:"scope" binding:"omitempty,oneof=gateway-only read-only admin"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type APIKeyResponse struct {
	ID         uuid.UUID  `json:"id"`
	ProjectID  uuid.UUID  `json:"project_id"`
	Name       string     `json:"name"`
	Prefix     string     `json:"prefix"`
	Scope      string     `json:"scope"`
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
	KeyStrategy    string `json:"key_strategy" binding:"omitempty"` // api_key, ip, header:<name>
	Limit          int    `json:"limit" binding:"required,gt=0"`
	Period         int    `json:"period" binding:"required,gt=0"` // in seconds
	Burst          int    `json:"burst" binding:"omitempty,gte=0"`
	Priority       int    `json:"priority" binding:"omitempty,gte=0,lte=1000"` // lower = matched first
	CustomResponse string `json:"custom_response" binding:"omitempty,max=500"` // custom 429 message
}

type RuleResponse struct {
	ID           uuid.UUID `json:"id"`
	ProjectID    uuid.UUID `json:"project_id"`
	Name         string    `json:"name"`
	RoutePattern string    `json:"route_pattern"`
	Algorithm    string    `json:"algorithm"`
	KeyStrategy    string    `json:"key_strategy"`
	Limit          int       `json:"limit"`
	Period         int       `json:"period"`
	Burst          int       `json:"burst"`
	Priority       int       `json:"priority"`
	CustomResponse string    `json:"custom_response"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type UpdateRuleRequest struct {
	Name         *string `json:"name" binding:"omitempty,min=3,max=100"`
	RoutePattern *string `json:"route_pattern" binding:"omitempty"`
	Algorithm    *string `json:"algorithm" binding:"omitempty,oneof=token_bucket fixed_window sliding_window_counter sliding_window_log leaky_bucket"`
	KeyStrategy    *string `json:"key_strategy" binding:"omitempty"`
	Limit          *int    `json:"limit" binding:"omitempty,gt=0"`
	Period         *int    `json:"period" binding:"omitempty,gt=0"`
	Burst          *int    `json:"burst" binding:"omitempty,gte=0"`
	Priority       *int    `json:"priority" binding:"omitempty,gte=0,lte=1000"`
	CustomResponse *string `json:"custom_response" binding:"omitempty,max=500"`
	IsActive       *bool   `json:"is_active" binding:"omitempty"`
}

// Subscription DTOs
type SubscriptionResponse struct {
	UserID    uuid.UUID  `json:"user_id"`
	PlanID    string     `json:"plan_id"`
	Status    string     `json:"status"`
	StartsAt  *time.Time `json:"starts_at,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type UpgradeSubscriptionRequest struct {
	PlanID string `json:"plan_id" binding:"required,oneof=free pro enterprise"`
	Reason string `json:"reason" binding:"max=200"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

type AddMemberRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=admin member"`
}

type MemberResponse struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"project_id"`
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin member"`
}

type InviteMemberRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=admin member"`
}

type InviteResponse struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"project_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type AcceptInviteRequest struct {
	Token string `json:"token" binding:"required"`
}

type AcceptInviteResponse struct {
	ProjectID   uuid.UUID `json:"project_id"`
	ProjectName string    `json:"project_name"`
	Role        string    `json:"role"`
}

type AuditEventResponse struct {
	ID         uuid.UUID              `json:"id"`
	ProjectID  uuid.UUID              `json:"project_id"`
	ActorID    uuid.UUID              `json:"actor_id"`
	Action     string                 `json:"action"`
	TargetType string                 `json:"target_type"`
	TargetID   uuid.UUID              `json:"target_id"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}

type SimulationRequest struct {
	RequestsPerSecond float64 `json:"requests_per_second" binding:"required,gt=0"`
	NumRequests       int     `json:"num_requests" binding:"required,gt=0,lte=500"`
}

type SimulationStep struct {
	RequestNumber int       `json:"request_number"`
	Timestamp     time.Time `json:"timestamp"`
	Allowed       bool      `json:"allowed"`
	Remaining     int       `json:"remaining"`
	Limit         int       `json:"limit"`
}

// Alert DTOs
type CreateAlertRequest struct {
	Name          string  `json:"name" binding:"required,min=3,max=100"`
	Metric        string  `json:"metric" binding:"required,oneof=block_rate traffic_spike avg_latency"`
	Threshold     float64 `json:"threshold" binding:"required,gt=0"`
	WindowMinutes int     `json:"window_minutes" binding:"omitempty,gt=0,lte=1440"`
	Channel       string  `json:"channel" binding:"required,oneof=email webhook slack"`
	Target        string  `json:"target" binding:"required"`
}

type UpdateAlertRequest struct {
	Name          *string  `json:"name" binding:"omitempty,min=3,max=100"`
	Threshold     *float64 `json:"threshold" binding:"omitempty,gt=0"`
	WindowMinutes *int     `json:"window_minutes" binding:"omitempty,gt=0,lte=1440"`
	Channel       *string  `json:"channel" binding:"omitempty,oneof=email webhook slack"`
	Target        *string  `json:"target" binding:"omitempty"`
	IsActive      *bool    `json:"is_active" binding:"omitempty"`
}

// MFA DTOs
type MFAVerifyRequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

type MFALoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Code     string `json:"code" binding:"required,len=6"`
}

type MFASetupResponse struct {
	Secret     string `json:"secret"`
	OTPAuthURL string `json:"otpauth_url"`
}

// Session DTOs
type SessionResponse struct {
	ID        uuid.UUID `json:"id"`
	UserAgent string    `json:"user_agent"`
	ClientIP  string    `json:"client_ip"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Current   bool      `json:"current"`
}

// IP access DTOs
type CreateIPRuleRequest struct {
	Action string `json:"action" binding:"required,oneof=allow deny"`
	Value  string `json:"value" binding:"required"`
	Note   string `json:"note" binding:"max=200"`
}
