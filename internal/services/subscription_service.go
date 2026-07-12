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
}

type subscriptionService struct {
	subRepo   repository.SubscriptionRepository
	cacheRepo repository.CacheRepository
}

func NewSubscriptionService(subRepo repository.SubscriptionRepository, cacheRepo repository.CacheRepository) SubscriptionService {
	return &subscriptionService{
		subRepo:   subRepo,
		cacheRepo: cacheRepo,
	}
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
