package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"limiter.io/internal/models"
	"limiter.io/internal/services"
)

type NotificationHandler struct {
	svc *services.NotificationService
	db  *gorm.DB
}

func NewNotificationHandler(db *gorm.DB) *NotificationHandler {
	return &NotificationHandler{
		svc: services.NewNotificationService(db),
		db:  db,
	}
}

func (h *NotificationHandler) GetPreferences(c *gin.Context) {
	userID := c.GetString("user_id")
	projectID := c.Param("id")
	prefs, err := h.svc.GetPreferences(userID, projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, prefs)
}

func (h *NotificationHandler) UpdatePreferences(c *gin.Context) {
	userID := c.GetString("user_id")
	projectID := c.Param("id")
	var prefs models.NotificationPreferences
	if err := c.ShouldBindJSON(&prefs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	prefs.UserID = userID
	prefs.ProjectID = projectID
	if err := h.svc.UpdatePreferences(&prefs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, prefs)
}
