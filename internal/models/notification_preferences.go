package models

import "time"

type NotificationPreferences struct {
	ID                  string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID              string    `gorm:"type:uuid;not null;index" json:"user_id"`
	ProjectID           string    `gorm:"type:uuid;not null;index" json:"project_id"`
	EmailNotifications  bool      `gorm:"default:true" json:"email_notifications"`
	SlackNotifications  bool      `gorm:"default:false" json:"slack_notifications"`
	SlackWebhookURL     string    `gorm:"type:text" json:"slack_webhook_url,omitempty"`
	RateLimitAlerts     bool      `gorm:"default:true" json:"rate_limit_alerts"`
	MemberJoinAlerts    bool      `gorm:"default:true" json:"member_join_alerts"`
	KeyRotationAlerts   bool      `gorm:"default:false" json:"key_rotation_alerts"`
	WeeklyDigest        bool      `gorm:"default:true" json:"weekly_digest"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func (NotificationPreferences) TableName() string {
	return "notification_preferences"
}
