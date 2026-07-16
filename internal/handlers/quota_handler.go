package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"limiter.io/internal/models"
	"limiter.io/internal/services"
)

type QuotaHandler struct {
	svc *services.QuotaService
}

func NewQuotaHandler(db *gorm.DB) *QuotaHandler {
	return &QuotaHandler{
		svc: services.NewQuotaService(db),
	}
}

func (h *QuotaHandler) GetQuota(c *gin.Context) {
	projectID := c.Param("id")
	quota, err := h.svc.GetQuota(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "quota not found"})
		return
	}
	c.JSON(http.StatusOK, quota)
}

func (h *QuotaHandler) SetQuota(c *gin.Context) {
	var quota models.Quota
	if err := c.ShouldBindJSON(&quota); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	quota.ProjectID = c.Param("id")
	if err := h.svc.SetQuota(&quota); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, quota)
}

func (h *QuotaHandler) CheckQuota(c *gin.Context) {
	projectID := c.Param("id")
	allowed, err := h.svc.CheckQuota(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"allowed": allowed})
}
