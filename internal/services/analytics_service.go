package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"limiter.io/internal/models"
	"limiter.io/internal/repository"
)

type AnalyticsService struct {
	analRepo   repository.AnalyticsRepository
	projRepo   repository.ProjectRepository
	memberRepo repository.ProjectMemberRepository
}

func NewAnalyticsService(analRepo repository.AnalyticsRepository, projRepo repository.ProjectRepository, memberRepo repository.ProjectMemberRepository) *AnalyticsService {
	return &AnalyticsService{analRepo: analRepo, projRepo: projRepo, memberRepo: memberRepo}
}

func (s *AnalyticsService) GetStats(ctx context.Context, projectID uuid.UUID) (map[string]interface{}, error) {
	return map[string]interface{}{
		"total_requests":  0,
		"blocked_requests": 0,
		"avg_latency_ms":  0,
	}, nil
}

func (s *AnalyticsService) GetLogs(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]models.AnalyticsLog, error) {
	return s.analRepo.GetLogs(ctx, projectID, limit, offset)
}

func (s *AnalyticsService) GetTimeSeries(ctx context.Context, projectID uuid.UUID, startStr, endStr, bucket string) ([]map[string]interface{}, error) {
	return s.analRepo.GetTimeSeries(ctx, projectID, time.Now().Add(-24*time.Hour), time.Now(), bucket)
}

type AnalyticsDataService struct {
	db *gorm.DB
}

func NewAnalyticsDataService(db *gorm.DB) *AnalyticsDataService {
	return &AnalyticsDataService{db: db}
}

func (s *AnalyticsDataService) SaveView(view *models.SavedAnalyticsView) error {
	return s.db.Create(view).Error
}

func (s *AnalyticsDataService) ListViews(projectID, userID string) ([]models.SavedAnalyticsView, error) {
	var views []models.SavedAnalyticsView
	err := s.db.Where("project_id = ? AND (user_id = ? OR is_shared = ?)", projectID, userID, true).Find(&views).Error
	if views == nil {
		views = []models.SavedAnalyticsView{}
	}
	return views, err
}

func (s *AnalyticsDataService) GetView(id string) (*models.SavedAnalyticsView, error) {
	var view models.SavedAnalyticsView
	err := s.db.First(&view, "id = ?", id).Error
	return &view, err
}

func (s *AnalyticsDataService) DeleteView(id, userID string) error {
	return s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.SavedAnalyticsView{}).Error
}

func (s *AnalyticsDataService) GetAnomalyConfig(projectID string) (*models.AnomalyDetectionConfig, error) {
	var cfg models.AnomalyDetectionConfig
	err := s.db.Where("project_id = ?", projectID).First(&cfg).Error
	if err != nil {
		return &models.AnomalyDetectionConfig{
			ProjectID:   projectID,
			Enabled:     false,
			Sensitivity: 2.0,
		}, nil
	}
	return &cfg, nil
}

func (s *AnalyticsDataService) UpdateAnomalyConfig(cfg *models.AnomalyDetectionConfig) error {
	var existing models.AnomalyDetectionConfig
	result := s.db.Where("project_id = ?", cfg.ProjectID).First(&existing)
	if result.Error != nil {
		return s.db.Create(cfg).Error
	}
	cfg.ID = existing.ID
	cfg.CreatedAt = existing.CreatedAt
	cfg.UpdatedAt = time.Now()
	return s.db.Save(cfg).Error
}

func (s *AnalyticsDataService) DetectAnomalies(projectID string) ([]string, error) {
	cfg, err := s.GetAnomalyConfig(projectID)
	if err != nil || !cfg.Enabled {
		return nil, nil
	}

	var recentLogs []models.AnalyticsLog
	s.db.Where("project_id = ? AND created_at > ?", projectID, time.Now().Add(-time.Duration(cfg.LookbackMinutes)*time.Minute)).
		Order("created_at DESC").Limit(100).Find(&recentLogs)

	if len(recentLogs) < 10 {
		return nil, nil
	}

	requestCount := 0
	blockedCount := 0
	for _, l := range recentLogs {
		requestCount++
		if l.Decision == "blocked" {
			blockedCount++
		}
	}
	if requestCount > 0 {
		blockRate := float64(blockedCount) / float64(requestCount)
		if blockRate > 0.5 {
			return []string{"High block rate detected"}, nil
		}
	}

	var values []float64
	for _, l := range recentLogs {
		if l.LatencyMs > 0 {
			values = append(values, float64(l.LatencyMs))
		}
	}

	alerts := []string{}
	if len(values) >= 10 {
		mean, std := meanStd(values)
		threshold := mean + cfg.Sensitivity*std
		for _, l := range recentLogs {
			if float64(l.LatencyMs) > threshold && l.LatencyMs > 5000 {
				alerts = append(alerts, "Latency anomaly detected")
				break
			}
		}
	}
	return alerts, nil
}

func (s *AnalyticsDataService) GetAnalytics(projectID string) (map[string]interface{}, error) {
	var totalRequests int64
	var blockedRequests int64
	var avgLatency float64

	query := s.db.Model(&models.AnalyticsLog{}).Where("project_id = ?", projectID)
	query.Count(&totalRequests)
	query.Where("decision = ?", "blocked").Count(&blockedRequests)
	query.Select("AVG(latency_ms)").Scan(&avgLatency)

	var topRoutes []struct {
		Route  string
		Count  int64
	}
	s.db.Model(&models.AnalyticsLog{}).
		Select("route, COUNT(*) as count").
		Where("project_id = ?", projectID).
		Group("route").
		Order("count DESC").
		Limit(10).
		Scan(&topRoutes)

	var topKeys []struct {
		APIKeyID string
		Count    int64
	}
	s.db.Model(&models.AnalyticsLog{}).
		Select("api_key_id, COUNT(*) as count").
		Where("project_id = ?", projectID).
		Group("api_key_id").
		Order("count DESC").
		Limit(10).
		Scan(&topKeys)

	var latencies []float64
	s.db.Model(&models.AnalyticsLog{}).
		Where("project_id = ? AND latency_ms > 0", projectID).
		Order("latency_ms ASC").
		Pluck("latency_ms", &latencies)

	p95, p99 := 0.0, 0.0
	if len(latencies) > 0 {
		n := len(latencies)
		p95Idx := int(math.Ceil(float64(n)*0.95)) - 1
		p99Idx := int(math.Ceil(float64(n)*0.99)) - 1
		if p95Idx >= n {
			p95Idx = n - 1
		}
		if p99Idx >= n {
			p99Idx = n - 1
		}
		p95 = latencies[p95Idx]
		p99 = latencies[p99Idx]
	}

	return map[string]interface{}{
		"total_requests":   totalRequests,
		"blocked_requests":  blockedRequests,
		"avg_latency_ms":   avgLatency,
		"p95":              p95,
		"p99":              p99,
		"top_routes":       topRoutes,
		"top_api_keys":     topKeys,
	}, nil
}

func (s *AnalyticsDataService) SaveAnalyticsView(view *models.SavedAnalyticsView) error {
	return s.SaveView(view)
}

func (s *AnalyticsDataService) ListAnalyticsViews(projectID, userID string) ([]models.SavedAnalyticsView, error) {
	return s.ListViews(projectID, userID)
}

func (s *AnalyticsDataService) GetAnalyticsView(id string) (*models.SavedAnalyticsView, error) {
	return s.GetView(id)
}

func (s *AnalyticsDataService) DeleteAnalyticsView(id, userID string) error {
	return s.DeleteView(id, userID)
}

func (s *AnalyticsDataService) GetAnomalyConfigItem(projectID string) (*models.AnomalyDetectionConfig, error) {
	return s.GetAnomalyConfig(projectID)
}

func (s *AnalyticsDataService) UpdateAnomalyConfigItem(cfg *models.AnomalyDetectionConfig) error {
	return s.UpdateAnomalyConfig(cfg)
}

type AnalyticsViewConfig struct {
	Metrics     []string `json:"metrics"`
	TimeRange   string   `json:"time_range"`
	Granularity string   `json:"granularity"`
	Filters     struct {
		Routes  []string `json:"routes,omitempty"`
		APIKeys []string `json:"api_keys,omitempty"`
		Status  []int    `json:"status,omitempty"`
	} `json:"filters,omitempty"`
}

func ParseViewConfig(config string) (*AnalyticsViewConfig, error) {
	var vc AnalyticsViewConfig
	err := json.Unmarshal([]byte(config), &vc)
	return &vc, err
}

func meanStd(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))
	var variance float64
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	variance /= float64(len(values))
	return mean, math.Sqrt(variance)
}

var _ = fmt.Sprintf
