package postgres

import (
	"context"

	"limiter.io/internal/models"
	"limiter.io/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type projectAuditRepo struct {
	db *gorm.DB
}

func NewProjectAuditRepository(db *gorm.DB) repository.ProjectAuditRepository {
	return &projectAuditRepo{db: db}
}

func (r *projectAuditRepo) Create(ctx context.Context, event *models.ProjectAuditEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *projectAuditRepo) ListByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]models.ProjectAuditEvent, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	var events []models.ProjectAuditEvent
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error
	return events, err
}
