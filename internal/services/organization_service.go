package services

import (
	"fmt"
	"strings"
	"time"

	"limiter.io/internal/models"
	"gorm.io/gorm"
)

type OrganizationService struct {
	db *gorm.DB
}

func NewOrganizationService(db *gorm.DB) *OrganizationService {
	return &OrganizationService{db: db}
}

func (s *OrganizationService) Create(org *models.Organization) error {
	org.Slug = strings.ToLower(strings.ReplaceAll(org.Name, " ", "-"))
	var count int64
	s.db.Model(&models.Organization{}).Where("slug = ?", org.Slug).Count(&count)
	if count > 0 {
		org.Slug = fmt.Sprintf("%s-%d", org.Slug, time.Now().Unix())
	}
	return s.db.Create(org).Error
}

func (s *OrganizationService) GetByID(id string) (*models.Organization, error) {
	var org models.Organization
	err := s.db.First(&org, "id = ?", id).Error
	return &org, err
}

func (s *OrganizationService) ListByUser(userID string) ([]models.Organization, error) {
	var orgs []models.Organization
	err := s.db.Joins("JOIN organization_members ON organization_members.organization_id = organizations.id").
		Where("organization_members.user_id = ?", userID).
		Find(&orgs).Error
	return orgs, err
}

func (s *OrganizationService) AddMember(orgID, userID, role string) error {
	member := models.OrganizationMember{
		OrganizationID: orgID,
		UserID:         userID,
		Role:           role,
		JoinedAt:       time.Now(),
	}
	return s.db.Create(&member).Error
}

func (s *OrganizationService) RemoveMember(orgID, userID string) error {
	return s.db.Where("organization_id = ? AND user_id = ?", orgID, userID).
		Delete(&models.OrganizationMember{}).Error
}

func (s *OrganizationService) CreateGroup(group *models.OrganizationGroup) error {
	return s.db.Create(group).Error
}

func (s *OrganizationService) AddToGroup(groupID, userID string) error {
	member := models.OrganizationGroupMember{
		GroupID: groupID,
		UserID:  userID,
	}
	return s.db.Create(&member).Error
}

func (s *OrganizationService) RemoveFromGroup(groupID, userID string) error {
	return s.db.Where("group_id = ? AND user_id = ?", groupID, userID).
		Delete(&models.OrganizationGroupMember{}).Error
}

func (s *OrganizationService) ListGroups(orgID string) ([]models.OrganizationGroup, error) {
	var groups []models.OrganizationGroup
	err := s.db.Where("organization_id = ?", orgID).Find(&groups).Error
	return groups, err
}

func (s *OrganizationService) DeleteGroup(groupID string) error {
	return s.db.Where("id = ?", groupID).Delete(&models.OrganizationGroup{}).Error
}

func (s *OrganizationService) ListMembers(orgID string) ([]models.OrganizationMember, error) {
	var members []models.OrganizationMember
	err := s.db.Where("organization_id = ?", orgID).Find(&members).Error
	return members, err
}
