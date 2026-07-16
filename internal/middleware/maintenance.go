package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MaintenanceMode struct {
	Mu      sync.RWMutex
	Enabled bool
	Message string
	DB      *gorm.DB
}

func (m *MaintenanceMode) RLock() {
	m.Mu.RLock()
}

func (m *MaintenanceMode) RUnlock() {
	m.Mu.RUnlock()
}

func GetMaintenanceStatus() (bool, string) {
	if GlobalMaintenance == nil {
		return false, ""
	}
	GlobalMaintenance.Mu.RLock()
	defer GlobalMaintenance.Mu.RUnlock()
	return GlobalMaintenance.Enabled, GlobalMaintenance.Message
}

var GlobalMaintenance *MaintenanceMode

func InitMaintenance(db *gorm.DB) {
	GlobalMaintenance = &MaintenanceMode{DB: db}
	var config struct {
		Enabled bool
		Message string
	}
	db.Raw("SELECT enabled, message FROM maintenance_config ORDER BY id DESC LIMIT 1").Scan(&config)
	GlobalMaintenance.Enabled = config.Enabled
	GlobalMaintenance.Message = config.Message
	if GlobalMaintenance.Message == "" {
		GlobalMaintenance.Message = "Service is under maintenance. Please try again later."
	}
}

func MaintenanceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if GlobalMaintenance == nil {
			c.Next()
			return
		}
		GlobalMaintenance.Mu.RLock()
		enabled := GlobalMaintenance.Enabled
		message := GlobalMaintenance.Message
		GlobalMaintenance.Mu.RUnlock()

		if enabled {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error":   "maintenance_mode",
				"message": message,
				"retry_after": time.Now().Add(5 * time.Minute).Unix(),
			})
			return
		}
		c.Next()
	}
}

func SetMaintenance(enabled bool, message string) {
	if GlobalMaintenance == nil {
		return
	}
	GlobalMaintenance.Mu.Lock()
	defer GlobalMaintenance.Mu.Unlock()
	GlobalMaintenance.Enabled = enabled
	if message != "" {
		GlobalMaintenance.Message = message
	}
	if GlobalMaintenance.DB != nil {
		GlobalMaintenance.DB.Exec("INSERT INTO maintenance_config (enabled, message, created_at) VALUES (?, ?, NOW())", enabled, GlobalMaintenance.Message)
	}
}

func EmergencyBlock(key string, duration time.Duration) {
	if GlobalMaintenance == nil || GlobalMaintenance.DB == nil {
		return
	}
	expiresAt := time.Now().Add(duration)
	GlobalMaintenance.DB.Exec(
		"INSERT INTO emergency_blocks (identifier, blocked_until, created_at) VALUES (?, ?, NOW()) ON CONFLICT (identifier) DO UPDATE SET blocked_until = ?",
		key, expiresAt, expiresAt,
	)
}

func IsEmergencyBlocked(identifier string) bool {
	if GlobalMaintenance == nil || GlobalMaintenance.DB == nil {
		return false
	}
	var count int64
	GlobalMaintenance.DB.Raw(
		"SELECT COUNT(*) FROM emergency_blocks WHERE identifier = ? AND blocked_until > NOW()",
		identifier,
	).Scan(&count)
	return count > 0
}
