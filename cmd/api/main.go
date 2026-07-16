package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"limiter.io/internal/config"
	"limiter.io/internal/database"
	deliveryhttp "limiter.io/internal/delivery/http"
	"limiter.io/internal/handlers"
	"limiter.io/internal/kafka"
	"limiter.io/internal/mailer"
	"limiter.io/internal/middleware"
	"limiter.io/internal/ratelimiter"
	internalredis "limiter.io/internal/redis"
	"limiter.io/internal/repository/postgres"
	repo_redis "limiter.io/internal/repository/redis"
	"limiter.io/internal/services"
	internalws "limiter.io/internal/websocket"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// @title API Rate Limiting Platform
// @version 1.0
// @description High-performance distributed API rate limiting platform with Redis, PostgreSQL, Kafka, and WebSockets.
// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer <your-token>" to authenticate.

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description API key used for rate-limited gateway requests.
func main() {
	// 1. Load config
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Setup structured logging
	logger := initLogger(cfg.Env)
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	logger.Info("Starting API Rate Limiter Platform...", zap.String("env", cfg.Env))

	// 3. Connect to Database (Postgres)
	db, err := database.ConnectDB(cfg)
	if err != nil {
		logger.Fatal("Database connection failed", zap.Error(err))
	}
	logger.Info("Connected to PostgreSQL database successfully")

	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal("Failed to access SQL connection pool", zap.Error(err))
	}
	defer sqlDB.Close()

	// 4. Run Migrations and Seeding
	if err := database.MigrateAndSeed(db, cfg); err != nil {
		logger.Fatal("Database migration & seeding failed", zap.Error(err))
	}
	logger.Info("Database migration & seeding completed")

	// 5. Connect to Redis
	rc, err := internalredis.ConnectRedis(cfg)
	if err != nil {
		logger.Fatal("Redis connection failed", zap.Error(err))
	}
	logger.Info("Connected to Redis and preloaded Lua scripts successfully")

	// 6. Connect to Kafka Producer
	producer, err := kafka.NewKafkaProducer(cfg)
	if err != nil {
		logger.Fatal("Kafka producer initialization failed", zap.Error(err))
	}
	logger.Info("Connected to Kafka Producer successfully")

	// 7. Initialize Repositories
	userRepo := postgres.NewUserRepository(db)
	rtRepo := postgres.NewRefreshTokenRepository(db)
	projRepo := postgres.NewProjectRepository(db)
	keyRepo := postgres.NewAPIKeyRepository(db)
	ruleRepo := postgres.NewRateLimitRuleRepository(db)
	subRepo := postgres.NewSubscriptionRepository(db)
	analRepo := postgres.NewAnalyticsRepository(db)
	cacheRepo := repo_redis.NewCacheRepository(rc)
	prtRepo := postgres.NewPasswordResetTokenRepository(db)
	webhookRepo := postgres.NewWebhookEventRepository(db)
	memberRepo := postgres.NewProjectMemberRepository(db)
	inviteRepo := postgres.NewProjectInviteRepository(db)
	auditRepo := postgres.NewProjectAuditRepository(db)

	// Transactional email (Resend, or logged in dev when no key)
	mail := mailer.New(cfg.ResendAPIKey, cfg.ResendFrom)

	// 8. Initialize Services
	authService := services.NewAuthService(userRepo, rtRepo, subRepo, cacheRepo, prtRepo, mail, cfg)
	projService := services.NewProjectService(projRepo, subRepo, memberRepo, userRepo, inviteRepo, auditRepo, mail, cfg)
	keyService := services.NewAPIKeyService(keyRepo, projRepo, subRepo, cacheRepo, memberRepo, auditRepo)
	policyService := services.NewPolicyService(ruleRepo, projRepo, subRepo, memberRepo, auditRepo)
	subService := services.NewSubscriptionService(subRepo, cacheRepo, projRepo, analRepo)
	_ = analRepo // used by analHandler
	_ = projRepo
	_ = memberRepo

	// Rate limiting engine
	redisLimiter := ratelimiter.NewRedisRateLimiter(rc)

	// WebSocket Hub
	hub := internalws.NewHub()
	go hub.Run()

	// Initialize alert service and start the periodic evaluator
	alertService := services.NewAlertService(
		postgres.NewAlertRepository(db),
		postgres.NewAnalyticsRepository(db),
		projRepo, memberRepo, mail,
	)
	alertCtx, alertCancel := context.WithCancel(context.Background())
	alertService.StartEvaluator(alertCtx, 1*time.Minute)

	// IP access service
	ipAccessRepo := postgres.NewIPAccessRepository(db)
	ipAccessService := services.NewIPAccessService(ipAccessRepo, projRepo, memberRepo)

	// Security service (MFA, sessions)
	securityService := services.NewSecurityService(userRepo, rtRepo)

	// Organization, Notification, Approval services
	orgService := services.NewOrganizationService(db)
	notifService := services.NewNotificationService(db)
	approvalService := services.NewApprovalService(db)
	quotaService := services.NewQuotaService(db)
	analyticsDataSvc := services.NewAnalyticsDataService(db)
	passkeyService := services.NewPasskeyService(db)
	immutableAuditService := services.NewImmutableAuditService(db)
	billingSvc := services.NewBillingService(db)
	sandboxService := services.NewSandboxService(db)

	// Use services (passed to handlers)
	_ = orgService
	_ = notifService
	_ = approvalService
	_ = quotaService
	_ = analyticsDataSvc
	_ = passkeyService
	_ = immutableAuditService
	_ = billingSvc
	_ = sandboxService

	// Key rotation reminder service
	keyRotationService := services.NewKeyRotationService(db)
	keyRotationService.StartRotationChecker(24 * time.Hour)

	// Initialize maintenance mode
	middleware.InitMaintenance(db)

	// 9. Initialize Handlers
	authHandler := handlers.NewAuthHandler(authService)
	projHandler := handlers.NewProjectHandler(projService)
	keyHandler := handlers.NewAPIKeyHandler(keyService)
	policyHandler := handlers.NewPolicyHandler(policyService)
	subHandler := handlers.NewSubscriptionHandler(subService)
	analHandler := handlers.NewAnalyticsHandler(analRepo, db)
	healthHandler := handlers.NewHealthHandler(db, rc.Client, cfg)
	wsHandler := handlers.NewWSHandler(hub, projRepo, memberRepo)
	billingHandler := handlers.NewBillingHandler(cfg, userRepo, subService, webhookRepo)
	billingHandler.SetDB(db) // enable billing DB features
	securityHandler := handlers.NewSecurityHandler(securityService)
	ipAccessHandler := handlers.NewIPAccessHandler(ipAccessService)
	notifHandler := handlers.NewNotificationHandler(db)
	orgHandler := handlers.NewOrganizationHandler(db)
	approvalH := handlers.NewApprovalHandler(db)
	quotaH := handlers.NewQuotaHandler(db)
	tenantHandler := handlers.NewTenantHandler(db)
	// analyticsFullHandler not needed separately; analHandler covers all analytics routes
	passkeyHandler := handlers.NewPasskeyHandler(db)
	immutableAuditHandler := handlers.NewImmutableAuditHandler(db)
	statusHandler := handlers.NewStatusHandler(db, "1.0.0")
	sandboxHandler := handlers.NewSandboxHandler(db)
	maintenanceHandler := handlers.NewMaintenanceHandler()
	ssoHandler := handlers.NewSSOHandler()

	// Start background analytics retention cleanup
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				deleted, err := analRepo.PurgeExpiredByPlan(context.Background())
				if err != nil {
					logger.Warn("analytics retention cleanup failed", zap.Error(err))
				} else if deleted > 0 {
					logger.Info("analytics retention cleanup completed", zap.Int64("deleted", deleted))
				}
			case <-alertCtx.Done():
				return
			}
		}
	}()

	// Start background invite expiration cleanup
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				cleaned, err := projService.CleanupExpiredInvites(context.Background())
				if err != nil {
					logger.Warn("invite expiration cleanup failed", zap.Error(err))
				} else if cleaned > 0 {
					logger.Info("expired invites cleaned up", zap.Int64("count", cleaned))
				}
			case <-alertCtx.Done():
				return
			}
		}
	}()

	// 10. Start Gin Engine
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()

	deliveryhttp.SetupRouter(deliveryhttp.RouterConfig{
		Engine:                r,
		Cfg:                   cfg,
		AuthHandler:           authHandler,
		ProjHandler:           projHandler,
		KeyHandler:            keyHandler,
		PolicyHandler:         policyHandler,
		SubHandler:            subHandler,
		AnalHandler:           analHandler,
		HealthHandler:         healthHandler,
		WSHandler:             wsHandler,
		BillingHandler:        billingHandler,
		SecurityHandler:       securityHandler,
		IPAccessHandler:       ipAccessHandler,
		NotifHandler:          notifHandler,
		OrgHandler:            orgHandler,
		ApprovalHandler:       approvalH,
		QuotaHandler:          quotaH,
		TenantHandler:         tenantHandler,
		PasskeyHandler:        passkeyHandler,
		ImmutableAuditHandler: immutableAuditHandler,
		StatusHandler:         statusHandler,
		SandboxHandler:        sandboxHandler,
		MaintenanceHandler:    maintenanceHandler,
		SSOHandler:            ssoHandler,
		Hub:                   hub,
		Limiter:               redisLimiter,
		RuleRepo:              ruleRepo,
		KeyRepo:               keyRepo,
		CacheRepo:             cacheRepo,
		AnalRepo:              analRepo,
		Producer:              producer,
		RedisClient:           rc.Client,
	})

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// 11. Run Server inside a goroutine to handle graceful shutdown
	go func() {
		logger.Info("HTTP Server listening", zap.String("port", cfg.Port))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("HTTP server failed to start", zap.Error(err))
		}
	}()

	// Graceful shutdown signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down HTTP server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("HTTP Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Stopping background workers...")
	alertCancel()

	logger.Info("Closing Kafka producer connection...")
	if err := producer.Close(); err != nil {
		logger.Error("Error closing Kafka producer", zap.Error(err))
	}

	logger.Info("Closing Redis connection...")
	if err := rc.Client.Close(); err != nil {
		logger.Error("Error closing Redis client", zap.Error(err))
	}

	logger.Info("Server exited gracefully")
}

func initLogger(env string) *zap.Logger {
	var config zap.Config

	if env == "production" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	log, _ := config.Build()
	return log
}
