package services

import (
	"context"
	"errors"

	"limiter.io/internal/dto"
	"limiter.io/internal/models"
	"limiter.io/internal/repository"

	"github.com/google/uuid"
)

type ProjectService interface {
	CreateProject(ctx context.Context, userID uuid.UUID, req dto.CreateProjectRequest) (*models.Project, error)
	GetProject(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) (*models.Project, error)
	ListProjects(ctx context.Context, userID uuid.UUID) ([]models.Project, error)
	DeleteProject(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) error
}

type projectService struct {
	projectRepo repository.ProjectRepository
	subRepo     repository.SubscriptionRepository
}

func NewProjectService(projectRepo repository.ProjectRepository, subRepo repository.SubscriptionRepository) ProjectService {
	return &projectService{
		projectRepo: projectRepo,
		subRepo:     subRepo,
	}
}

func (s *projectService) CreateProject(ctx context.Context, userID uuid.UUID, req dto.CreateProjectRequest) (*models.Project, error) {
	// Retrieve active subscription details
	sub, err := s.subRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("subscription not found for user")
	}

	// Count user's current projects
	count, err := s.projectRepo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Verify plan limit (e.g. Free plan limits to 3 projects)
	if sub.Plan.MaxProjects != -1 && count >= int64(sub.Plan.MaxProjects) {
		return nil, errors.New("you have reached the project limit for your plan. Please upgrade to create more projects")
	}

	project := &models.Project{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *projectService) GetProject(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) (*models.Project, error) {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}

	if project.UserID != userID {
		return nil, errors.New("unauthorized to access this project")
	}

	return project, nil
}

func (s *projectService) ListProjects(ctx context.Context, userID uuid.UUID) ([]models.Project, error) {
	return s.projectRepo.ListByUserID(ctx, userID)
}

func (s *projectService) DeleteProject(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) error {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return errors.New("project not found")
	}

	if project.UserID != userID {
		return errors.New("unauthorized to delete this project")
	}

	return s.projectRepo.Delete(ctx, projectID)
}
