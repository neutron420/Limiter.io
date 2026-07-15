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
	GetTimeSeries(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, duration time.Duration, bucket string) ([]map[string]interface{}, error)
}

type analyticsService struct {
	analyticsRepo repository.AnalyticsRepository
	projectRepo   repository.ProjectRepository
	memberRepo    repository.ProjectMemberRepository
}

func NewAnalyticsService(
	analyticsRepo repository.AnalyticsRepository,
	projectRepo repository.ProjectRepository,
	memberRepo repository.ProjectMemberRepository,
) AnalyticsService {
	return &analyticsService{
		analyticsRepo: analyticsRepo,
		projectRepo:   projectRepo,
		memberRepo:    memberRepo,
	}
}

func (s *analyticsService) checkProjectAccess(ctx context.Context, userID, projectID uuid.UUID) error {
	role := roleForProject(ctx, s.projectRepo, s.memberRepo, userID, projectID)
	if !canRead(role) {
		return errors.New("unauthorized to access this project's analytics")
	}
	return nil
}

func (s *analyticsService) GetProjectStats(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, duration time.Duration) (map[string]interface{}, error) {
	if err := s.checkProjectAccess(ctx, userID, projectID); err != nil {
		return nil, err
	}

	end := time.Now()
	start := end.Add(-duration)

	return s.analyticsRepo.GetAggregatedStats(ctx, projectID, start, end)
}

func (s *analyticsService) GetProjectLogs(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, limit, offset int) ([]models.AnalyticsLog, error) {
	if err := s.checkProjectAccess(ctx, userID, projectID); err != nil {
		return nil, err
	}

	if limit <= 0 || limit > 100 {
		limit = 50 // default/max limit protection
	}
	if offset < 0 {
		offset = 0
	}

	return s.analyticsRepo.GetLogs(ctx, projectID, limit, offset)
}

func (s *analyticsService) GetTimeSeries(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, duration time.Duration, bucket string) ([]map[string]interface{}, error) {
	if err := s.checkProjectAccess(ctx, userID, projectID); err != nil {
		return nil, err
	}

	end := time.Now()
	start := end.Add(-duration)

	return s.analyticsRepo.GetTimeSeries(ctx, projectID, start, end, bucket)
}
