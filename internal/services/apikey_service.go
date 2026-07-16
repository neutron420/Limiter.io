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
	memberRepo  repository.ProjectMemberRepository
	auditRepo   repository.ProjectAuditRepository
}

func NewAPIKeyService(
	apiKeyRepo repository.APIKeyRepository,
	projectRepo repository.ProjectRepository,
	subRepo repository.SubscriptionRepository,
	cacheRepo repository.CacheRepository,
	memberRepo repository.ProjectMemberRepository,
	auditRepo repository.ProjectAuditRepository,
) APIKeyService {
	return &apiKeyService{
		apiKeyRepo:  apiKeyRepo,
		projectRepo: projectRepo,
		subRepo:     subRepo,
		cacheRepo:   cacheRepo,
		memberRepo:  memberRepo,
		auditRepo:   auditRepo,
	}
}

func (s *apiKeyService) recordAudit(ctx context.Context, projectID, actorID uuid.UUID, action, targetType string, targetID uuid.UUID, metadata models.JSONMap) {
	if s.auditRepo == nil {
		return
	}
	_ = s.auditRepo.Create(ctx, &models.ProjectAuditEvent{
		ProjectID: projectID, ActorID: actorID, Action: action,
		TargetType: targetType, TargetID: targetID, Metadata: metadata,
	})
}

func (s *apiKeyService) checkProjectAccess(ctx context.Context, userID, projectID uuid.UUID) error {
	role := roleForProject(ctx, s.projectRepo, s.memberRepo, userID, projectID)
	if !canRead(role) {
		return errors.New("unauthorized to access this project")
	}
	return nil
}

func (s *apiKeyService) checkProjectWriteAccess(ctx context.Context, userID, projectID uuid.UUID) error {
	role := roleForProject(ctx, s.projectRepo, s.memberRepo, userID, projectID)
	if !canWrite(role) {
		return errors.New("insufficient role: read-only members cannot modify the project")
	}
	return nil
}

func (s *apiKeyService) CreateAPIKey(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, req dto.CreateAPIKeyRequest) (*models.APIKey, string, error) {
	proj, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, "", errors.New("project not found")
	}
	if err := s.checkProjectWriteAccess(ctx, userID, projectID); err != nil {
		return nil, "", err
	}

	// Retrieve active subscription limits for the project owner
	sub, err := s.subRepo.GetByUserID(ctx, proj.UserID)
	if err != nil {
		return nil, "", errors.New("subscription not found for project owner")
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

	prefix := plainKey[:12]

	scope := req.Scope
	if scope == "" {
		scope = "gateway-only"
	}

	apiKey := &models.APIKey{
		ID:        uuid.New(),
		ProjectID: projectID,
		Name:      req.Name,
		KeyHash:   hashedKey,
		Prefix:    prefix,
		Scope:     scope,
		ExpiresAt: req.ExpiresAt,
	}

	if err := s.apiKeyRepo.Create(ctx, apiKey); err != nil {
		return nil, "", err
	}

	s.recordAudit(ctx, projectID, userID, "apikey.created", "apikey", apiKey.ID, models.JSONMap{"name": apiKey.Name, "prefix": apiKey.Prefix, "scope": apiKey.Scope})
	return apiKey, plainKey, nil
}

func (s *apiKeyService) ListAPIKeys(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) ([]models.APIKey, error) {
	if err := s.checkProjectAccess(ctx, userID, projectID); err != nil {
		return nil, err
	}

	return s.apiKeyRepo.ListByProjectID(ctx, projectID)
}

func (s *apiKeyService) RotateAPIKey(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, keyID uuid.UUID) (*models.APIKey, string, error) {
	if err := s.checkProjectWriteAccess(ctx, userID, projectID); err != nil {
		return nil, "", err
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

	s.recordAudit(ctx, projectID, userID, "apikey.rotated", "apikey", keyID, models.JSONMap{"name": oldKey.Name})
	return oldKey, plainKey, nil
}

func (s *apiKeyService) RevokeAPIKey(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, keyID uuid.UUID) error {
	if err := s.checkProjectWriteAccess(ctx, userID, projectID); err != nil {
		return err
	}

	key, err := s.apiKeyRepo.GetByID(ctx, keyID)
	if err != nil {
		return errors.New("api key not found")
	}

	now := time.Now()
	key.RevokedAt = &now

	// Invalidate cache
	s.cacheRepo.DeleteAPIKey(ctx, key.KeyHash)

	err = s.apiKeyRepo.Update(ctx, key)
	if err == nil {
		s.recordAudit(ctx, projectID, userID, "apikey.revoked", "apikey", keyID, models.JSONMap{"name": key.Name})
	}
	return err
}

func (s *apiKeyService) DeleteAPIKey(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, keyID uuid.UUID) error {
	if err := s.checkProjectWriteAccess(ctx, userID, projectID); err != nil {
		return err
	}

	key, err := s.apiKeyRepo.GetByID(ctx, keyID)
	if err != nil {
		return errors.New("api key not found")
	}

	// Invalidate cache
	s.cacheRepo.DeleteAPIKey(ctx, key.KeyHash)

	err = s.apiKeyRepo.Delete(ctx, keyID)
	if err == nil {
		s.recordAudit(ctx, projectID, userID, "apikey.deleted", "apikey", keyID, models.JSONMap{"name": key.Name})
	}
	return err
}
