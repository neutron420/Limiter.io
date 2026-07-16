package services

import (
	"log"
	"time"

	"limiter.io/internal/models"
	"gorm.io/gorm"
)

type KeyRotationService struct {
	db *gorm.DB
}

func NewKeyRotationService(db *gorm.DB) *KeyRotationService {
	return &KeyRotationService{db: db}
}

func (s *KeyRotationService) CheckRotationReminders() {
	var keys []models.APIKey
	s.db.Where("created_at < ?", time.Now().Add(-90*24*time.Hour)).Find(&keys)

	for _, key := range keys {
		log.Printf("Rotation reminder: API key %s (%s) created at %s is overdue for rotation",
			key.ID, key.Name, key.CreatedAt.Format("2006-01-02"))
	}
}

func (s *KeyRotationService) StartRotationChecker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			s.CheckRotationReminders()
		}
	}()
}
