package http

import (
	"strings"

	"limiter.io/internal/config"
	"limiter.io/internal/handlers"
	"limiter.io/internal/kafka"
	"limiter.io/internal/middleware"
	"limiter.io/internal/ratelimiter"
	"limiter.io/internal/repository"
	internalws "limiter.io/internal/websocket"
	_ "limiter.io/docs" // Swagger documentation init

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
	Engine        *gin.Engine
	Cfg           *config.Config
	AuthHandler   *handlers.AuthHandler
	ProjHandler   *handlers.ProjectHandler
	KeyHandler    *handlers.APIKeyHandler
	PolicyHandler *handlers.PolicyHandler
	SubHandler    *handlers.SubscriptionHandler
	AnalHandler   *handlers.AnalyticsHandler
	HealthHandler *handlers.HealthHandler
	WSHandler     *handlers.WSHandler
	BillingHandler *handlers.BillingHandler
	Hub           *internalws.Hub
	Limiter       ratelimiter.RateLimiter
	RuleRepo      repository.RateLimitRuleRepository
	KeyRepo       repository.APIKeyRepository
	CacheRepo     repository.CacheRepository
	AnalRepo      repository.AnalyticsRepository
	Producer      kafka.Producer
	RedisClient   *redis.Client
}

func SetupRouter(c RouterConfig) {
	// Global Middlewares
	c.Engine.Use(middleware.RequestID())
	c.Engine.Use(middleware.Recovery(nil)) // nil logger will default to internal behavior or standard out
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

	// Health and Metrics endpoints
	c.Engine.GET("/healthz", c.HealthHandler.Liveness)
	c.Engine.GET("/readyz", c.HealthHandler.Readiness)
	c.Engine.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swagger documentation
	c.Engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// REST API v1 Group
	v1 := c.Engine.Group("/api/v1")
	{
		// Public Auth Routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", c.AuthHandler.Register)
			auth.POST("/login", c.AuthHandler.Login)
			auth.POST("/refresh", c.AuthHandler.Refresh)
			auth.POST("/forgot-password", c.AuthHandler.ForgotPassword)
			auth.POST("/reset-password", c.AuthHandler.ResetPassword)
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

			// Projects
			private.POST("/projects", c.ProjHandler.Create)
			private.GET("/projects", c.ProjHandler.List)
			private.GET("/projects/:projectId", c.ProjHandler.Get)
			private.DELETE("/projects/:projectId", c.ProjHandler.Delete)

			// Project Members
			private.GET("/projects/:projectId/members", c.ProjHandler.ListMembers)
			private.POST("/projects/:projectId/members", c.ProjHandler.AddMember)
			private.DELETE("/projects/:projectId/members/:memberId", c.ProjHandler.RemoveMember)

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
			private.GET("/projects/:projectId/analytics/timeseries", c.AnalHandler.GetTimeSeries)

			// Real-time WebSocket analytics stream
			private.GET("/projects/:projectId/ws", c.WSHandler.Connect)

			// Subscription
			private.GET("/subscription", c.SubHandler.Get)
			private.GET("/subscription/usage", c.SubHandler.GetUsage)
			private.POST("/subscription/upgrade", c.SubHandler.Upgrade)

			// Billing webhooks audit log
			private.GET("/billing/webhooks", c.BillingHandler.ListWebhooks)
		}

		// Developer API Gateway Simulation Endpoint
		// Protected by API Key validation & Rate Limiter middleware
		gateway := v1.Group("/gateway")
		gateway.Use(middleware.APIKeyAuth(c.KeyRepo, c.CacheRepo))
		gateway.Use(middleware.RateLimit(c.Limiter, c.RuleRepo, c.CacheRepo, c.Producer, c.AnalRepo, c.RedisClient, c.Hub))
		{
			// This wildcard maps any sub-path of /gateway and applies rate limiting rule matches
			gateway.Any("/*path", func(ctx *gin.Context) {
				ctx.JSON(200, gin.H{
					"status":  "success",
					"message": "request passed through rate limiter gateway successfully",
					"path":    ctx.Param("path"),
				})
			})
		}
	}
}
