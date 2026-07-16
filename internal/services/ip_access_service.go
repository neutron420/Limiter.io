package services

import (
	"context"
	"errors"
	"net"

	"limiter.io/internal/dto"
	"limiter.io/internal/models"
	"limiter.io/internal/repository"

	"github.com/google/uuid"
)

type IPAccessService interface {
	List(ctx context.Context, userID, projectID uuid.UUID) ([]models.IPAccessRule, error)
	Create(ctx context.Context, userID, projectID uuid.UUID, req dto.CreateIPRuleRequest) (*models.IPAccessRule, error)
	Delete(ctx context.Context, userID, projectID, ruleID uuid.UUID) error
}

type ipAccessService struct {
	ipAccessRepo repository.IPAccessRepository
	projectRepo  repository.ProjectRepository
	memberRepo   repository.ProjectMemberRepository
}

func NewIPAccessService(
	ipAccessRepo repository.IPAccessRepository,
	projectRepo repository.ProjectRepository,
	memberRepo repository.ProjectMemberRepository,
) IPAccessService {
	return &ipAccessService{
		ipAccessRepo: ipAccessRepo,
		projectRepo:  projectRepo,
		memberRepo:   memberRepo,
	}
}

func (s *ipAccessService) checkWriteAccess(ctx context.Context, userID, projectID uuid.UUID) error {
	role := roleForProject(ctx, s.projectRepo, s.memberRepo, userID, projectID)
	if !canWrite(role) {
		return errors.New("insufficient role: only owners and admins can manage IP access rules")
	}
	return nil
}

func (s *ipAccessService) checkReadAccess(ctx context.Context, userID, projectID uuid.UUID) error {
	role := roleForProject(ctx, s.projectRepo, s.memberRepo, userID, projectID)
	if !canRead(role) {
		return errors.New("unauthorized to access this project's IP rules")
	}
	return nil
}

func (s *ipAccessService) List(ctx context.Context, userID, projectID uuid.UUID) ([]models.IPAccessRule, error) {
	if err := s.checkReadAccess(ctx, userID, projectID); err != nil {
		return nil, err
	}
	return s.ipAccessRepo.ListByProject(ctx, projectID)
}

func (s *ipAccessService) Create(ctx context.Context, userID, projectID uuid.UUID, req dto.CreateIPRuleRequest) (*models.IPAccessRule, error) {
	if err := s.checkWriteAccess(ctx, userID, projectID); err != nil {
		return nil, err
	}

	if req.Action != "allow" && req.Action != "deny" {
		return nil, errors.New("action must be 'allow' or 'deny'")
	}

	if req.Value == "" {
		return nil, errors.New("value (IP or CIDR) is required")
	}

	if _, _, err := net.ParseCIDR(req.Value); err != nil {
		if net.ParseIP(req.Value) == nil {
			return nil, errors.New("value must be a valid IP address or CIDR notation")
		}
	}

	rule := &models.IPAccessRule{
		ID:        uuid.New(),
		ProjectID: projectID,
		Action:    req.Action,
		Value:     req.Value,
		Note:      req.Note,
	}

	if err := s.ipAccessRepo.Create(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *ipAccessService) Delete(ctx context.Context, userID, projectID, ruleID uuid.UUID) error {
	if err := s.checkWriteAccess(ctx, userID, projectID); err != nil {
		return err
	}
	return s.ipAccessRepo.Delete(ctx, projectID, ruleID)
}
