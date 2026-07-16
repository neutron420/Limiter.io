package database

import (
	"fmt"
	"log"
	"time"

	"limiter.io/internal/config"
	"limiter.io/internal/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSslMode,
	)

	var db *gorm.DB
	var err error

	gormLogger := logger.Default.LogMode(logger.Warn)
	if cfg.Env == "development" {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	// Retry database connection setup
	for i := 1; i <= 5; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: gormLogger,
		})
		if err == nil {
			break
		}
		log.Printf("Failed to connect to database (attempt %d/5): %v. Retrying in 3 seconds...", i, err)
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("could not connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func MigrateAndSeed(db *gorm.DB, cfg *config.Config) error {
	// Auto migrate tables
	err := db.AutoMigrate(
		&models.User{},
		&models.RefreshToken{},
		&models.Project{},
		&models.APIKey{},
		&models.Plan{},
		&models.Subscription{},
		&models.UpgradeHistory{},
		&models.RateLimitRule{},
		&models.AnalyticsLog{},
		&models.PasswordResetToken{},
		&models.WebhookEvent{},
		&models.ProjectMember{},
		&models.ProjectInvite{},
		&models.ProjectAuditEvent{},
		&models.AlertRule{},
		&models.AlertEvent{},
		&models.IPAccessRule{},
		&models.RuleVersion{},
		&models.NotificationPreferences{},
		&models.Organization{},
		&models.OrganizationMember{},
		&models.OrganizationGroup{},
		&models.OrganizationGroupMember{},
		&models.ApprovalWorkflow{},
		&models.ApprovalRequest{},
		&models.Quota{},
		&models.TenantConfig{},
		&models.SavedAnalyticsView{},
		&models.AnomalyDetectionConfig{},
		&models.Passkey{},
		&models.ImmutableAuditLog{},
		&models.UsageRecord{},
		&models.Invoice{},
		&models.SLAConfig{},
		&models.EmailTemplate{},
		&models.RegionConfig{},
	)
	if err != nil {
		return fmt.Errorf("failed to run database migration: %w", err)
	}

	// Seed plans
	plans := []models.Plan{
		{
			ID:                     "free",
			Name:                   "Free Plan",
			MaxProjects:            3,
			MaxKeysPerProject:      3,
			AllowedAlgorithms:      "token_bucket",
			AnalyticsRetentionDays: 7,
			RateLimitRequests:      100,
			RateLimitPeriod:        60, // 100 req per min
		},
		{
			ID:                     "pro",
			Name:                   "Pro Plan",
			MaxProjects:            -1, // unlimited
			MaxKeysPerProject:      -1, // unlimited
			AllowedAlgorithms:      "token_bucket,fixed_window,sliding_window_counter,sliding_window_log,leaky_bucket",
			AnalyticsRetentionDays: 90,
			RateLimitRequests:      10000,
			RateLimitPeriod:        60, // 10k req per min
		},
		{
			ID:                     "enterprise",
			Name:                   "Enterprise Plan",
			MaxProjects:            -1,
			MaxKeysPerProject:      -1,
			AllowedAlgorithms:      "token_bucket,fixed_window,sliding_window_counter,sliding_window_log,leaky_bucket",
			AnalyticsRetentionDays: 365,
			RateLimitRequests:      1000000,
			RateLimitPeriod:        60,
		},
	}

	for _, p := range plans {
		if err := db.Save(&p).Error; err != nil {
			return fmt.Errorf("failed to seed plan %s: %w", p.ID, err)
		}
	}

	// Seed Default Platform Admin if not present
	var adminCount int64
	db.Model(&models.User{}).Where("email = ?", cfg.AdminEmail).Count(&adminCount)
	if adminCount == 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash default admin password: %w", err)
		}

		admin := models.User{
			ID:           uuid.New(),
			Email:        cfg.AdminEmail,
			PasswordHash: string(hashedPassword),
		}

		if err := db.Create(&admin).Error; err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}

		// Seed active subscription for the admin
		now := time.Now()
		sub := models.Subscription{
			ID:              uuid.New(),
			UserID:          admin.ID,
			PlanID:          "pro", // Admin gets the Pro Plan by default
			Status:          "active",
			StartsAt:        now,
			BillingMetadata: models.JSONMap{"source": "system_seeding"},
		}

		if err := db.Create(&sub).Error; err != nil {
			return fmt.Errorf("failed to create admin subscription: %w", err)
		}

		log.Printf("Seeded default admin user: %s", cfg.AdminEmail)
	}

	return nil
}
