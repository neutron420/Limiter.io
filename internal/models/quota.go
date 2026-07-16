package models

import "time"

type Quota struct {
	ID             string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	ProjectID      string    `gorm:"type:uuid;not null;index" json:"project_id"`
	PerMinute      int64     `gorm:"default:0" json:"per_minute"`
	PerHour        int64     `gorm:"default:0" json:"per_hour"`
	PerDay         int64     `gorm:"default:0" json:"per_day"`
	PerMonth       int64     `gorm:"default:0" json:"per_month"`
	CurrentMinute  int64     `gorm:"default:0" json:"-"`
	CurrentHour    int64     `gorm:"default:0" json:"-"`
	CurrentDay     int64     `gorm:"default:0" json:"-"`
	CurrentMonth   int64     `gorm:"default:0" json:"-"`
	WindowStartMin time.Time `json:"-"`
	WindowStartHour time.Time `json:"-"`
	WindowStartDay  time.Time `json:"-"`
	WindowStartMonth time.Time `json:"-"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (Quota) TableName() string {
	return "quotas"
}
