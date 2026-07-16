package repository

import (
	"context"

	"limiter.io/internal/models"

	"github.com/google/uuid"
)

type ProjectAuditRepository interface {
	Create(ctx context.Context, event *models.ProjectAuditEvent) error
	ListByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]models.ProjectAuditEvent, error)
}
