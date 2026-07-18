package models

import "time"

type Organization struct {
	ID          string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Slug        string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"slug"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	OwnerID     string    `gorm:"type:uuid;not null;index" json:"owner_id"`
	Plan        string    `gorm:"type:varchar(50);default:'free'" json:"plan"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type OrganizationMember struct {
	ID             string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	OrganizationID string    `gorm:"type:uuid;not null;index" json:"organization_id"`
	UserID         string    `gorm:"type:uuid;not null;index" json:"user_id"`
	Role           string    `gorm:"type:varchar(50);default:'member'" json:"role"`
	JoinedAt       time.Time `json:"joined_at"`
}

type OrganizationGroup struct {
	ID             string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	OrganizationID string    `gorm:"type:uuid;not null;index" json:"organization_id"`
	Name           string    `gorm:"type:varchar(255);not null" json:"name"`
	Description    string    `gorm:"type:text" json:"description,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type OrganizationGroupMember struct {
	ID      string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	GroupID string `gorm:"type:uuid;not null;index" json:"group_id"`
	UserID  string `gorm:"type:uuid;not null;index" json:"user_id"`
}

type ApprovalWorkflow struct {
	ID             string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	OrganizationID string    `gorm:"type:uuid;not null;index" json:"organization_id"`
	Name           string    `gorm:"type:varchar(255);not null" json:"name"`
	ActionType     string    `gorm:"type:varchar(100);not null" json:"action_type"`
	MinApprovers   int       `gorm:"default:1" json:"min_approvers"`
	ApproverGroup  string    `gorm:"type:varchar(100)" json:"approver_group"`
	Enabled        bool      `gorm:"default:true" json:"enabled"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ApprovalRequest struct {
	ID              string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	WorkflowID      string    `gorm:"type:uuid;not null;index" json:"workflow_id"`
	RequestedBy     string    `gorm:"type:uuid;not null" json:"requested_by"`
	OrganizationID  string    `gorm:"type:uuid;not null;index" json:"organization_id"`
	Status          string    `gorm:"type:varchar(50);default:'pending'" json:"status"`
	TargetType      string    `gorm:"type:varchar(100)" json:"target_type"`
	TargetID        string    `gorm:"type:uuid" json:"target_id"`
	Reason          string    `gorm:"type:text" json:"reason"`
	ApprovedBy      []string  `gorm:"type:jsonb" json:"approved_by,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type SAMLConfig struct {
	ID              string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	OrganizationID  string    `gorm:"type:uuid;not null;uniqueIndex" json:"organization_id"`
	IDPEntityID     string    `gorm:"type:text" json:"idp_entity_id"`
	IDPSSOURL       string    `gorm:"type:text" json:"idp_sso_url"`
	IDPSSOBinding   string    `gorm:"type:varchar(50)" json:"idp_sso_binding"`
	IDPPublicCert   string    `gorm:"type:text" json:"idp_public_cert"`
	SPEntityID      string    `gorm:"type:text" json:"sp_entity_id"`
	SPACSURL        string    `gorm:"type:text" json:"sp_acs_url"`
	SPPrivateKey    string    `gorm:"type:text" json:"sp_private_key"`
	SPPublicCert    string    `gorm:"type:text" json:"sp_public_cert"`
	Enabled         bool      `gorm:"default:false" json:"enabled"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type OIDCConfig struct {
	ID              string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	OrganizationID  string    `gorm:"type:uuid;not null;uniqueIndex" json:"organization_id"`
	IssuerURL       string    `gorm:"type:text" json:"issuer_url"`
	ClientID        string    `gorm:"type:text" json:"client_id"`
	ClientSecret    string    `gorm:"type:text" json:"client_secret"`
	RedirectURL     string    `gorm:"type:text" json:"redirect_url"`
	Scopes          string    `gorm:"type:varchar(255)" json:"scopes"`
	Enabled         bool      `gorm:"default:false" json:"enabled"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (*Organization) TableName() string             { return "organizations" }
func (*OrganizationMember) TableName() string        { return "organization_members" }
func (*OrganizationGroup) TableName() string         { return "organization_groups" }
func (*OrganizationGroupMember) TableName() string   { return "organization_group_members" }
func (*ApprovalWorkflow) TableName() string          { return "approval_workflows" }
func (*ApprovalRequest) TableName() string           { return "approval_requests" }
func (*SAMLConfig) TableName() string                { return "saml_configs" }
func (*OIDCConfig) TableName() string                { return "oidc_configs" }
