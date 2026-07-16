package services

import (
	"sync"
	"time"

	"limiter.io/internal/models"
	"gorm.io/gorm"
)

type QuotaService struct {
	db    *gorm.DB
	mu    sync.RWMutex
	cache map[string]*models.Quota
}

func NewQuotaService(db *gorm.DB) *QuotaService {
	return &QuotaService{
		db:    db,
		cache: make(map[string]*models.Quota),
	}
}

func (s *QuotaService) GetQuota(projectID string) (*models.Quota, error) {
	s.mu.RLock()
	if q, ok := s.cache[projectID]; ok {
		s.mu.RUnlock()
		return q, nil
	}
	s.mu.RUnlock()

	var quota models.Quota
	err := s.db.Where("project_id = ?", projectID).First(&quota).Error
	if err != nil {
		quota = models.Quota{ProjectID: projectID}
		s.db.Create(&quota)
	}

	s.mu.Lock()
	s.cache[projectID] = &quota
	s.mu.Unlock()
	return &quota, nil
}

func (s *QuotaService) SetQuota(quota *models.Quota) error {
	var existing models.Quota
	result := s.db.Where("project_id = ?", quota.ProjectID).First(&existing)
	if result.Error != nil {
		return s.db.Create(quota).Error
	}
	quota.ID = existing.ID
	quota.CreatedAt = existing.CreatedAt
	quota.UpdatedAt = time.Now()

	s.mu.Lock()
	s.cache[quota.ProjectID] = quota
	s.mu.Unlock()

	return s.db.Save(quota).Error
}

func (s *QuotaService) CheckQuota(projectID string) (bool, error) {
	quota, err := s.GetQuota(projectID)
	if err != nil {
		return true, nil
	}
	now := time.Now().UTC()

	if quota.PerMinute > 0 {
		if now.Sub(quota.WindowStartMin) > time.Minute {
			quota.WindowStartMin = now
			quota.CurrentMinute = 0
		}
		if quota.CurrentMinute >= quota.PerMinute {
			return false, nil
		}
	}

	if quota.PerHour > 0 {
		if now.Sub(quota.WindowStartHour) > time.Hour {
			quota.WindowStartHour = now
			quota.CurrentHour = 0
		}
		if quota.CurrentHour >= quota.PerHour {
			return false, nil
		}
	}

	if quota.PerDay > 0 {
		if now.Sub(quota.WindowStartDay) > 24*time.Hour {
			quota.WindowStartDay = now
			quota.CurrentDay = 0
		}
		if quota.CurrentDay >= quota.PerDay {
			return false, nil
		}
	}

	if quota.PerMonth > 0 {
		if now.Sub(quota.WindowStartMonth) > 30*24*time.Hour {
			quota.WindowStartMonth = now
			quota.CurrentMonth = 0
		}
		if quota.CurrentMonth >= quota.PerMonth {
			return false, nil
		}
	}

	return true, nil
}

func (s *QuotaService) Increment(projectID string, weight int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	quota, ok := s.cache[projectID]
	if !ok {
		return
	}
	now := time.Now().UTC()

	if quota.PerMinute > 0 {
		if now.Sub(quota.WindowStartMin) > time.Minute {
			quota.WindowStartMin = now
			quota.CurrentMinute = 0
		}
		quota.CurrentMinute += weight
	}
	if quota.PerHour > 0 {
		if now.Sub(quota.WindowStartHour) > time.Hour {
			quota.WindowStartHour = now
			quota.CurrentHour = 0
		}
		quota.CurrentHour += weight
	}
	if quota.PerDay > 0 {
		if now.Sub(quota.WindowStartDay) > 24*time.Hour {
			quota.WindowStartDay = now
			quota.CurrentDay = 0
		}
		quota.CurrentDay += weight
	}
	if quota.PerMonth > 0 {
		if now.Sub(quota.WindowStartMonth) > 30*24*time.Hour {
			quota.WindowStartMonth = now
			quota.CurrentMonth = 0
		}
		quota.CurrentMonth += weight
	}
}
