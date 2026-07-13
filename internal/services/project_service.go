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
	AddMember(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, req dto.AddMemberRequest) (*models.ProjectMember, error)
	RemoveMember(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, memberID uuid.UUID) error
	ListMembers(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) ([]models.ProjectMember, error)
}

type projectService struct {
	projectRepo repository.ProjectRepository
	subRepo     repository.SubscriptionRepository
	memberRepo  repository.ProjectMemberRepository
	userRepo    repository.UserRepository
}

func NewProjectService(
	projectRepo repository.ProjectRepository,
	subRepo repository.SubscriptionRepository,
	memberRepo repository.ProjectMemberRepository,
	userRepo repository.UserRepository,
) ProjectService {
	return &projectService{
		projectRepo: projectRepo,
		subRepo:     subRepo,
		memberRepo:  memberRepo,
		userRepo:    userRepo,
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

	isOwner := project.UserID == userID
	isMember := false
	if !isOwner {
		isMember, _ = s.memberRepo.IsMember(ctx, projectID, userID)
	}

	if !isOwner && !isMember {
		return nil, errors.New("unauthorized to access this project")
	}

	return project, nil
}

func (s *projectService) ListProjects(ctx context.Context, userID uuid.UUID) ([]models.Project, error) {
	owned, err := s.projectRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	memberProjectIDs, err := s.memberRepo.ListProjectIDsByUser(ctx, userID)
	if err != nil || len(memberProjectIDs) == 0 {
		return owned, nil
	}

	memberProjects, err := s.projectRepo.ListByIDs(ctx, memberProjectIDs)
	if err != nil {
		return owned, nil
	}

	return append(owned, memberProjects...), nil
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

func (s *projectService) AddMember(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, req dto.AddMemberRequest) (*models.ProjectMember, error) {
	proj, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}
	if proj.UserID != userID {
		return nil, errors.New("only the project owner can manage team members")
	}

	targetUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("user with this email not found")
	}

	if targetUser.ID == userID {
		return nil, errors.New("cannot invite the owner of the project")
	}

	isMem, _ := s.memberRepo.IsMember(ctx, projectID, targetUser.ID)
	if isMem {
		return nil, errors.New("user is already a member of this project")
	}

	member := &models.ProjectMember{
		ID:        uuid.New(),
		ProjectID: projectID,
		UserID:    targetUser.ID,
		Email:     targetUser.Email,
		Role:      req.Role,
	}

	if err := s.memberRepo.Add(ctx, member); err != nil {
		return nil, err
	}

	return member, nil
}

func (s *projectService) RemoveMember(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, memberID uuid.UUID) error {
	proj, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return errors.New("project not found")
	}
	if proj.UserID != userID {
		return errors.New("only the project owner can manage team members")
	}

	return s.memberRepo.Remove(ctx, projectID, memberID)
}

func (s *projectService) ListMembers(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) ([]models.ProjectMember, error) {
	proj, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}

	isOwner := proj.UserID == userID
	isMem, _ := s.memberRepo.IsMember(ctx, projectID, userID)
	if !isOwner && !isMem {
		return nil, errors.New("unauthorized to view members for this project")
	}

	return s.memberRepo.ListByProject(ctx, projectID)
}
