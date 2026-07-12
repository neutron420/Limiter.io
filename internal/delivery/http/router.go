package http

import (
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
)

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
	Producer      kafka.Producer
	RedisClient   *redis.Client
}

func SetupRouter(c RouterConfig) {
	// Global Middlewares
	c.Engine.Use(middleware.RequestID())
	c.Engine.Use(middleware.Recovery(nil)) // nil logger will default to internal behavior or standard out
	c.Engine.Use(middleware.Metrics())

	// CORS Setup
	c.Engine.Use(func(ctx *gin.Context) {
		ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		ctx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
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

			// Analytics
			private.GET("/projects/:projectId/analytics/stats", c.AnalHandler.GetStats)
			private.GET("/projects/:projectId/analytics/logs", c.AnalHandler.GetLogs)

			// Real-time WebSocket analytics stream
			private.GET("/projects/:projectId/ws", c.WSHandler.Connect)

			// Subscription
			private.GET("/subscription", c.SubHandler.Get)
			private.POST("/subscription/upgrade", c.SubHandler.Upgrade)
		}

		// Developer API Gateway Simulation Endpoint
		// Protected by API Key validation & Rate Limiter middleware
		gateway := v1.Group("/gateway")
		gateway.Use(middleware.APIKeyAuth(c.KeyRepo, c.CacheRepo))
		gateway.Use(middleware.RateLimit(c.Limiter, c.RuleRepo, c.CacheRepo, c.Producer, c.RedisClient, c.Hub))
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
