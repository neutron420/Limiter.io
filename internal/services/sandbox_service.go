package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"limiter.io/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SandboxService struct {
	db *gorm.DB
}

func NewSandboxService(db *gorm.DB) *SandboxService {
	return &SandboxService{db: db}
}

func (s *SandboxService) CreateSandboxProject(userID string) (*models.Project, error) {
	project := models.Project{
		Name:        "Sandbox - " + time.Now().Format("2006-01-02 15:04:05"),
		Description: "Auto-created sandbox for testing",
	}
	if err := s.db.Create(&project).Error; err != nil {
		return nil, err
	}

	rawKey := fmt.Sprintf("rl_sandbox_%s", uuid.New().String()[:8])
	hash := sha256.Sum256([]byte(rawKey))

	apiKey := models.APIKey{
		Name:     "sandbox-key",
		ProjectID: project.ID,
		KeyHash:  hex.EncodeToString(hash[:]),
		Prefix:   rawKey[:12],
		Scope:    "admin",
	}
	if err := s.db.Create(&apiKey).Error; err != nil {
		return nil, err
	}

	rule := models.RateLimitRule{
		Name:         "sandbox-rule",
		RoutePattern: "/*",
		Limit:        50,
		Period:       60,
		IsActive:     true,
		Priority:     10,
		ProjectID:    project.ID,
	}
	s.db.Create(&rule)

	return &project, nil
}

func (s *SandboxService) CleanupSandboxProject(projectID uuid.UUID) error {
	s.db.Where("project_id = ?", projectID).Delete(&models.APIKey{})
	s.db.Where("project_id = ?", projectID).Delete(&models.RateLimitRule{})
	s.db.Where("project_id = ?", projectID).Delete(&models.AnalyticsLog{})
	return s.db.Delete(&models.Project{}, "id = ?", projectID).Error
}
