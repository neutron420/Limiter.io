package repository

import (
	"context"
	"time"

	"limiter.io/internal/models"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, rt *models.RefreshToken) error
	GetByToken(ctx context.Context, token string) (*models.RefreshToken, error)
	RevokeByUserID(ctx context.Context, userID uuid.UUID) error
}

type ProjectRepository interface {
	Create(ctx context.Context, project *models.Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Project, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]models.Project, error)
	ListByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Project, error)
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type APIKeyRepository interface {
	Create(ctx context.Context, apiKey *models.APIKey) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.APIKey, error)
	GetByKeyHash(ctx context.Context, hash string) (*models.APIKey, error)
	ListByProjectID(ctx context.Context, projectID uuid.UUID) ([]models.APIKey, error)
	CountByProjectID(ctx context.Context, projectID uuid.UUID) (int64, error)
	Update(ctx context.Context, apiKey *models.APIKey) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type RateLimitRuleRepository interface {
	Create(ctx context.Context, rule *models.RateLimitRule) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.RateLimitRule, error)
	ListByProjectID(ctx context.Context, projectID uuid.UUID) ([]models.RateLimitRule, error)
	Update(ctx context.Context, rule *models.RateLimitRule) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *models.Subscription) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Subscription, error)
	Update(ctx context.Context, sub *models.Subscription) error
	CreateUpgradeHistory(ctx context.Context, history *models.UpgradeHistory) error
	GetPlanByID(ctx context.Context, planID string) (*models.Plan, error)
}

type AnalyticsRepository interface {
	GetAggregatedStats(ctx context.Context, projectID uuid.UUID, start, end time.Time) (map[string]interface{}, error)
	GetLogs(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]models.AnalyticsLog, error)
	// GetTimeSeries returns per-bucket allowed/blocked counts for charting.
	GetTimeSeries(ctx context.Context, projectID uuid.UUID, start, end time.Time, bucket string) ([]map[string]interface{}, error)
	// CountRequestsByProjects counts total requests across projects in a window (usage metering).
	CountRequestsByProjects(ctx context.Context, projectIDs []uuid.UUID, start, end time.Time) (int64, error)
}

type PasswordResetTokenRepository interface {
	Create(ctx context.Context, t *models.PasswordResetToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*models.PasswordResetToken, error)
	MarkUsed(ctx context.Context, id uuid.UUID) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

type WebhookEventRepository interface {
	Create(ctx context.Context, e *models.WebhookEvent) error
	ListRecent(ctx context.Context, limit int) ([]models.WebhookEvent, error)
	ListByEmail(ctx context.Context, email string, limit int) ([]models.WebhookEvent, error)
}

type ProjectMemberRepository interface {
	Add(ctx context.Context, m *models.ProjectMember) error
	Remove(ctx context.Context, projectID, userID uuid.UUID) error
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]models.ProjectMember, error)
	ListProjectIDsByUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	IsMember(ctx context.Context, projectID, userID uuid.UUID) (bool, error)
}

type CacheRepository interface {
	GetAPIKey(ctx context.Context, hash string) (*models.APIKey, error)
	SetAPIKey(ctx context.Context, hash string, apiKey *models.APIKey, ttl time.Duration) error
	DeleteAPIKey(ctx context.Context, hash string) error

	GetSubscription(ctx context.Context, userID uuid.UUID) (*models.Subscription, error)
	SetSubscription(ctx context.Context, userID uuid.UUID, sub *models.Subscription, ttl time.Duration) error
	DeleteSubscription(ctx context.Context, userID uuid.UUID) error
}
