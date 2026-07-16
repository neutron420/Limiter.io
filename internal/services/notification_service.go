package services

import (
	"time"

	"limiter.io/internal/models"
	"gorm.io/gorm"
)

type NotificationService struct {
	db *gorm.DB
}

func NewNotificationService(db *gorm.DB) *NotificationService {
	return &NotificationService{db: db}
}

func (s *NotificationService) GetPreferences(userID, projectID string) (*models.NotificationPreferences, error) {
	var prefs models.NotificationPreferences
	err := s.db.Where("user_id = ? AND project_id = ?", userID, projectID).First(&prefs).Error
	if err != nil {
		return &models.NotificationPreferences{
			UserID:             userID,
			ProjectID:          projectID,
			EmailNotifications: true,
			RateLimitAlerts:    true,
			MemberJoinAlerts:   true,
			WeeklyDigest:       true,
		}, nil
	}
	return &prefs, nil
}

func (s *NotificationService) UpdatePreferences(prefs *models.NotificationPreferences) error {
	var existing models.NotificationPreferences
	result := s.db.Where("user_id = ? AND project_id = ?", prefs.UserID, prefs.ProjectID).First(&existing)
	if result.Error != nil {
		return s.db.Create(prefs).Error
	}
	prefs.ID = existing.ID
	prefs.CreatedAt = existing.CreatedAt
	prefs.UpdatedAt = time.Now()
	return s.db.Save(prefs).Error
}

func (s *NotificationService) SendRateLimitAlert(projectID string, ruleName string, threshold int) {
	s.db.Model(&models.NotificationPreferences{}).
		Where("project_id = ? AND rate_limit_alerts = ? AND email_notifications = ?", projectID, true, true).
		Find(&[]models.NotificationPreferences{})
}

func (s *NotificationService) SendMemberJoinAlert(projectID string, memberEmail string) {
	s.db.Model(&models.NotificationPreferences{}).
		Where("project_id = ? AND member_join_alerts = ? AND email_notifications = ?", projectID, true, true).
		Find(&[]models.NotificationPreferences{})
}
