package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"limiter.io/internal/models"
)

type TenantHandler struct {
	db *gorm.DB
}

func NewTenantHandler(db *gorm.DB) *TenantHandler {
	return &TenantHandler{db: db}
}

func (h *TenantHandler) Create(c *gin.Context) {
	projectID := c.Param("projectId")
	var tc models.TenantConfig
	if err := c.ShouldBindJSON(&tc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tc.ProjectID = projectID
	if err := h.db.Create(&tc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, tc)
}

func (h *TenantHandler) List(c *gin.Context) {
	var configs []models.TenantConfig
	h.db.Where("project_id = ?", c.Param("projectId")).Find(&configs)
	c.JSON(http.StatusOK, configs)
}

func (h *TenantHandler) Get(c *gin.Context) {
	var tc models.TenantConfig
	if err := h.db.Where("id = ? AND project_id = ?", c.Param("tenantId"), c.Param("projectId")).First(&tc).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tenant not found"})
		return
	}
	c.JSON(http.StatusOK, tc)
}

func (h *TenantHandler) Update(c *gin.Context) {
	var tc models.TenantConfig
	if err := h.db.Where("id = ? AND project_id = ?", c.Param("tenantId"), c.Param("projectId")).First(&tc).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tenant not found"})
		return
	}
	var updates models.TenantConfig
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.db.Model(&tc).Updates(map[string]interface{}{
		"max_req":   updates.MaxReq,
		"window_ms": updates.WindowMs,
		"enabled":   updates.Enabled,
		"metadata":  updates.Metadata,
	})
	c.JSON(http.StatusOK, tc)
}

func (h *TenantHandler) Delete(c *gin.Context) {
	h.db.Where("id = ? AND project_id = ?", c.Param("tenantId"), c.Param("projectId")).Delete(&models.TenantConfig{})
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
