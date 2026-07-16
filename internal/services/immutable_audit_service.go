package services

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"limiter.io/internal/models"
	"gorm.io/gorm"
)

type ImmutableAuditService struct {
	db *gorm.DB
}

func NewImmutableAuditService(db *gorm.DB) *ImmutableAuditService {
	return &ImmutableAuditService{db: db}
}

func (s *ImmutableAuditService) Log(userID, projectID, action, resource string, details interface{}, ipAddress, userAgent string) (*models.ImmutableAuditLog, error) {
	detailsJSON, _ := json.Marshal(details)

	var lastLog models.ImmutableAuditLog
	prevHash := ""
	s.db.Order("created_at DESC").Limit(1).First(&lastLog)
	if lastLog.ID != "" {
		prevHash = lastLog.Checksum
	}

	entry := models.ImmutableAuditLog{
		UserID:    userID,
		ProjectID: projectID,
		Action:    action,
		Resource:  resource,
		Details:   string(detailsJSON),
		IPAddress: ipAddress,
		UserAgent: userAgent,
		PrevHash:  prevHash,
		CreatedAt: time.Now().UTC(),
	}

	hashInput := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%d",
		entry.UserID, entry.ProjectID, entry.Action, entry.Resource,
		entry.Details, entry.IPAddress, entry.PrevHash, entry.CreatedAt.UnixNano())
	hash := sha256.Sum256([]byte(hashInput))
	entry.Checksum = hex.EncodeToString(hash[:])

	if err := s.db.Create(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

func (s *ImmutableAuditService) VerifyChain() (bool, error) {
	var logs []models.ImmutableAuditLog
	if err := s.db.Order("created_at ASC").Find(&logs).Error; err != nil {
		return false, err
	}
	var prevHash string
	for _, entry := range logs {
		hashInput := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%d",
			entry.UserID, entry.ProjectID, entry.Action, entry.Resource,
			entry.Details, entry.IPAddress, entry.PrevHash, entry.CreatedAt.UnixNano())
		hash := sha256.Sum256([]byte(hashInput))
		expectedHash := hex.EncodeToString(hash[:])
		if entry.Checksum != expectedHash {
			return false, fmt.Errorf("tampered log: %s", entry.ID)
		}
		if entry.PrevHash != prevHash {
			return false, fmt.Errorf("broken chain at: %s", entry.ID)
		}
		prevHash = entry.Checksum
	}
	return true, nil
}

func (s *ImmutableAuditService) List(projectID string, limit, offset int) ([]models.ImmutableAuditLog, int64, error) {
	var logs []models.ImmutableAuditLog
	var total int64
	query := s.db.Model(&models.ImmutableAuditLog{})
	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, total, err
}

func (s *ImmutableAuditService) GetByID(id string) (*models.ImmutableAuditLog, error) {
	var log models.ImmutableAuditLog
	err := s.db.First(&log, "id = ?", id).Error
	return &log, err
}
