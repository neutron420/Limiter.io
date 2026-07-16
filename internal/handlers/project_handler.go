package handlers

import (
	"net/http"

	"limiter.io/internal/dto"
	"limiter.io/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProjectHandler struct {
	projectService services.ProjectService
}

func NewProjectHandler(projectService services.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectService: projectService}
}

func (h *ProjectHandler) Create(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	var req dto.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	userID := uuid.MustParse(userIDStr.(string))
	project, err := h.projectService.CreateProject(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.ProjectResponse{
		ID:          project.ID,
		UserID:      project.UserID,
		Name:        project.Name,
		Description: project.Description,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	})
}

func (h *ProjectHandler) Get(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid project ID format"})
		return
	}

	userID := uuid.MustParse(userIDStr.(string))
	project, role, err := h.projectService.GetProject(c.Request.Context(), userID, projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          project.ID,
		"user_id":     project.UserID,
		"name":        project.Name,
		"description": project.Description,
		"created_at":  project.CreatedAt,
		"updated_at":  project.UpdatedAt,
		"role":        role,
	})
}

func (h *ProjectHandler) List(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	userID := uuid.MustParse(userIDStr.(string))
	projects, roles, err := h.projectService.ListProjects(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	resp := make([]gin.H, len(projects))
	for i, project := range projects {
		resp[i] = gin.H{
			"id":          project.ID,
			"user_id":     project.UserID,
			"name":        project.Name,
			"description": project.Description,
			"created_at":  project.CreatedAt,
			"updated_at":  project.UpdatedAt,
			"role":        roles[project.ID],
		}
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ProjectHandler) Delete(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid project ID format"})
		return
	}

	userID := uuid.MustParse(userIDStr.(string))
	err = h.projectService.DeleteProject(c.Request.Context(), userID, projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "project deleted successfully"})
}

func (h *ProjectHandler) AddMember(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid project ID format"})
		return
	}

	var req dto.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	userID := uuid.MustParse(userIDStr.(string))
	member, err := h.projectService.AddMember(c.Request.Context(), userID, projectID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.MemberResponse{
		ID:        member.ID,
		ProjectID: member.ProjectID,
		UserID:    member.UserID,
		Email:     member.Email,
		Role:      member.Role,
		CreatedAt: member.CreatedAt,
	})
}

func (h *ProjectHandler) RemoveMember(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid project ID format"})
		return
	}

	memberIDStr := c.Param("memberId")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid member ID format"})
		return
	}

	userID := uuid.MustParse(userIDStr.(string))
	err = h.projectService.RemoveMember(c.Request.Context(), userID, projectID, memberID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "member removed successfully"})
}

func (h *ProjectHandler) UpdateMemberRole(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid project ID format"})
		return
	}

	memberID, err := uuid.Parse(c.Param("memberId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid member ID format"})
		return
	}

	var req dto.UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	userID := uuid.MustParse(userIDStr.(string))
	member, err := h.projectService.UpdateMemberRole(c.Request.Context(), userID, projectID, memberID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.MemberResponse{
		ID:        member.ID,
		ProjectID: member.ProjectID,
		UserID:    member.UserID,
		Email:     member.Email,
		Role:      member.Role,
		CreatedAt: member.CreatedAt,
	})
}

func (h *ProjectHandler) ListMembers(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid project ID format"})
		return
	}

	userID := uuid.MustParse(userIDStr.(string))
	members, err := h.projectService.ListMembers(c.Request.Context(), userID, projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	resp := make([]dto.MemberResponse, len(members))
	for i, m := range members {
		resp[i] = dto.MemberResponse{
			ID:        m.ID,
			ProjectID: m.ProjectID,
			UserID:    m.UserID,
			Email:     m.Email,
			Role:      m.Role,
			CreatedAt: m.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ProjectHandler) InviteMember(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid project ID format"})
		return
	}

	var req dto.InviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	userID := uuid.MustParse(userIDStr.(string))
	invite, err := h.projectService.InviteMember(c.Request.Context(), userID, projectID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.InviteResponse{
		ID:        invite.ID,
		ProjectID: invite.ProjectID,
		Email:     invite.Email,
		Role:      invite.Role,
		Status:    invite.Status,
		ExpiresAt: invite.ExpiresAt,
		CreatedAt: invite.CreatedAt,
	})
}

func (h *ProjectHandler) AcceptInvite(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	var req dto.AcceptInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	userID := uuid.MustParse(userIDStr.(string))
	resp, err := h.projectService.AcceptInvite(c.Request.Context(), userID, req.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ProjectHandler) ListInvites(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid project ID format"})
		return
	}

	userID := uuid.MustParse(userIDStr.(string))
	invites, err := h.projectService.ListInvites(c.Request.Context(), userID, projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	resp := make([]dto.InviteResponse, len(invites))
	for i, inv := range invites {
		resp[i] = dto.InviteResponse{
			ID:        inv.ID,
			ProjectID: inv.ProjectID,
			Email:     inv.Email,
			Role:      inv.Role,
			Status:    inv.Status,
			ExpiresAt: inv.ExpiresAt,
			CreatedAt: inv.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ProjectHandler) ResendInvite(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid project ID format"})
		return
	}
	inviteID, err := uuid.Parse(c.Param("inviteId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid invite ID format"})
		return
	}
	userID := uuid.MustParse(userIDStr.(string))
	invite, err := h.projectService.ResendInvite(c.Request.Context(), userID, projectID, inviteID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.InviteResponse{
		ID:        invite.ID,
		ProjectID: invite.ProjectID,
		Email:     invite.Email,
		Role:      invite.Role,
		Status:    invite.Status,
		ExpiresAt: invite.ExpiresAt,
		CreatedAt: invite.CreatedAt,
	})
}

func (h *ProjectHandler) RevokeInvite(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid project ID format"})
		return
	}

	inviteIDStr := c.Param("inviteId")
	inviteID, err := uuid.Parse(inviteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid invite ID format"})
		return
	}

	userID := uuid.MustParse(userIDStr.(string))
	err = h.projectService.RevokeInvite(c.Request.Context(), userID, projectID, inviteID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "invite revoked successfully"})
}

func (h *ProjectHandler) ListMyInvites(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	userID := uuid.MustParse(userIDStr.(string))
	invites, err := h.projectService.ListMyInvites(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	resp := make([]dto.InviteResponse, len(invites))
	for i, inv := range invites {
		resp[i] = dto.InviteResponse{
			ID:        inv.ID,
			ProjectID: inv.ProjectID,
			Email:     inv.Email,
			Role:      inv.Role,
			Status:    inv.Status,
			ExpiresAt: inv.ExpiresAt,
			CreatedAt: inv.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ProjectHandler) ListAuditEvents(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid project ID format"})
		return
	}
	userID := uuid.MustParse(userIDStr.(string))
	events, err := h.projectService.ListAuditEvents(c.Request.Context(), userID, projectID, 50, 0)
	if err != nil {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: err.Error()})
		return
	}
	resp := make([]dto.AuditEventResponse, len(events))
	for i, event := range events {
		resp[i] = dto.AuditEventResponse{
			ID:         event.ID,
			ProjectID:  event.ProjectID,
			ActorID:    event.ActorID,
			Action:     event.Action,
			TargetType: event.TargetType,
			TargetID:   event.TargetID,
			Metadata:   event.Metadata,
			CreatedAt:  event.CreatedAt,
		}
	}
	c.JSON(http.StatusOK, resp)
}
