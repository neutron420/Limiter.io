package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Email        string         `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string         `gorm:"not null" json:"-"`
	AvatarURL    string         `json:"avatar_url"`
	// MFA (TOTP). Secret is stored server-side only; enabled after first verify.
	TOTPSecret string         `json:"-"`
	MFAEnabled bool           `gorm:"default:false;not null" json:"mfa_enabled"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	Projects     []Project     `gorm:"foreignKey:UserID" json:"projects,omitempty"`
	Subscription *Subscription `gorm:"foreignKey:UserID" json:"subscription,omitempty"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	Token     string    `gorm:"uniqueIndex;not null" json:"token"`
	// Session/device metadata for the sessions dashboard.
	UserAgent string    `json:"user_agent"`
	ClientIP  string    `json:"client_ip"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	Revoked   bool      `gorm:"default:false;not null" json:"revoked"`
	CreatedAt time.Time `json:"created_at"`
}

func (rt *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	if rt.ID == uuid.Nil {
		rt.ID = uuid.New()
	}
	return nil
}

type Project struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      uuid.UUID      `gorm:"type:uuid;index;not null" json:"user_id"`
	Name        string         `gorm:"not null" json:"name"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	APIKeys        []APIKey        `gorm:"foreignKey:ProjectID" json:"api_keys,omitempty"`
	RateLimitRules []RateLimitRule `gorm:"foreignKey:ProjectID" json:"rate_limit_rules,omitempty"`
}

func (p *Project) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

type APIKey struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID  uuid.UUID      `gorm:"type:uuid;index;not null" json:"project_id"`
	Name       string         `gorm:"not null" json:"name"`
	KeyHash    string         `gorm:"uniqueIndex;not null" json:"-"`
	Prefix     string         `gorm:"not null" json:"prefix"`
	Scope      string         `gorm:"default:gateway-only;not null" json:"scope"` // gateway-only, read-only, admin
	ExpiresAt  *time.Time     `json:"expires_at,omitempty"`
	RevokedAt  *time.Time     `json:"revoked_at,omitempty"`
	LastUsedAt *time.Time     `json:"last_used_at,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ak *APIKey) BeforeCreate(tx *gorm.DB) error {
	if ak.ID == uuid.Nil {
		ak.ID = uuid.New()
	}
	return nil
}

type Plan struct {
	ID                     string    `gorm:"primaryKey" json:"id"` // free, pro, enterprise
	Name                   string    `gorm:"not null" json:"name"`
	MaxProjects            int       `gorm:"not null" json:"max_projects"` // -1 for unlimited
	MaxKeysPerProject      int       `gorm:"not null" json:"max_keys_per_project"` // -1 for unlimited
	AllowedAlgorithms      string    `gorm:"not null" json:"allowed_algorithms"` // comma separated: token_bucket,fixed_window,etc.
	AnalyticsRetentionDays int       `gorm:"not null" json:"analytics_retention_days"`
	RateLimitRequests      int       `gorm:"not null" json:"rate_limit_requests"`
	RateLimitPeriod        int       `gorm:"not null" json:"rate_limit_period"` // in seconds
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, j)
}

type Subscription struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	UserID          uuid.UUID      `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	PlanID          string         `gorm:"not null;index" json:"plan_id"`
	Status          string         `gorm:"not null" json:"status"` // active, expired, cancelled
	StartsAt        time.Time      `gorm:"not null" json:"starts_at"`
	ExpiresAt       *time.Time     `json:"expires_at,omitempty"`
	BillingMetadata JSONMap        `gorm:"type:jsonb" json:"billing_metadata,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	Plan Plan `gorm:"foreignKey:PlanID" json:"plan,omitempty"`
}

func (s *Subscription) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

type UpgradeHistory struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	OldPlanID string    `gorm:"not null" json:"old_plan_id"`
	NewPlanID string    `gorm:"not null" json:"new_plan_id"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

func (uh *UpgradeHistory) BeforeCreate(tx *gorm.DB) error {
	if uh.ID == uuid.Nil {
		uh.ID = uuid.New()
	}
	return nil
}

type RateLimitRule struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID    uuid.UUID      `gorm:"type:uuid;index;not null" json:"project_id"`
	Name         string         `gorm:"not null" json:"name"`
	RoutePattern string         `gorm:"not null" json:"route_pattern"` // e.g. /api/v1/users/* or *
	Algorithm    string         `gorm:"not null" json:"algorithm"`     // token_bucket, fixed_window, etc.
	// KeyStrategy decides what the limiter counter is bucketed by:
	//   api_key (default) — one bucket per API key
	//   ip                — one bucket per client IP
	KeyStrategy string         `gorm:"default:api_key;not null" json:"key_strategy"`
	Limit       int            `gorm:"not null" json:"limit"`
	Period      int            `gorm:"not null" json:"period"` // in seconds
	Burst       int            `gorm:"default:0" json:"burst"`  // used for Token Bucket/Leaky Bucket
	// Priority orders rules when several patterns match the same path — lower
	// number wins (checked first). Rules with equal priority keep list order.
	Priority int `gorm:"default:100;not null" json:"priority"`
	// CustomResponse, when set, is returned as the JSON "error" message of a
	// 429 instead of the generic default.
	CustomResponse string         `json:"custom_response"`
	IsActive       bool           `gorm:"default:true;not null" json:"is_active"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (rl *RateLimitRule) BeforeCreate(tx *gorm.DB) error {
	if rl.ID == uuid.Nil {
		rl.ID = uuid.New()
	}
	return nil
}

type AnalyticsLog struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID     uuid.UUID `gorm:"type:uuid;index;not null" json:"project_id"`
	APIKeyID      uuid.UUID `gorm:"type:uuid;index" json:"api_key_id"`
	RequestID     uuid.UUID `gorm:"type:uuid" json:"request_id"`
	ClientIP      string    `json:"client_ip"`
	Route         string    `json:"route"`
	StatusCode    int       `json:"status_code"`
	LatencyMs     int       `json:"latency_ms"`
	Decision      string    `gorm:"not null" json:"decision"` // allowed, blocked
	BlockedReason string    `json:"blocked_reason,omitempty"`
	Timestamp     time.Time `gorm:"index;not null" json:"timestamp"`
}

func (al *AnalyticsLog) BeforeCreate(tx *gorm.DB) error {
	if al.ID == uuid.Nil {
		al.ID = uuid.New()
	}
	return nil
}

// PasswordResetToken backs the forgot-password / reset flow.
type PasswordResetToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	TokenHash string     `gorm:"uniqueIndex;not null" json:"-"` // SHA-256 of the emailed token
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

func (prt *PasswordResetToken) BeforeCreate(tx *gorm.DB) error {
	if prt.ID == uuid.Nil {
		prt.ID = uuid.New()
	}
	return nil
}

// WebhookEvent records every inbound billing webhook for auditing/debugging.
type WebhookEvent struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Source     string    `gorm:"not null" json:"source"` // e.g. lemon_squeezy
	EventName  string    `json:"event_name"`
	Email      string    `json:"email"`
	Verified   bool      `json:"verified"` // HMAC signature valid?
	Status     string    `json:"status"`   // processed, ignored, rejected, error
	Detail     string    `json:"detail"`
	ReceivedAt time.Time `gorm:"index;not null" json:"received_at"`
}

func (we *WebhookEvent) BeforeCreate(tx *gorm.DB) error {
	if we.ID == uuid.Nil {
		we.ID = uuid.New()
	}
	return nil
}

// ProjectMember generalizes project access beyond the single owner (teams).
type ProjectMember struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID uuid.UUID `gorm:"type:uuid;index:idx_project_user,unique;not null" json:"project_id"`
	UserID    uuid.UUID `gorm:"type:uuid;index:idx_project_user,unique;not null" json:"user_id"`
	Role      string    `gorm:"not null;default:member" json:"role"` // owner, admin, member
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func (pm *ProjectMember) BeforeCreate(tx *gorm.DB) error {
	if pm.ID == uuid.Nil {
		pm.ID = uuid.New()
	}
	return nil
}

// ProjectInvite represents a pending invitation to join a project.
type ProjectInvite struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID  uuid.UUID  `gorm:"type:uuid;index;not null" json:"project_id"`
	Email      string     `gorm:"index;not null" json:"email"`           // invitee (may not be a user yet)
	Role       string     `gorm:"not null;default:member" json:"role"`   // admin | member
	TokenHash  string     `gorm:"uniqueIndex;not null" json:"-"`         // SHA-256 of the emailed token
	InvitedBy  uuid.UUID  `gorm:"type:uuid;not null" json:"invited_by"`
	Status     string     `gorm:"not null;default:pending" json:"status"` // pending | accepted | revoked | expired
	ExpiresAt  time.Time  `gorm:"not null" json:"expires_at"`             // 7 days
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (pi *ProjectInvite) BeforeCreate(tx *gorm.DB) error {
	if pi.ID == uuid.Nil {
		pi.ID = uuid.New()
	}
	return nil
}

// AlertRule defines a project alert condition evaluated periodically.
type AlertRule struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID uuid.UUID `gorm:"type:uuid;index;not null" json:"project_id"`
	Name      string    `gorm:"not null" json:"name"`
	// Metric: block_rate (%), traffic_spike (req/window), avg_latency_ms
	Metric string `gorm:"not null" json:"metric"`
	// Threshold the metric must exceed to fire.
	Threshold float64 `gorm:"not null" json:"threshold"`
	// WindowMinutes the evaluator looks back over.
	WindowMinutes int `gorm:"default:5;not null" json:"window_minutes"`
	// Channel: email | webhook
	Channel string `gorm:"not null;default:email" json:"channel"`
	// Target: email address or webhook URL, depending on channel.
	Target      string     `gorm:"not null" json:"target"`
	IsActive    bool       `gorm:"default:true;not null;index" json:"is_active"`
	LastFiredAt *time.Time `json:"last_fired_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (ar *AlertRule) BeforeCreate(tx *gorm.DB) error {
	if ar.ID == uuid.Nil {
		ar.ID = uuid.New()
	}
	return nil
}

// AlertEvent records each time an alert fired (history for the UI).
type AlertEvent struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	RuleID    uuid.UUID `gorm:"type:uuid;index;not null" json:"rule_id"`
	ProjectID uuid.UUID `gorm:"type:uuid;index;not null" json:"project_id"`
	Metric    string    `json:"metric"`
	Value     float64   `json:"value"`
	Threshold float64   `json:"threshold"`
	Message   string    `json:"message"`
	Delivered bool      `json:"delivered"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

func (ae *AlertEvent) BeforeCreate(tx *gorm.DB) error {
	if ae.ID == uuid.Nil {
		ae.ID = uuid.New()
	}
	return nil
}

// IPAccessRule is a project-level allow/deny entry checked before rate limiting.
// Value is an exact IP or CIDR (e.g. 1.2.3.4 or 10.0.0.0/8).
type IPAccessRule struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID uuid.UUID `gorm:"type:uuid;index;not null" json:"project_id"`
	Action    string    `gorm:"not null" json:"action"` // allow | deny
	Value     string    `gorm:"not null" json:"value"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

func (ip *IPAccessRule) BeforeCreate(tx *gorm.DB) error {
	if ip.ID == uuid.Nil {
		ip.ID = uuid.New()
	}
	return nil
}

type RuleVersion struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	RuleID    uuid.UUID      `gorm:"type:uuid;index;not null" json:"rule_id"`
	ProjectID uuid.UUID      `gorm:"type:uuid;index;not null" json:"project_id"`
	Version   int            `gorm:"not null" json:"version"`
	Snapshot  JSONMap        `gorm:"type:jsonb;not null" json:"snapshot"`
	CreatedBy uuid.UUID      `gorm:"type:uuid" json:"created_by"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (rv *RuleVersion) BeforeCreate(tx *gorm.DB) error {
	if rv.ID == uuid.Nil {
		rv.ID = uuid.New()
	}
	return nil
}
