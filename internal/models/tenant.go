package models

import "time"

type TenantConfig struct {
	ID         string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	ProjectID  string    `gorm:"type:uuid;not null;index" json:"project_id"`
	TenantID   string    `gorm:"type:varchar(255);not null;index" json:"tenant_id"`
	CustomerID string    `gorm:"type:varchar(255)" json:"customer_id"`
	MaxReq     int64     `gorm:"default:1000" json:"max_req"`
	WindowMs   int64     `gorm:"default:60000" json:"window_ms"`
	Enabled    bool      `gorm:"default:true" json:"enabled"`
	Metadata   string    `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (TenantConfig) TableName() string {
	return "tenant_configs"
}
