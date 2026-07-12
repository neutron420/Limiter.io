package services

import (
	"context"
	"errors"
	"time"

	"limiter.io/internal/dto"
	"limiter.io/internal/models"
	"limiter.io/internal/repository"
	"limiter.io/internal/utils"

	"github.com/google/uuid"
)

type APIKeyService interface {
	CreateAPIKey(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, req dto.CreateAPIKeyRequest) (*models.APIKey, string, error)
	ListAPIKeys(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) ([]models.APIKey, error)
	RotateAPIKey(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, keyID uuid.UUID) (*models.APIKey, string, error)
	RevokeAPIKey(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, keyID uuid.UUID) error
	DeleteAPIKey(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, keyID uuid.UUID) error
}

type apiKeyService struct {
	apiKeyRepo  repository.APIKeyRepository
	projectRepo repository.ProjectRepository
	subRepo     repository.SubscriptionRepository
	cacheRepo   repository.CacheRepository
}

func NewAPIKeyService(
	apiKeyRepo repository.APIKeyRepository,
	projectRepo repository.ProjectRepository,
	subRepo repository.SubscriptionRepository,
	cacheRepo repository.CacheRepository,
) APIKeyService {
	return &apiKeyService{
		apiKeyRepo:  apiKeyRepo,
		projectRepo: projectRepo,
		subRepo:     subRepo,
		cacheRepo:   cacheRepo,
	}
}

func (s *apiKeyService) CreateAPIKey(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, req dto.CreateAPIKeyRequest) (*models.APIKey, string, error) {
	// Verify project ownership
	proj, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, "", errors.New("project not found")
	}
	if proj.UserID != userID {
		return nil, "", errors.New("unauthorized to manage this project")
	}

	// Retrieve active subscription limits
	sub, err := s.subRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, "", errors.New("subscription not found for user")
	}

	// Count existing keys in project
	count, err := s.apiKeyRepo.CountByProjectID(ctx, projectID)
	if err != nil {
		return nil, "", err
	}

	// Verify plan limit (e.g. Free plan limits to 3 keys per project)
	if sub.Plan.MaxKeysPerProject != -1 && count >= int64(sub.Plan.MaxKeysPerProject) {
		return nil, "", errors.New("you have reached the API Key limit for this project. Please upgrade to create more API Keys")
	}

	// Generate key
	plainKey, hashedKey, err := utils.GenerateAPIKey()
	if err != nil {
		return nil, "", err
	}

	// First 12 characters are prefix (e.g. "rk_live_1234")
	prefix := plainKey[:12]

	apiKey := &models.APIKey{
		ID:        uuid.New(),
		ProjectID: projectID,
		Name:      req.Name,
		KeyHash:   hashedKey,
		Prefix:    prefix,
		ExpiresAt: req.ExpiresAt,
	}

	if err := s.apiKeyRepo.Create(ctx, apiKey); err != nil {
		return nil, "", err
	}

	return apiKey, plainKey, nil
}

func (s *apiKeyService) ListAPIKeys(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) ([]models.APIKey, error) {
	// Verify project ownership
	proj, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}
	if proj.UserID != userID {
		return nil, errors.New("unauthorized to access this project")
	}

	return s.apiKeyRepo.ListByProjectID(ctx, projectID)
}

func (s *apiKeyService) RotateAPIKey(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, keyID uuid.UUID) (*models.APIKey, string, error) {
	// Verify project ownership
	proj, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, "", errors.New("project not found")
	}
	if proj.UserID != userID {
		return nil, "", errors.New("unauthorized")
	}

	// Fetch existing key
	oldKey, err := s.apiKeyRepo.GetByID(ctx, keyID)
	if err != nil {
		return nil, "", errors.New("api key not found")
	}
	if oldKey.ProjectID != projectID {
		return nil, "", errors.New("api key does not belong to this project")
	}

	// Generate new key
	plainKey, hashedKey, err := utils.GenerateAPIKey()
	if err != nil {
		return nil, "", err
	}

	// Invalidate old key cache in Redis
	s.cacheRepo.DeleteAPIKey(ctx, oldKey.KeyHash)

	// Update existing record
	oldKey.KeyHash = hashedKey
	oldKey.Prefix = plainKey[:12]
	oldKey.UpdatedAt = time.Now()

	if err := s.apiKeyRepo.Update(ctx, oldKey); err != nil {
		return nil, "", err
	}

	return oldKey, plainKey, nil
}

func (s *apiKeyService) RevokeAPIKey(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, keyID uuid.UUID) error {
	proj, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return errors.New("project not found")
	}
	if proj.UserID != userID {
		return errors.New("unauthorized")
	}

	key, err := s.apiKeyRepo.GetByID(ctx, keyID)
	if err != nil {
		return errors.New("api key not found")
	}

	now := time.Now()
	key.RevokedAt = &now

	// Invalidate cache
	s.cacheRepo.DeleteAPIKey(ctx, key.KeyHash)

	return s.apiKeyRepo.Update(ctx, key)
}

func (s *apiKeyService) DeleteAPIKey(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, keyID uuid.UUID) error {
	proj, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return errors.New("project not found")
	}
	if proj.UserID != userID {
		return errors.New("unauthorized")
	}

	key, err := s.apiKeyRepo.GetByID(ctx, keyID)
	if err != nil {
		return errors.New("api key not found")
	}

	// Invalidate cache
	s.cacheRepo.DeleteAPIKey(ctx, key.KeyHash)

	return s.apiKeyRepo.Delete(ctx, keyID)
}
