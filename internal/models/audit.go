package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectAuditEvent struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID  uuid.UUID `gorm:"type:uuid;index;not null" json:"project_id"`
	ActorID    uuid.UUID `gorm:"type:uuid;index;not null" json:"actor_id"`
	Action     string    `gorm:"not null;index" json:"action"`
	TargetType string    `gorm:"not null" json:"target_type"`
	TargetID   uuid.UUID `gorm:"type:uuid;index;not null" json:"target_id"`
	Metadata   JSONMap   `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt  time.Time `gorm:"index;not null" json:"created_at"`
}

func (e *ProjectAuditEvent) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}
