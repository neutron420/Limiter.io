package services

import (
	"context"
	"errors"
	"time"

	"limiter.io/internal/models"
	"limiter.io/internal/repository"

	"github.com/google/uuid"
)

type AnalyticsService interface {
	GetProjectStats(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, duration time.Duration) (map[string]interface{}, error)
	GetProjectLogs(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, limit, offset int) ([]models.AnalyticsLog, error)
}

type analyticsService struct {
	analyticsRepo repository.AnalyticsRepository
	projectRepo   repository.ProjectRepository
}

func NewAnalyticsService(analyticsRepo repository.AnalyticsRepository, projectRepo repository.ProjectRepository) AnalyticsService {
	return &analyticsService{
		analyticsRepo: analyticsRepo,
		projectRepo:   projectRepo,
	}
}

func (s *analyticsService) GetProjectStats(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, duration time.Duration) (map[string]interface{}, error) {
	// Verify project ownership
	proj, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}
	if proj.UserID != userID {
		return nil, errors.New("unauthorized to access this project's analytics")
	}

	end := time.Now()
	start := end.Add(-duration)

	return s.analyticsRepo.GetAggregatedStats(ctx, projectID, start, end)
}

func (s *analyticsService) GetProjectLogs(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, limit, offset int) ([]models.AnalyticsLog, error) {
	// Verify project ownership
	proj, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}
	if proj.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	if limit <= 0 || limit > 100 {
		limit = 50 // default/max limit protection
	}
	if offset < 0 {
		offset = 0
	}

	return s.analyticsRepo.GetLogs(ctx, projectID, limit, offset)
}
