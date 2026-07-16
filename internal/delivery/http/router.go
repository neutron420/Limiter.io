package http

import (
	"strings"

	_ "limiter.io/docs" // Swagger documentation init
	"limiter.io/internal/config"
	"limiter.io/internal/handlers"
	"limiter.io/internal/kafka"
	"limiter.io/internal/middleware"
	"limiter.io/internal/ratelimiter"
	"limiter.io/internal/repository"
	internalws "limiter.io/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// parseCORSOrigins splits the comma-separated CORS_ALLOWED_ORIGINS config into
// a trimmed slice. An empty value falls back to "*" (allow any).
func parseCORSOrigins(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{"*"}
	}
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			origins = append(origins, t)
		}
	}
	if len(origins) == 0 {
		return []string{"*"}
	}
	return origins
}

func originAllowed(origin string, allowed []string) bool {
	for _, a := range allowed {
		if a == origin {
			return true
		}
	}
	return false
}

type RouterConfig struct {
	Engine              *gin.Engine
	Cfg                 *config.Config
	AuthHandler         *handlers.AuthHandler
	ProjHandler         *handlers.ProjectHandler
	KeyHandler          *handlers.APIKeyHandler
	PolicyHandler       *handlers.PolicyHandler
	SubHandler          *handlers.SubscriptionHandler
	AnalHandler         *handlers.AnalyticsHandler
	HealthHandler       *handlers.HealthHandler
	WSHandler           *handlers.WSHandler
	BillingHandler      *handlers.BillingHandler
	SecurityHandler     *handlers.SecurityHandler
	IPAccessHandler     *handlers.IPAccessHandler
	NotifHandler         *handlers.NotificationHandler
	OrgHandler           *handlers.OrganizationHandler
	ApprovalHandler     *handlers.ApprovalHandler
	QuotaHandler         *handlers.QuotaHandler
	TenantHandler        *handlers.TenantHandler
	PasskeyHandler       *handlers.PasskeyHandler
	ImmutableAuditHandler *handlers.ImmutableAuditHandler
	StatusHandler       *handlers.StatusHandler
	SandboxHandler      *handlers.SandboxHandler
	MaintenanceHandler  *handlers.MaintenanceHandler
	SSOHandler          *handlers.SSOHandler
	Hub                 *internalws.Hub
	Limiter             ratelimiter.RateLimiter
	RuleRepo            repository.RateLimitRuleRepository
	KeyRepo             repository.APIKeyRepository
	CacheRepo           repository.CacheRepository
	AnalRepo            repository.AnalyticsRepository
	Producer            kafka.Producer
	RedisClient         *redis.Client
}

func SetupRouter(c RouterConfig) {
	// Global Middlewares
	c.Engine.Use(middleware.RequestID())
	c.Engine.Use(middleware.Recovery(nil))   // nil logger will default to internal behavior or standard out
	c.Engine.Use(middleware.Logger(zap.L())) // structured per-request logging (uses global zap logger)
	c.Engine.Use(middleware.Metrics())

	// CORS Setup — origins are configurable via CORS_ALLOWED_ORIGINS
	// ("*" for any, or a comma-separated allowlist). Credentials are only
	// enabled for an explicit allowlist (browsers reject "*" + credentials).
	allowedOrigins := parseCORSOrigins(c.Cfg.CORSAllowedOrigins)
	allowAny := len(allowedOrigins) == 1 && allowedOrigins[0] == "*"

	c.Engine.Use(func(ctx *gin.Context) {
		origin := ctx.GetHeader("Origin")
		if allowAny {
			ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else if origin != "" && originAllowed(origin, allowedOrigins) {
			ctx.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			ctx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			ctx.Writer.Header().Add("Vary", "Origin")
		}
		ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-API-Key")
		ctx.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(204)
			return
		}
		ctx.Next()
	})

	// Security headers for all routes
	c.Engine.Use(middleware.SecurityHeaders())

	// Health and Metrics endpoints
	c.Engine.GET("/healthz", c.HealthHandler.Liveness)
	c.Engine.GET("/readyz", c.HealthHandler.Readiness)
	c.Engine.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swagger documentation
	c.Engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// REST API v1 Group
	v1 := c.Engine.Group("/api/v1")
	{
		// Public Auth Routes (with rate limiting to prevent brute force)
		auth := v1.Group("/auth")
		auth.Use(middleware.RateLimitAuth(c.Cfg))
		{
			auth.POST("/register", c.AuthHandler.Register)
			auth.POST("/login", c.AuthHandler.Login)
			auth.POST("/login/mfa", c.AuthHandler.LoginMFA)
			auth.POST("/refresh", c.AuthHandler.Refresh)
			auth.POST("/forgot-password", c.AuthHandler.ForgotPassword)
			auth.POST("/reset-password", c.AuthHandler.ResetPassword)
			auth.POST("/google", c.AuthHandler.LoginWithGoogle)
		}

		// Billing Webhooks (Public endpoint for Lemon Squeezy callback)
		v1.POST("/billing/webhook", c.BillingHandler.LemonSqueezyWebhook)

		// Private JWT Sessions Group
		private := v1.Group("")
		private.Use(middleware.JWTAuth(c.Cfg.JWTSecret))
		{
			// Auth
			private.POST("/auth/logout", c.AuthHandler.Logout)
			private.POST("/auth/change-password", c.AuthHandler.ChangePassword)

			// MFA (TOTP)
			private.POST("/auth/mfa/setup", c.SecurityHandler.SetupMFA)
			private.POST("/auth/mfa/verify", c.SecurityHandler.VerifyAndEnableMFA)
			private.POST("/auth/mfa/disable", c.SecurityHandler.DisableMFA)

			// Session management
			private.GET("/auth/sessions", c.SecurityHandler.ListSessions)
			private.DELETE("/auth/sessions/:sessionId", c.SecurityHandler.RevokeSession)

			// Projects
			private.POST("/projects", c.ProjHandler.Create)
			private.GET("/projects", c.ProjHandler.List)
			private.GET("/projects/:projectId", c.ProjHandler.Get)
			private.DELETE("/projects/:projectId", c.ProjHandler.Delete)

			// Project Members
			private.GET("/projects/:projectId/members", c.ProjHandler.ListMembers)
			private.POST("/projects/:projectId/members", c.ProjHandler.AddMember)
			private.PATCH("/projects/:projectId/members/:memberId", c.ProjHandler.UpdateMemberRole)
			private.DELETE("/projects/:projectId/members/:memberId", c.ProjHandler.RemoveMember)

			// Project Invites
			private.POST("/projects/:projectId/invites", c.ProjHandler.InviteMember)
			private.GET("/projects/:projectId/invites", c.ProjHandler.ListInvites)
			private.GET("/projects/:projectId/audit-events", c.ProjHandler.ListAuditEvents)
			private.POST("/projects/:projectId/invites/:inviteId/resend", c.ProjHandler.ResendInvite)
			private.DELETE("/projects/:projectId/invites/:inviteId", c.ProjHandler.RevokeInvite)
			private.GET("/invites", c.ProjHandler.ListMyInvites)
			private.POST("/invites/accept", c.ProjHandler.AcceptInvite)

			// API Keys
			private.POST("/projects/:projectId/keys", c.KeyHandler.Create)
			private.GET("/projects/:projectId/keys", c.KeyHandler.List)
			private.POST("/projects/:projectId/keys/:keyId/rotate", c.KeyHandler.Rotate)
			private.POST("/projects/:projectId/keys/:keyId/revoke", c.KeyHandler.Revoke)
			private.DELETE("/projects/:projectId/keys/:keyId", c.KeyHandler.Delete)

			// Rate Limiting Rules (Policies)
			private.POST("/projects/:projectId/rules", c.PolicyHandler.Create)
			private.GET("/projects/:projectId/rules", c.PolicyHandler.List)
			private.GET("/projects/:projectId/rules/:ruleId", c.PolicyHandler.Get)
			private.PUT("/projects/:projectId/rules/:ruleId", c.PolicyHandler.Update)
			private.DELETE("/projects/:projectId/rules/:ruleId", c.PolicyHandler.Delete)
			private.POST("/projects/:projectId/rules/:ruleId/simulate", c.PolicyHandler.Simulate)

			// Analytics
			private.GET("/projects/:projectId/analytics/stats", c.AnalHandler.GetStats)
			private.GET("/projects/:projectId/analytics/logs", c.AnalHandler.GetLogs)
			private.GET("/projects/:projectId/analytics/export", c.AnalHandler.ExportLogs)
			private.GET("/projects/:projectId/analytics/timeseries", c.AnalHandler.GetTimeSeries)

			// Real-time WebSocket analytics stream
			private.GET("/projects/:projectId/ws", c.WSHandler.Connect)

			// Subscription
			private.GET("/subscription", c.SubHandler.Get)
			private.GET("/subscription/usage", c.SubHandler.GetUsage)
			private.POST("/subscription/upgrade", c.SubHandler.Upgrade)

			// Billing webhooks audit log
			private.GET("/billing/webhooks", c.BillingHandler.ListWebhooks)

			// IP Access Rules (allowlist/denylist)
			private.GET("/projects/:projectId/ip-rules", c.IPAccessHandler.List)
			private.POST("/projects/:projectId/ip-rules", c.IPAccessHandler.Create)
			private.DELETE("/projects/:projectId/ip-rules/:ruleId", c.IPAccessHandler.Delete)

		// Notification Preferences
		private.GET("/projects/:projectId/notification-preferences", c.NotifHandler.GetPreferences)
		private.PUT("/projects/:projectId/notification-preferences", c.NotifHandler.UpdatePreferences)

		// Organizations
		private.POST("/organizations", c.OrgHandler.Create)
		private.GET("/organizations", c.OrgHandler.ListByUser)
		private.GET("/organizations/:orgId", c.OrgHandler.GetByID)
		private.POST("/organizations/:orgId/members", c.OrgHandler.AddMember)
		private.GET("/organizations/:orgId/members", c.OrgHandler.ListMembers)
		private.DELETE("/organizations/:orgId/members/:userId", c.OrgHandler.RemoveMember)
		private.POST("/organizations/:orgId/groups", c.OrgHandler.CreateGroup)
		private.GET("/organizations/:orgId/groups", c.OrgHandler.ListGroups)
		private.DELETE("/organizations/:orgId/groups/:groupId", c.OrgHandler.DeleteGroup)
		private.POST("/organizations/:orgId/groups/:groupId/members", c.OrgHandler.AddToGroup)
		private.DELETE("/organizations/:orgId/groups/:groupId/members/:userId", c.OrgHandler.RemoveFromGroup)

		// Approval Workflows
		private.POST("/organizations/:orgId/approval-workflows", c.ApprovalHandler.CreateWorkflow)
		private.GET("/organizations/:orgId/approval-workflows", c.ApprovalHandler.ListWorkflows)
		private.POST("/organizations/:orgId/approval-requests", c.ApprovalHandler.RequestApproval)
		private.GET("/organizations/:orgId/approval-requests", c.ApprovalHandler.ListRequests)
		private.POST("/approval-requests/:id/approve", c.ApprovalHandler.Approve)
		private.POST("/approval-requests/:id/reject", c.ApprovalHandler.Reject)
		private.GET("/approval-requests/:id", c.ApprovalHandler.GetRequest)

		// Quotas
		private.GET("/projects/:projectId/quotas", c.QuotaHandler.GetQuota)
		private.PUT("/projects/:projectId/quotas", c.QuotaHandler.SetQuota)
		private.GET("/projects/:projectId/quotas/check", c.QuotaHandler.CheckQuota)

		// Tenant Configs
		private.GET("/projects/:projectId/tenants", c.TenantHandler.List)
		private.POST("/projects/:projectId/tenants", c.TenantHandler.Create)
		private.GET("/projects/:projectId/tenants/:tenantId", c.TenantHandler.Get)
		private.PUT("/projects/:projectId/tenants/:tenantId", c.TenantHandler.Update)
		private.DELETE("/projects/:projectId/tenants/:tenantId", c.TenantHandler.Delete)

		// Analytics (extended)
		private.GET("/projects/:projectId/analytics/data", c.AnalHandler.GetAnalyticsData)
		private.POST("/projects/:projectId/analytics/views", c.AnalHandler.SaveView)
		private.GET("/projects/:projectId/analytics/views", c.AnalHandler.ListViews)
		private.GET("/projects/:projectId/analytics/views/:viewId", c.AnalHandler.GetView)
		private.DELETE("/projects/:projectId/analytics/views/:viewId", c.AnalHandler.DeleteView)
		private.GET("/projects/:projectId/analytics/anomaly-config", c.AnalHandler.GetAnomalyConfig)
		private.PUT("/projects/:projectId/analytics/anomaly-config", c.AnalHandler.UpdateAnomalyConfig)
		private.POST("/projects/:projectId/analytics/detect-anomalies", c.AnalHandler.DetectAnomalies)

		// Passkeys (WebAuthn)
		private.POST("/auth/passkeys/register/begin", c.PasskeyHandler.BeginRegistration)
		private.POST("/auth/passkeys/register/complete", c.PasskeyHandler.CompleteRegistration)
		private.GET("/auth/passkeys", c.PasskeyHandler.ListPasskeys)
		private.DELETE("/auth/passkeys/:id", c.PasskeyHandler.DeletePasskey)

		// Immutable Audit Logs
		private.GET("/audit-logs", c.ImmutableAuditHandler.List)
		private.GET("/audit-logs/:id", c.ImmutableAuditHandler.GetByID)
		private.GET("/audit-logs/verify-chain", c.ImmutableAuditHandler.VerifyChain)

		// Sandbox
		private.POST("/sandbox/create", c.SandboxHandler.Create)
		private.POST("/sandbox/:projectId/cleanup", c.SandboxHandler.Cleanup)

		// SLA Configs
		private.GET("/organizations/:orgId/sla-config", c.BillingHandler.GetSLAConfig)
		private.PUT("/organizations/:orgId/sla-config", c.BillingHandler.UpdateSLAConfig)

		// Email Templates
		private.GET("/organizations/:orgId/email-templates", c.BillingHandler.GetEmailTemplate)
		private.PUT("/organizations/:orgId/email-templates", c.BillingHandler.SaveEmailTemplate)

		// Region Configs
		private.GET("/organizations/:orgId/regions", c.BillingHandler.ListRegionConfigs)
		private.PUT("/organizations/:orgId/regions", c.BillingHandler.SaveRegionConfig)
		private.GET("/organizations/:orgId/regions/:region", c.BillingHandler.GetRegionConfig)

		// Invoices
		private.GET("/projects/:projectId/invoices", c.BillingHandler.ListInvoices)
		private.GET("/projects/:projectId/invoices/:invoiceId", c.BillingHandler.GetInvoice)

		// Usage
		private.GET("/projects/:projectId/usage", c.BillingHandler.GetUsage)

		// SSO
		private.POST("/organizations/:orgId/sso/saml", c.SSOHandler.SetSAMLConfig)
		private.GET("/organizations/:orgId/sso/saml", c.SSOHandler.GetSAMLConfig)
		private.POST("/organizations/:orgId/sso/oidc", c.SSOHandler.SetOIDCConfig)
		private.GET("/organizations/:orgId/sso/oidc", c.SSOHandler.GetOIDCConfig)
	}

	// Public passkey login (no auth)
	v1.POST("/auth/passkeys/login/begin", c.PasskeyHandler.BeginLogin)
	v1.POST("/auth/passkeys/login/complete", c.PasskeyHandler.CompleteLogin)

	// Status endpoint (public)
	c.Engine.GET("/status", c.StatusHandler.Status)
	c.Engine.GET("/health", c.StatusHandler.Health)

	// Maintenance mode middleware (applied after status/health)
	c.Engine.Use(middleware.MaintenanceMiddleware())

	// Weighted request middleware (applied within gateway)
	// Developer API Gateway Simulation Endpoint
		// Protected by API Key validation & Rate Limiter middleware
		gateway := v1.Group("/gateway")
		gateway.Use(middleware.APIKeyAuth(c.KeyRepo, c.CacheRepo))
		gateway.Use(middleware.RateLimit(c.Limiter, c.RuleRepo, c.CacheRepo, c.Producer, c.AnalRepo, c.RedisClient, c.Hub))
		{
			// This wildcard maps any sub-path of /gateway and applies rate limiting rule matches
			gateway.Any("/*path", func(ctx *gin.Context) {
				resp := gin.H{
					"status":  "success",
					"message": "request passed through rate limiter gateway successfully",
					"path":    ctx.Param("path"),
				}
				// Include matched rule info if available (set by rate limit middleware)
				if ruleName, exists := ctx.Get("MatchedRuleName"); exists {
					resp["matched_rule"] = ruleName
				}
				if ruleAlgo, exists := ctx.Get("MatchedRuleAlgorithm"); exists {
					resp["matched_algorithm"] = ruleAlgo
				}
				if dryRunResult, exists := ctx.Get("DryRunResult"); exists {
					resp["dry_run_result"] = dryRunResult
				}
				ctx.JSON(200, resp)
			})
		}
	}
}
