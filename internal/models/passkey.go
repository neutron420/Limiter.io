package models

import "time"

type Passkey struct {
	ID             string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID         string    `gorm:"type:uuid;not null;index" json:"user_id"`
	CredentialID   string    `gorm:"type:text;not null;uniqueIndex" json:"credential_id"`
	PublicKey      string    `gorm:"type:text;not null" json:"public_key"`
	AttestationType string   `gorm:"type:varchar(50)" json:"attestation_type"`
	AAGUID         string    `gorm:"type:uuid" json:"aaguid"`
	Nickname       string    `gorm:"type:varchar(255)" json:"nickname"`
	SignCount      uint32    `gorm:"default:0" json:"sign_count"`
	LastUsedAt     *time.Time `json:"last_used_at,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (Passkey) TableName() string {
	return "passkeys"
}
