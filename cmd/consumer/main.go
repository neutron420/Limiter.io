package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"limiter.io/internal/config"
	"limiter.io/internal/database"
	"limiter.io/internal/kafka"

	"go.uber.org/zap"
)

func main() {
	// 1. Load config
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Setup structured logging
	var logConfig zap.Config
	if cfg.Env == "production" {
		logConfig = zap.NewProductionConfig()
	} else {
		logConfig = zap.NewDevelopmentConfig()
	}
	logger, _ := logConfig.Build()
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	logger.Info("Starting Kafka consumer worker...", zap.String("env", cfg.Env))

	// 3. Connect to Postgres
	db, err := database.ConnectDB(cfg)
	if err != nil {
		logger.Fatal("Database connection failed", zap.Error(err))
	}
	logger.Info("Connected to PostgreSQL successfully")

	// 4. Initialize Consumer
	consumer := kafka.NewKafkaConsumer(cfg, db)
	logger.Info("Initialized Kafka Consumer successfully")

	// 5. Start consuming with context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		if err := consumer.Start(ctx); err != nil {
			errChan <- err
		}
	}()

	logger.Info("Consumer group worker is now reading messages...")

	// 6. Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-quit:
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
	case err := <-errChan:
		if err != context.Canceled {
			logger.Error("Consumer run loop stopped with error", zap.Error(err))
		}
	}

	logger.Info("Cancelling consumer context...")
	cancel() // notifies worker to flush and terminate

	logger.Info("Closing Kafka consumer reader...")
	if err := consumer.Close(); err != nil {
		logger.Error("Error closing Kafka reader", zap.Error(err))
	}

	// Wait briefly for DB writes to flush
	time.Sleep(1 * time.Second)
	logger.Info("Consumer worker stopped successfully")
}
