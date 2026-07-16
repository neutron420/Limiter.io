package services

import (
	"strings"
	"time"

	"limiter.io/internal/models"
	"gorm.io/gorm"
)

type ApprovalService struct {
	db *gorm.DB
}

func NewApprovalService(db *gorm.DB) *ApprovalService {
	return &ApprovalService{db: db}
}

func (s *ApprovalService) CreateWorkflow(wf *models.ApprovalWorkflow) error {
	return s.db.Create(wf).Error
}

func (s *ApprovalService) ListWorkflows(orgID string) ([]models.ApprovalWorkflow, error) {
	var wfs []models.ApprovalWorkflow
	err := s.db.Where("organization_id = ?", orgID).Find(&wfs).Error
	return wfs, err
}

func (s *ApprovalService) RequestApproval(req *models.ApprovalRequest) error {
	req.Status = "pending"
	return s.db.Create(req).Error
}

func (s *ApprovalService) Approve(reqID, userID string) error {
	var req models.ApprovalRequest
	if err := s.db.First(&req, "id = ?", reqID).Error; err != nil {
		return err
	}
	var existing []string
	if req.ApprovedBy != nil {
		existing = req.ApprovedBy
	}
	for _, id := range existing {
		if id == userID {
			return nil
		}
	}
	req.ApprovedBy = append(existing, userID)
	var workflow models.ApprovalWorkflow
	s.db.First(&workflow, "id = ?", req.WorkflowID)
	if len(req.ApprovedBy) >= workflow.MinApprovers {
		req.Status = "approved"
	}
	req.UpdatedAt = time.Now()
	return s.db.Save(&req).Error
}

func (s *ApprovalService) Reject(reqID, userID string, reason string) error {
	return s.db.Model(&models.ApprovalRequest{}).Where("id = ?", reqID).
		Updates(map[string]interface{}{"status": "rejected", "reason": reason, "updated_at": time.Now()}).Error
}

func (s *ApprovalService) ListRequests(orgID string) ([]models.ApprovalRequest, error) {
	var reqs []models.ApprovalRequest
	err := s.db.Where("organization_id = ?", orgID).Order("created_at DESC").Find(&reqs).Error
	return reqs, err
}

func (s *ApprovalService) GetRequest(id string) (*models.ApprovalRequest, error) {
	var req models.ApprovalRequest
	err := s.db.First(&req, "id = ?", id).Error
	return &req, err
}

func (s *ApprovalService) RequiresApproval(orgID, actionType string, userID string) (bool, string, error) {
	var workflow models.ApprovalWorkflow
	err := s.db.Where("organization_id = ? AND action_type = ? AND enabled = ?", orgID, actionType, true).
		First(&workflow).Error
	if err != nil {
		return false, "", nil
	}
	isApprover := false
	if workflow.ApproverGroup != "" {
		for _, g := range strings.Split(workflow.ApproverGroup, ",") {
			if strings.TrimSpace(g) == userID {
				isApprover = true
				break
			}
		}
	}
	if isApprover {
		return false, "", nil
	}
	return true, workflow.ID, nil
}
