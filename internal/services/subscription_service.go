package services

import (
	"context"
	"errors"
	"time"

	"limiter.io/internal/dto"
	"limiter.io/internal/models"
	"limiter.io/internal/repository"

	"github.com/google/uuid"
)

type SubscriptionService interface {
	GetSubscription(ctx context.Context, userID uuid.UUID) (*models.Subscription, error)
	UpgradeSubscription(ctx context.Context, userID uuid.UUID, req dto.UpgradeSubscriptionRequest) (*models.Subscription, error)
	GetUsage(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error)
}

type subscriptionService struct {
	subRepo     repository.SubscriptionRepository
	cacheRepo   repository.CacheRepository
	projectRepo repository.ProjectRepository
	analRepo    repository.AnalyticsRepository
}

func NewSubscriptionService(
	subRepo repository.SubscriptionRepository,
	cacheRepo repository.CacheRepository,
	projectRepo repository.ProjectRepository,
	analRepo repository.AnalyticsRepository,
) SubscriptionService {
	return &subscriptionService{
		subRepo:     subRepo,
		cacheRepo:   cacheRepo,
		projectRepo: projectRepo,
		analRepo:    analRepo,
	}
}

// GetUsage meters total gateway requests across the user's projects in the
// current calendar month — the basis for usage-based billing.
func (s *subscriptionService) GetUsage(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	sub, err := s.subRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	projects, err := s.projectRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(projects))
	for _, p := range projects {
		ids = append(ids, p.ID)
	}

	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var used int64
	if len(ids) > 0 {
		used, _ = s.analRepo.CountRequestsByProjects(ctx, ids, periodStart, now)
	}

	return map[string]interface{}{
		"plan_id":       sub.PlanID,
		"requests_used": used,
		"projects":      len(ids),
		"period_start":  periodStart,
		"period_end":    now,
	}, nil
}

func (s *subscriptionService) GetSubscription(ctx context.Context, userID uuid.UUID) (*models.Subscription, error) {
	// Try loading from cache first
	cachedSub, err := s.cacheRepo.GetSubscription(ctx, userID)
	if err == nil {
		return cachedSub, nil
	}

	// Cache miss: load from PostgreSQL
	sub, err := s.subRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Store in cache for subsequent calls (e.g. 1 hour TTL)
	_ = s.cacheRepo.SetSubscription(ctx, userID, sub, 1*time.Hour)

	return sub, nil
}

func (s *subscriptionService) UpgradeSubscription(ctx context.Context, userID uuid.UUID, req dto.UpgradeSubscriptionRequest) (*models.Subscription, error) {
	// Retrieve current subscription
	sub, err := s.subRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if sub.PlanID == req.PlanID {
		return nil, errors.New("already subscribed to this plan")
	}

	// Retrieve new plan details
	newPlan, err := s.subRepo.GetPlanByID(ctx, req.PlanID)
	if err != nil {
		return nil, errors.New("target plan not found")
	}

	oldPlanID := sub.PlanID

	// Update subscription
	sub.PlanID = newPlan.ID
	sub.Status = "active"
	sub.UpdatedAt = time.Now()
	// Optionally set expires_at (if billing logic were added, but here infinite/none)
	sub.ExpiresAt = nil 

	if err := s.subRepo.Update(ctx, sub); err != nil {
		return nil, err
	}

	// Create upgrade history log
	history := &models.UpgradeHistory{
		ID:        uuid.New(),
		UserID:    userID,
		OldPlanID: oldPlanID,
		NewPlanID: newPlan.ID,
		Reason:    req.Reason,
	}
	_ = s.subRepo.CreateUpgradeHistory(ctx, history)

	// Refresh cache: invalidate the old cache entry
	s.cacheRepo.DeleteSubscription(ctx, userID)

	// Prefetch and store updated subscription in cache
	updatedSub, err := s.subRepo.GetByUserID(ctx, userID)
	if err == nil {
		_ = s.cacheRepo.SetSubscription(ctx, userID, updatedSub, 1*time.Hour)
	}

	return updatedSub, nil
}
