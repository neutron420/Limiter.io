package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"limiter.io/internal/models"
	"limiter.io/internal/services"
)

type OrganizationHandler struct {
	svc *services.OrganizationService
	db  *gorm.DB
}

func NewOrganizationHandler(db *gorm.DB) *OrganizationHandler {
	return &OrganizationHandler{
		svc: services.NewOrganizationService(db),
		db:  db,
	}
}

func (h *OrganizationHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")
	var org models.Organization
	if err := c.ShouldBindJSON(&org); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	org.OwnerID = userID
	if err := h.svc.Create(&org); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.svc.AddMember(org.ID, userID, "owner")
	c.JSON(http.StatusCreated, org)
}

func (h *OrganizationHandler) GetByID(c *gin.Context) {
	org, err := h.svc.GetByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		return
	}
	c.JSON(http.StatusOK, org)
}

func (h *OrganizationHandler) ListByUser(c *gin.Context) {
	userID := c.GetString("user_id")
	orgs, err := h.svc.ListByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orgs)
}

func (h *OrganizationHandler) AddMember(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.AddMember(c.Param("id"), req.UserID, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "member added"})
}

func (h *OrganizationHandler) RemoveMember(c *gin.Context) {
	userID := c.Param("userId")
	if err := h.svc.RemoveMember(c.Param("id"), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "member removed"})
}

func (h *OrganizationHandler) ListMembers(c *gin.Context) {
	members, err := h.svc.ListMembers(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, members)
}

func (h *OrganizationHandler) CreateGroup(c *gin.Context) {
	var group models.OrganizationGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	group.OrganizationID = c.Param("id")
	if err := h.svc.CreateGroup(&group); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, group)
}

func (h *OrganizationHandler) DeleteGroup(c *gin.Context) {
	if err := h.svc.DeleteGroup(c.Param("groupId")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "group deleted"})
}

func (h *OrganizationHandler) ListGroups(c *gin.Context) {
	groups, err := h.svc.ListGroups(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, groups)
}

func (h *OrganizationHandler) AddToGroup(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.AddToGroup(c.Param("groupId"), req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "added to group"})
}

func (h *OrganizationHandler) RemoveFromGroup(c *gin.Context) {
	if err := h.svc.RemoveFromGroup(c.Param("groupId"), c.Param("userId")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "removed from group"})
}

type ApprovalHandler struct {
	svc *services.ApprovalService
}

func NewApprovalHandler(db *gorm.DB) *ApprovalHandler {
	return &ApprovalHandler{svc: services.NewApprovalService(db)}
}

func (h *ApprovalHandler) CreateWorkflow(c *gin.Context) {
	var wf models.ApprovalWorkflow
	if err := c.ShouldBindJSON(&wf); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	wf.OrganizationID = c.Param("orgId")
	if err := h.svc.CreateWorkflow(&wf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, wf)
}

func (h *ApprovalHandler) ListWorkflows(c *gin.Context) {
	wfs, err := h.svc.ListWorkflows(c.Param("orgId"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, wfs)
}

func (h *ApprovalHandler) RequestApproval(c *gin.Context) {
	var req models.ApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.OrganizationID = c.Param("orgId")
	req.RequestedBy = c.GetString("user_id")
	if err := h.svc.RequestApproval(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, req)
}

func (h *ApprovalHandler) Approve(c *gin.Context) {
	userID := c.GetString("user_id")
	if err := h.svc.Approve(c.Param("id"), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "approved"})
}

func (h *ApprovalHandler) Reject(c *gin.Context) {
	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)
	userID := c.GetString("user_id")
	if err := h.svc.Reject(c.Param("id"), userID, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "rejected"})
}

func (h *ApprovalHandler) ListRequests(c *gin.Context) {
	reqs, err := h.svc.ListRequests(c.Param("orgId"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, reqs)
}

func (h *ApprovalHandler) GetRequest(c *gin.Context) {
	req, err := h.svc.GetRequest(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "request not found"})
		return
	}
	c.JSON(http.StatusOK, req)
}
