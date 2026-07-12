package handlers

import (
	"context"
	"net/http"
	"time"

	"limiter.io/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db  *gorm.DB
	rc  *redis.Client
	cfg *config.Config
}

func NewHealthHandler(db *gorm.DB, rc *redis.Client, cfg *config.Config) *HealthHandler {
	return &HealthHandler{
		db:  db,
		rc:  rc,
		cfg: cfg,
	}
}

func (h *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "UP"})
}

func (h *HealthHandler) Readiness(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	postgresOK := true
	redisOK := true
	kafkaOK := true

	// Check PostgreSQL
	sqlDB, err := h.db.DB()
	if err != nil {
		postgresOK = false
	} else if err := sqlDB.PingContext(ctx); err != nil {
		postgresOK = false
	}

	// Check Redis
	if err := h.rc.Ping(ctx).Err(); err != nil {
		redisOK = false
	}

	// Check Kafka
	conn, err := kafka.DialContext(ctx, "tcp", h.cfg.KafkaBrokers)
	if err != nil {
		kafkaOK = false
	} else {
		_ = conn.Close()
	}

	status := http.StatusOK
	if !postgresOK || !redisOK || !kafkaOK {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"postgres": map[string]interface{}{"status": getStatusString(postgresOK)},
		"redis":    map[string]interface{}{"status": getStatusString(redisOK)},
		"kafka":    map[string]interface{}{"status": getStatusString(kafkaOK)},
	})
}

func getStatusString(ok bool) string {
	if ok {
		return "UP"
	}
	return "DOWN"
}
