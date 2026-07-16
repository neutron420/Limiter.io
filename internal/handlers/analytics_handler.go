package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"limiter.io/internal/models"
	"limiter.io/internal/repository"
	"limiter.io/internal/services"
)

type AnalyticsHandler struct {
	analRepo repository.AnalyticsRepository
	dataSvc  *services.AnalyticsDataService
	db       *gorm.DB
}

func NewAnalyticsHandler(analRepo repository.AnalyticsRepository, db *gorm.DB) *AnalyticsHandler {
	return &AnalyticsHandler{
		analRepo: analRepo,
		dataSvc:  services.NewAnalyticsDataService(db),
		db:       db,
	}
}

func (h *AnalyticsHandler) GetStats(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}
	logs, err := h.analRepo.GetLogs(c.Request.Context(), projectID, 100, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"total": len(logs)})
}

func (h *AnalyticsHandler) GetLogs(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	logs, err := h.analRepo.GetLogs(c.Request.Context(), projectID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}

func (h *AnalyticsHandler) ExportLogs(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}
	format := c.DefaultQuery("format", "json")
	logs, _ := h.analRepo.GetLogs(c.Request.Context(), projectID, 10000, 0)
	if format == "csv" {
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=analytics.csv")
		c.String(200, "id,project_id,api_key_id,route,status_code,latency_ms,decision,timestamp\n")
		for _, l := range logs {
			c.String(200, "%s,%s,%s,%s,%d,%d,%s,%s\n", l.ID, l.ProjectID, l.APIKeyID, l.Route, l.StatusCode, l.LatencyMs, l.Decision, l.Timestamp)
		}
		return
	}
	c.JSON(http.StatusOK, logs)
}

func (h *AnalyticsHandler) GetTimeSeries(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}
	bucket := c.DefaultQuery("bucket", "hour")
	now := time.Now()
	data, err := h.analRepo.GetTimeSeries(c.Request.Context(), projectID, now.Add(-24*time.Hour), now, bucket)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *AnalyticsHandler) GetAnalyticsData(c *gin.Context) {
	data, _ := h.dataSvc.GetAnalytics(c.Param("projectId"))
	c.JSON(http.StatusOK, data)
}

func (h *AnalyticsHandler) SaveView(c *gin.Context) {
	var view models.SavedAnalyticsView
	if err := c.ShouldBindJSON(&view); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	view.ProjectID = c.Param("projectId")
	view.UserID = c.GetString("user_id")
	if err := h.dataSvc.SaveAnalyticsView(&view); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, view)
}

func (h *AnalyticsHandler) ListViews(c *gin.Context) {
	views, _ := h.dataSvc.ListAnalyticsViews(c.Param("projectId"), c.GetString("user_id"))
	c.JSON(http.StatusOK, views)
}

func (h *AnalyticsHandler) GetView(c *gin.Context) {
	view, err := h.dataSvc.GetAnalyticsView(c.Param("viewId"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "view not found"})
		return
	}
	c.JSON(http.StatusOK, view)
}

func (h *AnalyticsHandler) DeleteView(c *gin.Context) {
	if err := h.dataSvc.DeleteAnalyticsView(c.Param("viewId"), c.GetString("user_id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "view not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "view deleted"})
}

func (h *AnalyticsHandler) GetAnomalyConfig(c *gin.Context) {
	cfg, _ := h.dataSvc.GetAnomalyConfigItem(c.Param("projectId"))
	c.JSON(http.StatusOK, cfg)
}

func (h *AnalyticsHandler) UpdateAnomalyConfig(c *gin.Context) {
	var cfg models.AnomalyDetectionConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cfg.ProjectID = c.Param("projectId")
	if err := h.dataSvc.UpdateAnomalyConfigItem(&cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

func (h *AnalyticsHandler) DetectAnomalies(c *gin.Context) {
	alerts, _ := h.dataSvc.DetectAnomalies(c.Param("projectId"))
	c.JSON(http.StatusOK, gin.H{"alerts": alerts})
}
