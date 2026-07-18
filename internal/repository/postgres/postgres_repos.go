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

type ssoRepo struct {
	db *gorm.DB
}

func (r *ssoRepo) GetSAMLConfig(ctx context.Context, orgID string) (*models.SAMLConfig, error) {
	var cfg models.SAMLConfig
	err := r.db.WithContext(ctx).Where("organization_id = ?", orgID).First(&cfg).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (r *ssoRepo) SaveSAMLConfig(ctx context.Context, cfg *models.SAMLConfig) error {
	return r.db.WithContext(ctx).Save(cfg).Error
}

func (r *ssoRepo) GetOIDCConfig(ctx context.Context, orgID string) (*models.OIDCConfig, error) {
	var cfg models.OIDCConfig
	err := r.db.WithContext(ctx).Where("organization_id = ?", orgID).First(&cfg).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (r *ssoRepo) SaveOIDCConfig(ctx context.Context, cfg *models.OIDCConfig) error {
	return r.db.WithContext(ctx).Save(cfg).Error
}

func NewSSORepository(db *gorm.DB) repository.SSORepository {
	return &ssoRepo{db: db}
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

func (r *projectRepo) ListByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Project, error) {
	var projects []models.Project
	if len(ids) == 0 {
		return projects, nil
	}
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&projects).Error
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

// Store persists a single analytics log directly (API-server write path).
func (r *analyticsRepo) Store(ctx context.Context, logEntry *models.AnalyticsLog) error {
	return r.db.WithContext(ctx).Create(logEntry).Error
}

func NewAnalyticsRepository(db *gorm.DB) repository.AnalyticsRepository {
	return &analyticsRepo{db: db}
}

// PurgeExpiredByPlan removes analytics logs older than each plan's retention
// window. It queries plans and their associated subscriptions, then deletes
// logs for each retention tier.
func (r *analyticsRepo) PurgeExpiredByPlan(ctx context.Context) (int64, error) {
	var total int64
	var plans []models.Plan
	if err := r.db.WithContext(ctx).Find(&plans).Error; err != nil {
		return 0, err
	}
	for _, plan := range plans {
		if plan.AnalyticsRetentionDays <= 0 {
			continue
		}
		cutoff := time.Now().Add(-time.Duration(plan.AnalyticsRetentionDays) * 24 * time.Hour)
		var subUserIDs []uuid.UUID
		r.db.WithContext(ctx).Model(&models.Subscription{}).
			Where("plan_id = ? AND status = ?", plan.ID, "active").
			Pluck("user_id", &subUserIDs)
		if len(subUserIDs) == 0 {
			continue
		}
		var projectIDs []uuid.UUID
		r.db.WithContext(ctx).Model(&models.Project{}).
			Where("user_id IN ?", subUserIDs).
			Pluck("id", &projectIDs)
		if len(projectIDs) == 0 {
			continue
		}
		result := r.db.WithContext(ctx).
			Where("project_id IN ? AND timestamp < ?", projectIDs, cutoff).
			Delete(&models.AnalyticsLog{})
		if result.Error != nil {
			return total, result.Error
		}
		total += result.RowsAffected
	}
	return total, nil
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

func (r *analyticsRepo) GetTimeSeries(ctx context.Context, projectID uuid.UUID, start, end time.Time, bucket string) ([]map[string]interface{}, error) {
	validBuckets := map[string]bool{
		"minute": true,
		"hour":   true,
		"day":    true,
	}
	if !validBuckets[bucket] {
		bucket = "hour"
	}

	type DBResult struct {
		BucketTime time.Time `gorm:"column:bucket_time"`
		Decision   string    `gorm:"column:decision"`
		Count      int64     `gorm:"column:count"`
	}
	var dbResults []DBResult

	// postgres date_trunc
	query := `
		SELECT date_trunc(?, timestamp) as bucket_time, decision, COUNT(*) as count
		FROM analytics_logs
		WHERE project_id = ? AND timestamp BETWEEN ? AND ?
		GROUP BY bucket_time, decision
		ORDER BY bucket_time ASC
	`
	err := r.db.WithContext(ctx).Raw(query, bucket, projectID, start, end).Scan(&dbResults).Error
	if err != nil {
		return nil, err
	}

	timeMap := make(map[time.Time]map[string]interface{})
	for _, res := range dbResults {
		if _, ok := timeMap[res.BucketTime]; !ok {
			timeMap[res.BucketTime] = map[string]interface{}{
				"time":    res.BucketTime.Format(time.RFC3339),
				"allowed": int64(0),
				"blocked": int64(0),
			}
		}
		if res.Decision == "allowed" {
			timeMap[res.BucketTime]["allowed"] = res.Count
		} else if res.Decision == "blocked" {
			timeMap[res.BucketTime]["blocked"] = res.Count
		}
	}

	var results []map[string]interface{}
	seen := make(map[time.Time]bool)
	for _, res := range dbResults {
		if seen[res.BucketTime] {
			continue
		}
		seen[res.BucketTime] = true
		results = append(results, timeMap[res.BucketTime])
	}

	return results, nil
}

func (r *analyticsRepo) CountRequestsByProjects(ctx context.Context, projectIDs []uuid.UUID, start, end time.Time) (int64, error) {
	var count int64
	if len(projectIDs) == 0 {
		return 0, nil
	}
	err := r.db.WithContext(ctx).Model(&models.AnalyticsLog{}).
		Where("project_id IN ? AND timestamp BETWEEN ? AND ?", projectIDs, start, end).
		Count(&count).Error
	return count, err
}

// Password Reset Token Repo
type passwordResetTokenRepo struct {
	db *gorm.DB
}

func NewPasswordResetTokenRepository(db *gorm.DB) repository.PasswordResetTokenRepository {
	return &passwordResetTokenRepo{db: db}
}

func (r *passwordResetTokenRepo) Create(ctx context.Context, t *models.PasswordResetToken) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *passwordResetTokenRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*models.PasswordResetToken, error) {
	var t models.PasswordResetToken
	err := r.db.WithContext(ctx).First(&t, "token_hash = ? AND expires_at > ? AND used_at IS NULL", tokenHash, time.Now()).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *passwordResetTokenRepo) MarkUsed(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.PasswordResetToken{}).
		Where("id = ?", id).
		Update("used_at", &now).Error
}

func (r *passwordResetTokenRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.PasswordResetToken{}).Error
}

// Webhook Event Repo
type webhookEventRepo struct {
	db *gorm.DB
}

func NewWebhookEventRepository(db *gorm.DB) repository.WebhookEventRepository {
	return &webhookEventRepo{db: db}
}

func (r *webhookEventRepo) Create(ctx context.Context, e *models.WebhookEvent) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *webhookEventRepo) ListRecent(ctx context.Context, limit int) ([]models.WebhookEvent, error) {
	var events []models.WebhookEvent
	err := r.db.WithContext(ctx).Order("received_at DESC").Limit(limit).Find(&events).Error
	return events, err
}

func (r *webhookEventRepo) ListByEmail(ctx context.Context, email string, limit int) ([]models.WebhookEvent, error) {
	var events []models.WebhookEvent
	err := r.db.WithContext(ctx).Where("email = ?", email).Order("received_at DESC").Limit(limit).Find(&events).Error
	return events, err
}

// Project Member Repo
type projectMemberRepo struct {
	db *gorm.DB
}

func NewProjectMemberRepository(db *gorm.DB) repository.ProjectMemberRepository {
	return &projectMemberRepo{db: db}
}

func (r *projectMemberRepo) Add(ctx context.Context, m *models.ProjectMember) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *projectMemberRepo) Remove(ctx context.Context, projectID, memberID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("project_id = ? AND id = ?", projectID, memberID).
		Delete(&models.ProjectMember{}).Error
}

func (r *projectMemberRepo) UpdateRole(ctx context.Context, projectID, memberID uuid.UUID, role string) error {
	return r.db.WithContext(ctx).
		Model(&models.ProjectMember{}).
		Where("project_id = ? AND id = ?", projectID, memberID).
		Update("role", role).Error
}
func (r *projectMemberRepo) ListByProject(ctx context.Context, projectID uuid.UUID) ([]models.ProjectMember, error) {
	var members []models.ProjectMember
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&members).Error
	return members, err
}

func (r *projectMemberRepo) ListProjectIDsByUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	var projectIDs []uuid.UUID
	err := r.db.WithContext(ctx).Model(&models.ProjectMember{}).
		Where("user_id = ?", userID).
		Pluck("project_id", &projectIDs).Error
	return projectIDs, err
}

func (r *projectMemberRepo) IsMember(ctx context.Context, projectID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.ProjectMember{}).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Project Invite Repo
type projectInviteRepo struct {
	db *gorm.DB
}

func NewProjectInviteRepository(db *gorm.DB) repository.ProjectInviteRepository {
	return &projectInviteRepo{db: db}
}

func (r *projectInviteRepo) Create(ctx context.Context, inv *models.ProjectInvite) error {
	return r.db.WithContext(ctx).Create(inv).Error
}

func (r *projectInviteRepo) GetByTokenHash(ctx context.Context, hash string) (*models.ProjectInvite, error) {
	var inv models.ProjectInvite
	err := r.db.WithContext(ctx).First(&inv, "token_hash = ?", hash).Error
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *projectInviteRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.ProjectInvite, error) {
	var inv models.ProjectInvite
	err := r.db.WithContext(ctx).First(&inv, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *projectInviteRepo) ListByProject(ctx context.Context, projectID uuid.UUID) ([]models.ProjectInvite, error) {
	var invites []models.ProjectInvite
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND status = ?", projectID, "pending").
		Order("created_at DESC").
		Find(&invites).Error
	return invites, err
}

func (r *projectInviteRepo) ListPendingByEmail(ctx context.Context, email string) ([]models.ProjectInvite, error) {
	var invites []models.ProjectInvite
	err := r.db.WithContext(ctx).
		Where("email = ? AND status = ?", email, "pending").
		Order("created_at DESC").
		Find(&invites).Error
	return invites, err
}

func (r *projectInviteRepo) Update(ctx context.Context, inv *models.ProjectInvite) error {
	return r.db.WithContext(ctx).Save(inv).Error
}

func (r *projectInviteRepo) ListExpired(ctx context.Context) ([]models.ProjectInvite, error) {
	var invites []models.ProjectInvite
	err := r.db.WithContext(ctx).
		Where("status = ? AND expires_at < NOW()", "pending").
		Find(&invites).Error
	return invites, err
}

func (r *projectInviteRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.ProjectInvite{}, "id = ?", id).Error
}

// Session management for refresh tokens
func (r *refreshTokenRepo) ListActiveByUserID(ctx context.Context, userID uuid.UUID) ([]models.RefreshToken, error) {
	var tokens []models.RefreshToken
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND revoked = false AND expires_at > NOW()", userID).
		Order("created_at DESC").
		Find(&tokens).Error
	return tokens, err
}

func (r *refreshTokenRepo) RevokeByID(ctx context.Context, userID, tokenID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.RefreshToken{}).
		Where("id = ? AND user_id = ?", tokenID, userID).
		Update("revoked", true).Error
}

// Alert Repo
type alertRepo struct {
	db *gorm.DB
}

func NewAlertRepository(db *gorm.DB) repository.AlertRepository {
	return &alertRepo{db: db}
}

func (r *alertRepo) CreateRule(ctx context.Context, rule *models.AlertRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

func (r *alertRepo) GetRule(ctx context.Context, id uuid.UUID) (*models.AlertRule, error) {
	var rule models.AlertRule
	if err := r.db.WithContext(ctx).First(&rule, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *alertRepo) ListRulesByProject(ctx context.Context, projectID uuid.UUID) ([]models.AlertRule, error) {
	var rules []models.AlertRule
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Order("created_at DESC").Find(&rules).Error
	return rules, err
}

func (r *alertRepo) ListActiveRules(ctx context.Context) ([]models.AlertRule, error) {
	var rules []models.AlertRule
	err := r.db.WithContext(ctx).Where("is_active = true").Find(&rules).Error
	return rules, err
}

func (r *alertRepo) UpdateRule(ctx context.Context, rule *models.AlertRule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

func (r *alertRepo) DeleteRule(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.AlertEvent{}, "rule_id = ?", id).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Delete(&models.AlertRule{}, "id = ?", id).Error
}

func (r *alertRepo) CreateEvent(ctx context.Context, e *models.AlertEvent) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *alertRepo) ListEventsByProject(ctx context.Context, projectID uuid.UUID, limit int) ([]models.AlertEvent, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var events []models.AlertEvent
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).
		Order("created_at DESC").Limit(limit).Find(&events).Error
	return events, err
}

// IP Access Repo
type ipAccessRepo struct {
	db *gorm.DB
}

func NewIPAccessRepository(db *gorm.DB) repository.IPAccessRepository {
	return &ipAccessRepo{db: db}
}

func (r *ipAccessRepo) Create(ctx context.Context, rule *models.IPAccessRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

func (r *ipAccessRepo) ListByProject(ctx context.Context, projectID uuid.UUID) ([]models.IPAccessRule, error) {
	var rules []models.IPAccessRule
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Order("created_at DESC").Find(&rules).Error
	return rules, err
}

func (r *ipAccessRepo) Delete(ctx context.Context, projectID, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.IPAccessRule{}, "id = ? AND project_id = ?", id, projectID).Error
}
