package models

import "time"

type UsageRecord struct {
	ID          string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	ProjectID   string    `gorm:"type:uuid;not null;index" json:"project_id"`
	RequestCount int64    `gorm:"not null;default:0" json:"request_count"`
	BlockedCount int64    `gorm:"not null;default:0" json:"blocked_count"`
	PeriodStart time.Time `gorm:"not null" json:"period_start"`
	PeriodEnd   time.Time `gorm:"not null" json:"period_end"`
	Tier        string    `gorm:"type:varchar(50)" json:"tier"`
	CreatedAt   time.Time `json:"created_at"`
}

type Invoice struct {
	ID            string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	OrganizationID string   `gorm:"type:uuid;index" json:"organization_id"`
	ProjectID     string    `gorm:"type:uuid;index" json:"project_id"`
	Amount        int64     `gorm:"not null" json:"amount"`
	Currency      string    `gorm:"type:varchar(3);default:'USD'" json:"currency"`
	Status        string    `gorm:"type:varchar(50);default:'pending'" json:"status"`
	PeriodStart   time.Time `json:"period_start"`
	PeriodEnd     time.Time `json:"period_end"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`
	StripeInvoiceID string  `gorm:"type:varchar(255)" json:"stripe_invoice_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type SLAConfig struct {
	ID              string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	OrganizationID  string    `gorm:"type:uuid;not null;index" json:"organization_id"`
	UptimeSLA       float64   `gorm:"default:99.9" json:"uptime_sla"`
	ResponseTimeP99 int       `gorm:"default:100" json:"response_time_p99_ms"`
	SupportLevel    string    `gorm:"type:varchar(50);default:'standard'" json:"support_level"`
	SupportContact  string    `gorm:"type:text" json:"support_contact,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type EmailTemplate struct {
	ID            string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	OrganizationID string   `gorm:"type:uuid;not null;index" json:"organization_id"`
	Name          string    `gorm:"type:varchar(255);not null" json:"name"`
	Subject       string    `gorm:"type:text;not null" json:"subject"`
	HTMLBody      string    `gorm:"type:text;not null" json:"html_body"`
	IsActive      bool      `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type RegionConfig struct {
	ID             string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	OrganizationID string    `gorm:"type:uuid;not null;index" json:"organization_id"`
	Region         string    `gorm:"type:varchar(100);not null" json:"region"`
	GatewayURL     string    `gorm:"type:text" json:"gateway_url"`
	DataResidency  bool      `gorm:"default:true" json:"data_residency"`
	Enabled        bool      `gorm:"default:true" json:"enabled"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (*UsageRecord) TableName() string  { return "usage_records" }
func (*Invoice) TableName() string       { return "invoices" }
func (*SLAConfig) TableName() string     { return "sla_configs" }
func (*EmailTemplate) TableName() string { return "email_templates" }
func (*RegionConfig) TableName() string  { return "region_configs" }
