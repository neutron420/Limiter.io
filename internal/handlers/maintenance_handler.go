package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"limiter.io/internal/middleware"
)

type MaintenanceHandler struct{}

func NewMaintenanceHandler() *MaintenanceHandler {
	return &MaintenanceHandler{}
}

func (h *MaintenanceHandler) GetStatus(c *gin.Context) {
	if middleware.GlobalMaintenance == nil {
		c.JSON(http.StatusOK, gin.H{"enabled": false})
		return
	}
	enabled, message := middleware.GetMaintenanceStatus()
	c.JSON(http.StatusOK, gin.H{
		"enabled": enabled,
		"message": message,
	})
}

func (h *MaintenanceHandler) SetMaintenance(c *gin.Context) {
	var req struct {
		Enabled bool   `json:"enabled"`
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	middleware.SetMaintenance(req.Enabled, req.Message)
	c.JSON(http.StatusOK, gin.H{"message": "maintenance mode updated"})
}

func (h *MaintenanceHandler) EmergencyBlock(c *gin.Context) {
	var req struct {
		Identifier string `json:"identifier"`
		Duration   string `json:"duration"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dur, err := time.ParseDuration(req.Duration)
	if err != nil {
		dur = 5 * time.Minute
	}
	middleware.EmergencyBlock(req.Identifier, dur)
	c.JSON(http.StatusOK, gin.H{"message": "emergency block applied"})
}

func (h *MaintenanceHandler) CheckBlock(c *gin.Context) {
	identifier := c.Query("identifier")
	if identifier == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "identifier required"})
		return
	}
	blocked := middleware.IsEmergencyBlocked(identifier)
	c.JSON(http.StatusOK, gin.H{"blocked": blocked})
}
