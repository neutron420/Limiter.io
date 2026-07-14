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

	// 8. Initialize Services
	authService := services.NewAuthService(userRepo, rtRepo, subRepo, cacheRepo, prtRepo, cfg)
	projService := services.NewProjectService(projRepo, subRepo, memberRepo, userRepo)
	keyService := services.NewAPIKeyService(keyRepo, projRepo, subRepo, cacheRepo, memberRepo)
	policyService := services.NewPolicyService(ruleRepo, projRepo, subRepo, memberRepo)
	subService := services.NewSubscriptionService(subRepo, cacheRepo, projRepo, analRepo)
	analService := services.NewAnalyticsService(analRepo, projRepo, memberRepo)

	// Rate limiting engine
	redisLimiter := ratelimiter.NewRedisRateLimiter(rc)

	// WebSocket Hub
	hub := internalws.NewHub()
	go hub.Run()

	// 9. Initialize Handlers
	authHandler := handlers.NewAuthHandler(authService)
	projHandler := handlers.NewProjectHandler(projService)
	keyHandler := handlers.NewAPIKeyHandler(keyService)
	policyHandler := handlers.NewPolicyHandler(policyService)
	subHandler := handlers.NewSubscriptionHandler(subService)
	analHandler := handlers.NewAnalyticsHandler(analService)
	healthHandler := handlers.NewHealthHandler(db, rc.Client, cfg)
	wsHandler := handlers.NewWSHandler(hub, projRepo)
	billingHandler := handlers.NewBillingHandler(cfg, userRepo, subService, webhookRepo)

	// 10. Start Gin Engine
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()

	deliveryhttp.SetupRouter(deliveryhttp.RouterConfig{
		Engine:         r,
		Cfg:            cfg,
		AuthHandler:    authHandler,
		ProjHandler:    projHandler,
		KeyHandler:     keyHandler,
		PolicyHandler:  policyHandler,
		SubHandler:     subHandler,
		AnalHandler:    analHandler,
		HealthHandler:  healthHandler,
		WSHandler:      wsHandler,
		BillingHandler: billingHandler,
		Hub:            hub,
		Limiter:        redisLimiter,
		RuleRepo:       ruleRepo,
		KeyRepo:        keyRepo,
		CacheRepo:      cacheRepo,
		AnalRepo:       analRepo,
		Producer:       producer,
		RedisClient:    rc.Client,
	})

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
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
