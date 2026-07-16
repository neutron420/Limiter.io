package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"limiter.io/internal/config"
	"limiter.io/internal/dto"
	"limiter.io/internal/mailer"
	"limiter.io/internal/models"
	"limiter.io/internal/repository"
	"limiter.io/internal/utils"

	"github.com/google/uuid"
)

type ProjectService interface {
	CreateProject(ctx context.Context, userID uuid.UUID, req dto.CreateProjectRequest) (*models.Project, error)
	GetProject(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) (*models.Project, string, error)
	ListProjects(ctx context.Context, userID uuid.UUID) ([]models.Project, map[uuid.UUID]string, error)
	DeleteProject(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) error
	AddMember(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, req dto.AddMemberRequest) (*models.ProjectMember, error)
	RemoveMember(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, memberID uuid.UUID) error
	UpdateMemberRole(ctx context.Context, userID, projectID, memberID uuid.UUID, req dto.UpdateMemberRoleRequest) (*models.ProjectMember, error)
	ListMembers(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) ([]models.ProjectMember, error)
	InviteMember(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, req dto.InviteMemberRequest) (*models.ProjectInvite, error)
	ResendInvite(ctx context.Context, userID, projectID, inviteID uuid.UUID) (*models.ProjectInvite, error)
	AcceptInvite(ctx context.Context, userID uuid.UUID, token string) (*dto.AcceptInviteResponse, error)
	ListInvites(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) ([]models.ProjectInvite, error)
	RevokeInvite(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, inviteID uuid.UUID) error
	ListMyInvites(ctx context.Context, userID uuid.UUID) ([]models.ProjectInvite, error)
	ListAuditEvents(ctx context.Context, userID, projectID uuid.UUID, limit, offset int) ([]models.ProjectAuditEvent, error)
	CleanupExpiredInvites(ctx context.Context) (int64, error)
}

type projectService struct {
	projectRepo repository.ProjectRepository
	subRepo     repository.SubscriptionRepository
	memberRepo  repository.ProjectMemberRepository
	userRepo    repository.UserRepository
	inviteRepo  repository.ProjectInviteRepository
	auditRepo   repository.ProjectAuditRepository
	mailer      mailer.Mailer
	cfg         *config.Config
}

func NewProjectService(
	projectRepo repository.ProjectRepository,
	subRepo repository.SubscriptionRepository,
	memberRepo repository.ProjectMemberRepository,
	userRepo repository.UserRepository,
	inviteRepo repository.ProjectInviteRepository,
	auditRepo repository.ProjectAuditRepository,
	mailer mailer.Mailer,
	cfg *config.Config,
) ProjectService {
	return &projectService{
		projectRepo: projectRepo,
		subRepo:     subRepo,
		memberRepo:  memberRepo,
		userRepo:    userRepo,
		inviteRepo:  inviteRepo,
		auditRepo:   auditRepo,
		mailer:      mailer,
		cfg:         cfg,
	}
}

func (s *projectService) CreateProject(ctx context.Context, userID uuid.UUID, req dto.CreateProjectRequest) (*models.Project, error) {
	// Retrieve active subscription details
	sub, err := s.subRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("subscription not found for user")
	}

	// Count user's current projects
	count, err := s.projectRepo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Verify plan limit (e.g. Free plan limits to 3 projects)
	if sub.Plan.MaxProjects != -1 && count >= int64(sub.Plan.MaxProjects) {
		return nil, errors.New("You have used your limit.")
	}

	project := &models.Project{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *projectService) GetProject(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) (*models.Project, string, error) {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, "", errors.New("project not found")
	}

	role := roleForProject(ctx, s.projectRepo, s.memberRepo, userID, projectID)
	if role == "" {
		return nil, "", errors.New("unauthorized to access this project")
	}

	return project, role, nil
}

func (s *projectService) ListProjects(ctx context.Context, userID uuid.UUID) ([]models.Project, map[uuid.UUID]string, error) {
	owned, err := s.projectRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	memberProjectIDs, err := s.memberRepo.ListProjectIDsByUser(ctx, userID)
	if err != nil || len(memberProjectIDs) == 0 {
		roles := make(map[uuid.UUID]string)
		for _, proj := range owned {
			roles[proj.ID] = "owner"
		}
		return owned, roles, nil
	}

	memberProjects, err := s.projectRepo.ListByIDs(ctx, memberProjectIDs)
	if err != nil {
		roles := make(map[uuid.UUID]string)
		for _, proj := range owned {
			roles[proj.ID] = "owner"
		}
		return owned, roles, nil
	}

	// Build roles map
	roles := make(map[uuid.UUID]string)
	for _, proj := range owned {
		roles[proj.ID] = "owner"
	}

	for _, proj := range memberProjects {
		role := roleForProject(ctx, s.projectRepo, s.memberRepo, userID, proj.ID)
		roles[proj.ID] = role
	}

	return append(owned, memberProjects...), roles, nil
}

func (s *projectService) DeleteProject(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) error {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return errors.New("project not found")
	}

	if project.UserID != userID {
		return errors.New("unauthorized to delete this project")
	}

	err = s.projectRepo.Delete(ctx, projectID)
	if err == nil {
		s.recordAudit(ctx, projectID, userID, "project.deleted", "project", projectID, models.JSONMap{"name": project.Name})
	}
	return err
}

func (s *projectService) AddMember(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, req dto.AddMemberRequest) (*models.ProjectMember, error) {
	proj, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}
	if proj.UserID != userID {
		return nil, errors.New("only the project owner can manage team members")
	}

	targetUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("user with this email not found")
	}

	if targetUser.ID == userID {
		return nil, errors.New("cannot invite the owner of the project")
	}

	isMem, _ := s.memberRepo.IsMember(ctx, projectID, targetUser.ID)
	if isMem {
		return nil, errors.New("user is already a member of this project")
	}

	member := &models.ProjectMember{
		ID:        uuid.New(),
		ProjectID: projectID,
		UserID:    targetUser.ID,
		Email:     targetUser.Email,
		Role:      req.Role,
	}

	if err := s.memberRepo.Add(ctx, member); err != nil {
		return nil, err
	}

	return member, nil
}

func (s *projectService) RemoveMember(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, memberID uuid.UUID) error {
	role := roleForProject(ctx, s.projectRepo, s.memberRepo, userID, projectID)
	if role != "owner" && role != "admin" {
		return errors.New("insufficient role: only owners and admins can remove members")
	}

	// Admin cannot remove the owner
	member, err := s.memberRepo.ListByProject(ctx, projectID)
	if err == nil {
		for _, m := range member {
			if m.ID == memberID && m.Role == "owner" {
				return errors.New("cannot remove the project owner")
			}
		}
	}

	err = s.memberRepo.Remove(ctx, projectID, memberID)
	if err == nil {
		s.recordAudit(ctx, projectID, userID, "member.removed", "member", memberID, nil)
	}
	return err
}

func (s *projectService) UpdateMemberRole(ctx context.Context, userID, projectID, memberID uuid.UUID, req dto.UpdateMemberRoleRequest) (*models.ProjectMember, error) {
	role := roleForProject(ctx, s.projectRepo, s.memberRepo, userID, projectID)
	if role != "owner" && role != "admin" {
		return nil, errors.New("insufficient role: only owners and admins can update member roles")
	}
	if req.Role != "admin" && req.Role != "member" {
		return nil, errors.New("invalid role")
	}

	members, err := s.memberRepo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	for _, member := range members {
		if member.ID != memberID {
			continue
		}
		if member.Role == req.Role {
			return &member, nil
		}
		if err := s.memberRepo.UpdateRole(ctx, projectID, memberID, req.Role); err != nil {
			return nil, err
		}
		member.Role = req.Role
		s.recordAudit(ctx, projectID, userID, "member.role_updated", "member", memberID, models.JSONMap{"email": member.Email, "role": req.Role})
		return &member, nil
	}

	return nil, errors.New("member not found")
}

func (s *projectService) ListMembers(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) ([]models.ProjectMember, error) {
	proj, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}

	isOwner := proj.UserID == userID
	isMem, _ := s.memberRepo.IsMember(ctx, projectID, userID)
	if !isOwner && !isMem {
		return nil, errors.New("unauthorized to view members for this project")
	}

	return s.memberRepo.ListByProject(ctx, projectID)
}

func (s *projectService) InviteMember(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, req dto.InviteMemberRequest) (*models.ProjectInvite, error) {
	proj, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}

	role := roleForProject(ctx, s.projectRepo, s.memberRepo, userID, projectID)
	if role != "owner" && role != "admin" {
		return nil, errors.New("insufficient role: only owners and admins can send invitations")
	}

	// Validate role
	if req.Role != "admin" && req.Role != "member" {
		return nil, errors.New("invalid role")
	}

	// Reject inviting the owner's own email
	caller, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("failed to get caller info")
	}
	if strings.EqualFold(caller.Email, req.Email) {
		return nil, errors.New("cannot invite yourself")
	}

	// Check if target user is already a member
	targetUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil {
		// User exists, check if already a member
		memberProjectIDs, err := s.memberRepo.ListProjectIDsByUser(ctx, targetUser.ID)
		if err == nil {
			for _, pid := range memberProjectIDs {
				if pid == projectID {
					return nil, errors.New("user is already a member of this project")
				}
			}
		}
	}

	// Revoke any previous pending invite for the same email+project
	pendingInvites, err := s.inviteRepo.ListPendingByEmail(ctx, req.Email)
	if err == nil {
		for _, inv := range pendingInvites {
			if inv.ProjectID == projectID {
				inv.Status = "revoked"
				s.inviteRepo.Update(ctx, &inv)
			}
		}
	}

	// Generate invite token
	rawToken, err := utils.GenerateRandomToken(32)
	if err != nil {
		return nil, errors.New("failed to generate invite token")
	}
	tokenHash := utils.HashAPIKey(rawToken)

	// Create invite (7-day expiry)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	invite := &models.ProjectInvite{
		ID:        uuid.New(),
		ProjectID: projectID,
		Email:     req.Email,
		Role:      req.Role,
		TokenHash: tokenHash,
		InvitedBy: userID,
		Status:    "pending",
		ExpiresAt: expiresAt,
	}

	if err := s.inviteRepo.Create(ctx, invite); err != nil {
		return nil, errors.New("failed to create invite")
	}

	s.recordAudit(ctx, projectID, userID, "invite.created", "invite", invite.ID, models.JSONMap{"email": req.Email, "role": req.Role})

	// Send invitation email
	inviteURL := fmt.Sprintf("%s/accept-invite?token=%s", strings.TrimRight(s.cfg.AppBaseURL, "/"), rawToken)
	subject := fmt.Sprintf("You've been invited to %q on Limiter.io", proj.Name)
	htmlBody := s.buildInviteEmailHTML(proj.Name, caller.Email, req.Role, inviteURL, rawToken)

	if err := s.mailer.Send(ctx, req.Email, subject, htmlBody); err != nil {
		// Log error but don't fail the invite creation
		fmt.Printf("Failed to send invitation email: %v\n", err)
	}

	return invite, nil
}

func (s *projectService) AcceptInvite(ctx context.Context, userID uuid.UUID, rawToken string) (*dto.AcceptInviteResponse, error) {
	// Get user's email
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Hash token and look up invite
	tokenHash := utils.HashAPIKey(rawToken)
	invite, err := s.inviteRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, errors.New("invalid or expired invite token")
	}

	// Check invite status
	if invite.Status != "pending" {
		return nil, errors.New("invite has already been used or revoked")
	}

	// Check expiry
	if time.Now().After(invite.ExpiresAt) {
		return nil, errors.New("invite has expired")
	}

	// Email must match (case-insensitive)
	if !strings.EqualFold(user.Email, invite.Email) {
		return nil, errors.New("this invite was sent to a different email address")
	}

	// Get project
	proj, err := s.projectRepo.GetByID(ctx, invite.ProjectID)
	if err != nil {
		return nil, errors.New("project not found")
	}

	// Create the ProjectMember row
	member := &models.ProjectMember{
		ID:        uuid.New(),
		ProjectID: invite.ProjectID,
		UserID:    userID,
		Email:     user.Email,
		Role:      invite.Role,
	}

	if err := s.memberRepo.Add(ctx, member); err != nil {
		return nil, errors.New("failed to add member")
	}

	s.recordAudit(ctx, invite.ProjectID, userID, "invite.accepted", "invite", invite.ID, models.JSONMap{"role": invite.Role})

	// Update invite status
	invite.Status = "accepted"
	now := time.Now()
	invite.AcceptedAt = &now
	if err := s.inviteRepo.Update(ctx, invite); err != nil {
		fmt.Printf("Failed to update invite status: %v\n", err)
	}

	// Send notification email to project owner
	owner, err := s.userRepo.GetByID(ctx, proj.UserID)
	if err == nil {
		subject := fmt.Sprintf("%s joined %q as %s", user.Email, proj.Name, invite.Role)
		htmlBody := s.buildAcceptedEmailHTML(proj.Name, user.Email, invite.Role)
		if err := s.mailer.Send(ctx, owner.Email, subject, htmlBody); err != nil {
			fmt.Printf("Failed to send accepted notification email: %v\n", err)
		}
	}

	return &dto.AcceptInviteResponse{
		ProjectID:   proj.ID,
		ProjectName: proj.Name,
		Role:        invite.Role,
	}, nil
}

func (s *projectService) ListInvites(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) ([]models.ProjectInvite, error) {
	role := roleForProject(ctx, s.projectRepo, s.memberRepo, userID, projectID)
	if role != "owner" && role != "admin" {
		return nil, errors.New("insufficient role: only owners and admins can view pending invitations")
	}

	return s.inviteRepo.ListByProject(ctx, projectID)
}

func (s *projectService) RevokeInvite(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, inviteID uuid.UUID) error {
	role := roleForProject(ctx, s.projectRepo, s.memberRepo, userID, projectID)
	if role != "owner" && role != "admin" {
		return errors.New("insufficient role: only owners and admins can revoke invitations")
	}

	// Get the invite
	invite, err := s.inviteRepo.GetByID(ctx, inviteID)
	if err != nil {
		return errors.New("invite not found")
	}

	// Check it belongs to this project
	if invite.ProjectID != projectID {
		return errors.New("invite does not belong to this project")
	}

	// Update status to revoked
	invite.Status = "revoked"
	if err := s.inviteRepo.Update(ctx, invite); err != nil {
		return err
	}
	s.recordAudit(ctx, projectID, userID, "invite.revoked", "invite", inviteID, nil)
	return nil
}

func (s *projectService) ResendInvite(ctx context.Context, userID, projectID, inviteID uuid.UUID) (*models.ProjectInvite, error) {
	proj, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}
	role := roleForProject(ctx, s.projectRepo, s.memberRepo, userID, projectID)
	if role != "owner" && role != "admin" {
		return nil, errors.New("insufficient role: only owners and admins can resend invitations")
	}
	invite, err := s.inviteRepo.GetByID(ctx, inviteID)
	if err != nil {
		return nil, errors.New("invite not found")
	}
	if invite.ProjectID != projectID {
		return nil, errors.New("invite does not belong to this project")
	}
	if invite.Status != "pending" {
		return nil, errors.New("can only resend pending invites")
	}
	caller, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("failed to get caller info")
	}
	rawToken, err := utils.GenerateRandomToken(32)
	if err != nil {
		return nil, errors.New("failed to generate invite token")
	}
	tokenHash := utils.HashAPIKey(rawToken)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	invite.TokenHash = tokenHash
	invite.ExpiresAt = expiresAt
	invite.InvitedBy = userID
	invite.CreatedAt = time.Now()
	if err := s.inviteRepo.Update(ctx, invite); err != nil {
		return nil, errors.New("failed to update invite")
	}
	s.recordAudit(ctx, projectID, userID, "invite.resend", "invite", invite.ID, models.JSONMap{"email": invite.Email, "role": invite.Role})
	inviteURL := fmt.Sprintf("%s/accept-invite?token=%s", strings.TrimRight(s.cfg.AppBaseURL, "/"), rawToken)
	subject := fmt.Sprintf("Reminder: You've been invited to %q on Limiter.io", proj.Name)
	htmlBody := s.buildInviteEmailHTML(proj.Name, caller.Email, invite.Role, inviteURL, rawToken)
	if err := s.mailer.Send(ctx, invite.Email, subject, htmlBody); err != nil {
		fmt.Printf("Failed to resend invitation email: %v\n", err)
	}
	return invite, nil
}

func (s *projectService) CleanupExpiredInvites(ctx context.Context) (int64, error) {
	expired, err := s.inviteRepo.ListExpired(ctx)
	if err != nil {
		return 0, err
	}
	for _, inv := range expired {
		inv.Status = "expired"
		_ = s.inviteRepo.Update(ctx, &inv)
	}
	return int64(len(expired)), nil
}

func (s *projectService) ListMyInvites(ctx context.Context, userID uuid.UUID) ([]models.ProjectInvite, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return s.inviteRepo.ListPendingByEmail(ctx, user.Email)
}

func (s *projectService) recordAudit(ctx context.Context, projectID, actorID uuid.UUID, action, targetType string, targetID uuid.UUID, metadata models.JSONMap) {
	if s.auditRepo == nil {
		return
	}
	_ = s.auditRepo.Create(ctx, &models.ProjectAuditEvent{ProjectID: projectID, ActorID: actorID, Action: action, TargetType: targetType, TargetID: targetID, Metadata: metadata})
}

func (s *projectService) ListAuditEvents(ctx context.Context, userID, projectID uuid.UUID, limit, offset int) ([]models.ProjectAuditEvent, error) {
	role := roleForProject(ctx, s.projectRepo, s.memberRepo, userID, projectID)
	if role != "owner" && role != "admin" {
		return nil, errors.New("insufficient role: only owners and admins can view audit events")
	}
	return s.auditRepo.ListByProject(ctx, projectID, limit, offset)
}

// Role helper functions for RBAC (package-level for shared use)

// RoleForProject returns the user's role for a project: "owner", "admin", "member", or "" (no access)
func RoleForProject(ctx context.Context, projectRepo repository.ProjectRepository, memberRepo repository.ProjectMemberRepository, userID, projectID uuid.UUID) string {
	proj, err := projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return ""
	}

	// Check if owner
	if proj.UserID == userID {
		return "owner"
	}

	// Check if member
	member, err := memberRepo.IsMember(ctx, projectID, userID)
	if err != nil || !member {
		return ""
	}

	// Get member's role
	members, err := memberRepo.ListByProject(ctx, projectID)
	if err != nil {
		return ""
	}

	for _, m := range members {
		if m.UserID == userID {
			return m.Role
		}
	}

	return ""
}

// CanRead returns true if the role has read access
func CanRead(role string) bool {
	return role != ""
}

// CanWrite returns true if the role has write access
func CanWrite(role string) bool {
	return role == "owner" || role == "admin"
}

// Internal lowercase versions for use within the package
func roleForProject(ctx context.Context, projectRepo repository.ProjectRepository, memberRepo repository.ProjectMemberRepository, userID, projectID uuid.UUID) string {
	return RoleForProject(ctx, projectRepo, memberRepo, userID, projectID)
}

func canRead(role string) bool {
	return CanRead(role)
}

func canWrite(role string) bool {
	return CanWrite(role)
}

func (s *projectService) buildInviteEmailHTML(projectName, inviterEmail, role, inviteURL, rawToken string) string {
	roleDesc := "read-only member"
	if role == "admin" {
		roleDesc = "admin (read-write)"
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>Project Invitation</title>
	<style>
		body { font-family: 'Courier New', monospace; background-color: #0a0a0a; color: #e5e5e5; padding: 40px; }
		.container { max-width: 600px; margin: 0 auto; border: 1px solid #333; padding: 30px; }
		h1 { color: #ea580c; margin-bottom: 20px; }
		p { line-height: 1.6; margin-bottom: 15px; }
		.button { display: inline-block; background-color: #ea580c; color: white; padding: 12px 24px; text-decoration: none; margin: 20px 0; }
		.button:hover { background-color: #c2410c; }
		.footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #333; font-size: 12px; color: #666; }
		.url { color: #ea580c; word-break: break-all; }
	</style>
</head>
<body>
	<div class="container">
		<h1>You've been invited to "%s"</h1>
		<p><strong>%s</strong> has invited you to join as a <strong>%s</strong>.</p>
		<p>This invitation expires in 7 days.</p>
		<a href="%s" class="button">Accept Invite</a>
		<p>If the button doesn't work, copy and paste this URL into your browser:</p>
		<p class="url">%s</p>
		<p>If you do not wish to accept this invitation, you can safely ignore this email.</p>
		<div class="footer">
			<p>This is an automated message from Limiter.io.</p>
		</div>
	</div>
</body>
</html>`, projectName, inviterEmail, roleDesc, inviteURL, inviteURL)
}

func (s *projectService) buildAcceptedEmailHTML(projectName, newMemberEmail, role string) string {
	roleDesc := "read-only member"
	if role == "admin" {
		roleDesc = "admin (read-write)"
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>Invite Accepted</title>
	<style>
		body { font-family: 'Courier New', monospace; background-color: #0a0a0a; color: #e5e5e5; padding: 40px; }
		.container { max-width: 600px; margin: 0 auto; border: 1px solid #333; padding: 30px; }
		h1 { color: #ea580c; margin-bottom: 20px; }
		p { line-height: 1.6; margin-bottom: 15px; }
		.footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #333; font-size: 12px; color: #666; }
		.email { color: #ea580c; }
	</style>
</head>
<body>
	<div class="container">
		<h1>Invite Accepted</h1>
		<p><strong class="email">%s</strong> has accepted your invitation to join <strong>%s</strong> as a <strong>%s</strong>.</p>
		<p>They now have access to the project according to their role permissions.</p>
		<div class="footer">
			<p>This is an automated message from Limiter.io.</p>
		</div>
	</div>
</body>
</html>`, newMemberEmail, projectName, roleDesc)
}
