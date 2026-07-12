package services

import (
	"context"
	"errors"
	"strings"

	"limiter.io/internal/dto"
	"limiter.io/internal/models"
	"limiter.io/internal/repository"

	"github.com/google/uuid"
)

type PolicyService interface {
	CreateRule(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, req dto.CreateRuleRequest) (*models.RateLimitRule, error)
	GetRule(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, ruleID uuid.UUID) (*models.RateLimitRule, error)
	ListRules(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) ([]models.RateLimitRule, error)
	UpdateRule(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, ruleID uuid.UUID, req dto.UpdateRuleRequest) (*models.RateLimitRule, error)
	DeleteRule(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, ruleID uuid.UUID) error
}

type policyService struct {
	ruleRepo    repository.RateLimitRuleRepository
	projectRepo repository.ProjectRepository
	subRepo     repository.SubscriptionRepository
}

func NewPolicyService(
	ruleRepo repository.RateLimitRuleRepository,
	projectRepo repository.ProjectRepository,
	subRepo repository.SubscriptionRepository,
) PolicyService {
	return &policyService{
		ruleRepo:    ruleRepo,
		projectRepo: projectRepo,
		subRepo:     subRepo,
	}
}

func (s *policyService) CreateRule(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, req dto.CreateRuleRequest) (*models.RateLimitRule, error) {
	// Verify project ownership
	proj, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}
	if proj.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	// Retrieve user subscription
	sub, err := s.subRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("subscription not found")
	}

	// Verify allowed algorithms for the subscription plan
	allowedAlgos := strings.Split(sub.Plan.AllowedAlgorithms, ",")
	algoAllowed := false
	for _, algo := range allowedAlgos {
		if strings.TrimSpace(algo) == req.Algorithm {
			algoAllowed = true
			break
		}
	}

	if !algoAllowed {
		return nil, errors.New("the selected rate-limiting algorithm is not available on your plan. Please upgrade to Pro")
	}

	rule := &models.RateLimitRule{
		ID:           uuid.New(),
		ProjectID:    projectID,
		Name:         req.Name,
		RoutePattern: req.RoutePattern,
		Algorithm:    req.Algorithm,
		Limit:        req.Limit,
		Period:       req.Period,
		Burst:        req.Burst,
		IsActive:     true,
	}

	if err := s.ruleRepo.Create(ctx, rule); err != nil {
		return nil, err
	}

	return rule, nil
}

func (s *policyService) GetRule(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, ruleID uuid.UUID) (*models.RateLimitRule, error) {
	proj, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}
	if proj.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	rule, err := s.ruleRepo.GetByID(ctx, ruleID)
	if err != nil {
		return nil, errors.New("rate limit rule not found")
	}

	if rule.ProjectID != projectID {
		return nil, errors.New("rule does not belong to this project")
	}

	return rule, nil
}

func (s *policyService) ListRules(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) ([]models.RateLimitRule, error) {
	proj, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}
	if proj.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	return s.ruleRepo.ListByProjectID(ctx, projectID)
}

func (s *policyService) UpdateRule(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, ruleID uuid.UUID, req dto.UpdateRuleRequest) (*models.RateLimitRule, error) {
	proj, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}
	if proj.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	rule, err := s.ruleRepo.GetByID(ctx, ruleID)
	if err != nil {
		return nil, errors.New("rule not found")
	}

	if rule.ProjectID != projectID {
		return nil, errors.New("rule does not belong to this project")
	}

	// Update fields if provided
	if req.Name != nil {
		rule.Name = *req.Name
	}
	if req.RoutePattern != nil {
		rule.RoutePattern = *req.RoutePattern
	}
	if req.IsActive != nil {
		rule.IsActive = *req.IsActive
	}
	if req.Limit != nil {
		rule.Limit = *req.Limit
	}
	if req.Period != nil {
		rule.Period = *req.Period
	}
	if req.Burst != nil {
		rule.Burst = *req.Burst
	}

	if req.Algorithm != nil {
		// Verify allowed algorithms for the subscription plan
		sub, err := s.subRepo.GetByUserID(ctx, userID)
		if err != nil {
			return nil, errors.New("subscription not found")
		}

		allowedAlgos := strings.Split(sub.Plan.AllowedAlgorithms, ",")
		algoAllowed := false
		for _, algo := range allowedAlgos {
			if strings.TrimSpace(algo) == *req.Algorithm {
				algoAllowed = true
				break
			}
		}

		if !algoAllowed {
			return nil, errors.New("the selected rate-limiting algorithm is not available on your plan. Please upgrade to Pro")
		}
		rule.Algorithm = *req.Algorithm
	}

	if err := s.ruleRepo.Update(ctx, rule); err != nil {
		return nil, err
	}

	return rule, nil
}

func (s *policyService) DeleteRule(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, ruleID uuid.UUID) error {
	proj, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return errors.New("project not found")
	}
	if proj.UserID != userID {
		return errors.New("unauthorized")
	}

	rule, err := s.ruleRepo.GetByID(ctx, ruleID)
	if err != nil {
		return errors.New("rule not found")
	}

	if rule.ProjectID != projectID {
		return errors.New("rule does not belong to this project")
	}

	return s.ruleRepo.Delete(ctx, ruleID)
}
