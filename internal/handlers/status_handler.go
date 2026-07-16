package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StatusHandler struct {
	db          *gorm.DB
	startTime   time.Time
	version     string
}

func NewStatusHandler(db *gorm.DB, version string) *StatusHandler {
	return &StatusHandler{
		db:        db,
		startTime: time.Now(),
		version:   version,
	}
}

func (h *StatusHandler) Status(c *gin.Context) {
	dbOK := true
	sqlDB, err := h.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		dbOK = false
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	c.JSON(http.StatusOK, gin.H{
		"status":        "ok",
		"version":       h.version,
		"uptime":        time.Since(h.startTime).Seconds(),
		"database":      dbOK,
		"go_version":    runtime.Version(),
		"goroutines":    runtime.NumGoroutine(),
		"memory_used_mb": m.Alloc / 1024 / 1024,
		"timestamp":     time.Now().UTC(),
	})
}

func (h *StatusHandler) Health(c *gin.Context) {
	dbOK := true
	sqlDB, err := h.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		dbOK = false
	}
	if !dbOK {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
