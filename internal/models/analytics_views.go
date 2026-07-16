package models

import "time"

type SavedAnalyticsView struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	ProjectID string    `gorm:"type:uuid;not null;index" json:"project_id"`
	UserID    string    `gorm:"type:uuid;not null;index" json:"user_id"`
	Name      string    `gorm:"type:varchar(255);not null" json:"name"`
	Config    string    `gorm:"type:jsonb;not null" json:"config"`
	IsShared  bool      `gorm:"default:false" json:"is_shared"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AnomalyDetectionConfig struct {
	ID                    string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	ProjectID             string    `gorm:"type:uuid;not null;index" json:"project_id"`
	Enabled               bool      `gorm:"default:false" json:"enabled"`
	Sensitivity           float64   `gorm:"default:2.0" json:"sensitivity"`
	LookbackMinutes       int       `gorm:"default:60" json:"lookback_minutes"`
	AlertOnSpike          bool      `gorm:"default:true" json:"alert_on_spike"`
	AlertOnDrop           bool      `gorm:"default:false" json:"alert_on_drop"`
	SlackWebhookURL       string    `gorm:"type:text" json:"slack_webhook_url,omitempty"`
	LastAlertedAt         *time.Time `json:"last_alerted_at,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

func (SavedAnalyticsView) TableName() string    { return "saved_analytics_views" }
func (AnomalyDetectionConfig) TableName() string { return "anomaly_detection_configs" }
