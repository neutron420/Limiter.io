package models

import "time"

type ImmutableAuditLog struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID    string    `gorm:"type:uuid;not null;index" json:"user_id"`
	ProjectID string    `gorm:"type:uuid;index" json:"project_id,omitempty"`
	Action    string    `gorm:"type:varchar(100);not null" json:"action"`
	Resource  string    `gorm:"type:varchar(100)" json:"resource"`
	Details   string    `gorm:"type:jsonb" json:"details"`
	IPAddress string    `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent string    `gorm:"type:text" json:"user_agent"`
	Checksum  string    `gorm:"type:varchar(64);not null" json:"checksum"`
	PrevHash  string    `gorm:"type:varchar(64);not null;default:''" json:"prev_hash"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

func (ImmutableAuditLog) TableName() string {
	return "immutable_audit_logs"
}
