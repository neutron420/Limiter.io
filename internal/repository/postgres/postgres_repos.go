package postgres

import (
	"context"
	"time"

	"limiter.io/internal/models"
	"limiter.io/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Refresh Token Repo
type refreshTokenRepo struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) repository.RefreshTokenRepository {
	return &refreshTokenRepo{db: db}
}

func (r *refreshTokenRepo) Create(ctx context.Context, rt *models.RefreshToken) error {
	return r.db.WithContext(ctx).Create(rt).Error
}

func (r *refreshTokenRepo) GetByToken(ctx context.Context, token string) (*models.RefreshToken, error) {
	var rt models.RefreshToken
	err := r.db.WithContext(ctx).First(&rt, "token = ?", token).Error
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *refreshTokenRepo) RevokeByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked = false", userID).
		Update("revoked", true).Error
}

// Project Repo
type projectRepo struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) repository.ProjectRepository {
	return &projectRepo{db: db}
}

func (r *projectRepo) Create(ctx context.Context, project *models.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

func (r *projectRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	var project models.Project
	err := r.db.WithContext(ctx).First(&project, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *projectRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]models.Project, error) {
	var projects []models.Project
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&projects).Error
	return projects, err
}

func (r *projectRepo) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Project{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

func (r *projectRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Project{}, "id = ?", id).Error
}

// API Key Repo
type apiKeyRepo struct {
	db *gorm.DB
}

func NewAPIKeyRepository(db *gorm.DB) repository.APIKeyRepository {
	return &apiKeyRepo{db: db}
}

func (r *apiKeyRepo) Create(ctx context.Context, apiKey *models.APIKey) error {
	return r.db.WithContext(ctx).Create(apiKey).Error
}

func (r *apiKeyRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.APIKey, error) {
	var apiKey models.APIKey
	err := r.db.WithContext(ctx).First(&apiKey, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

func (r *apiKeyRepo) GetByKeyHash(ctx context.Context, hash string) (*models.APIKey, error) {
	var apiKey models.APIKey
	err := r.db.WithContext(ctx).First(&apiKey, "key_hash = ? AND revoked_at IS NULL", hash).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

func (r *apiKeyRepo) ListByProjectID(ctx context.Context, projectID uuid.UUID) ([]models.APIKey, error) {
	var keys []models.APIKey
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&keys).Error
	return keys, err
}

func (r *apiKeyRepo) CountByProjectID(ctx context.Context, projectID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.APIKey{}).Where("project_id = ?", projectID).Count(&count).Error
	return count, err
}

func (r *apiKeyRepo) Update(ctx context.Context, apiKey *models.APIKey) error {
	return r.db.WithContext(ctx).Save(apiKey).Error
}

func (r *apiKeyRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.APIKey{}, "id = ?", id).Error
}

// Rate Limit Rule Repo
type ruleRepo struct {
	db *gorm.DB
}

func NewRateLimitRuleRepository(db *gorm.DB) repository.RateLimitRuleRepository {
	return &ruleRepo{db: db}
}

func (r *ruleRepo) Create(ctx context.Context, rule *models.RateLimitRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

func (r *ruleRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.RateLimitRule, error) {
	var rule models.RateLimitRule
	err := r.db.WithContext(ctx).First(&rule, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *ruleRepo) ListByProjectID(ctx context.Context, projectID uuid.UUID) ([]models.RateLimitRule, error) {
	var rules []models.RateLimitRule
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&rules).Error
	return rules, err
}

func (r *ruleRepo) Update(ctx context.Context, rule *models.RateLimitRule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

func (r *ruleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.RateLimitRule{}, "id = ?", id).Error
}

// Subscription Repo
type subRepo struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) repository.SubscriptionRepository {
	return &subRepo{db: db}
}

func (r *subRepo) Create(ctx context.Context, sub *models.Subscription) error {
	return r.db.WithContext(ctx).Create(sub).Error
}

func (r *subRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Subscription, error) {
	var sub models.Subscription
	err := r.db.WithContext(ctx).Preload("Plan").First(&sub, "user_id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *subRepo) Update(ctx context.Context, sub *models.Subscription) error {
	return r.db.WithContext(ctx).Save(sub).Error
}

func (r *subRepo) CreateUpgradeHistory(ctx context.Context, history *models.UpgradeHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

func (r *subRepo) GetPlanByID(ctx context.Context, planID string) (*models.Plan, error) {
	var plan models.Plan
	err := r.db.WithContext(ctx).First(&plan, "id = ?", planID).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

// Analytics Repo
type analyticsRepo struct {
	db *gorm.DB
}

func NewAnalyticsRepository(db *gorm.DB) repository.AnalyticsRepository {
	return &analyticsRepo{db: db}
}

func (r *analyticsRepo) GetAggregatedStats(ctx context.Context, projectID uuid.UUID, start, end time.Time) (map[string]interface{}, error) {
	var result struct {
		TotalRequests   int64   `gorm:"column:total_requests"`
		AllowedRequests int64   `gorm:"column:allowed_requests"`
		BlockedRequests int64   `gorm:"column:blocked_requests"`
		AvgLatency      float64 `gorm:"column:avg_latency"`
	}

	err := r.db.WithContext(ctx).Model(&models.AnalyticsLog{}).
		Select("COUNT(*) as total_requests, "+
			"COUNT(CASE WHEN decision = 'allowed' THEN 1 END) as allowed_requests, "+
			"COUNT(CASE WHEN decision = 'blocked' THEN 1 END) as blocked_requests, "+
			"COALESCE(AVG(latency_ms), 0) as avg_latency").
		Where("project_id = ? AND timestamp BETWEEN ? AND ?", projectID, start, end).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	// Fetch Top Blocked Routes
	type routeStat struct {
		Route string `json:"route"`
		Count int64  `json:"count"`
	}
	var topBlocked []routeStat
	r.db.WithContext(ctx).Model(&models.AnalyticsLog{}).
		Select("route, COUNT(*) as count").
		Where("project_id = ? AND decision = 'blocked' AND timestamp BETWEEN ? AND ?", projectID, start, end).
		Group("route").
		Order("count DESC").
		Limit(5).
		Scan(&topBlocked)

	stats := map[string]interface{}{
		"total_requests":   result.TotalRequests,
		"allowed_requests": result.AllowedRequests,
		"blocked_requests": result.BlockedRequests,
		"avg_latency_ms":   result.AvgLatency,
		"top_blocked":      topBlocked,
	}

	return stats, nil
}

func (r *analyticsRepo) GetLogs(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]models.AnalyticsLog, error) {
	var logs []models.AnalyticsLog
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}
