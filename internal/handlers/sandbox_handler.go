package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"limiter.io/internal/services"
)

type SandboxHandler struct {
	svc *services.SandboxService
}

func NewSandboxHandler(db *gorm.DB) *SandboxHandler {
	return &SandboxHandler{
		svc: services.NewSandboxService(db),
	}
}

func (h *SandboxHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")
	project, err := h.svc.CreateSandboxProject(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"project":     project,
		"message":     "Sandbox project created with a test API key and rate limit rule",
	})
}

func (h *SandboxHandler) Cleanup(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}
	if err := h.svc.CleanupSandboxProject(projectID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Sandbox project cleaned up"})
}
