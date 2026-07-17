package handlers

import (
	"net/http"
	"sort"
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
	logs, err := h.analRepo.GetLogs(c.Request.Context(), projectID, 10000, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	allowed, blocked, totalLatency := 0, 0, 0
	for _, l := range logs {
		if l.Decision == "allowed" {
			allowed++
		} else {
			blocked++
		}
		totalLatency += l.LatencyMs
	}
	avgLatency := 0.0
	if len(logs) > 0 {
		avgLatency = float64(totalLatency) / float64(len(logs))
	}
	// Compute top blocked routes
	type routeCount struct {
		Route string `json:"route"`
		Count int    `json:"count"`
	}
	blockedMap := make(map[string]int)
	for _, l := range logs {
		if l.Decision == "blocked" {
			blockedMap[l.Route]++
		}
	}
	blockedList := make([]routeCount, 0, len(blockedMap))
	for route, count := range blockedMap {
		blockedList = append(blockedList, routeCount{Route: route, Count: count})
	}
	sort.Slice(blockedList, func(i, j int) bool {
		return blockedList[i].Count > blockedList[j].Count
	})
	if len(blockedList) > 10 {
		blockedList = blockedList[:10]
	}

	c.JSON(http.StatusOK, gin.H{
		"total_requests":   len(logs),
		"allowed_requests": allowed,
		"blocked_requests": blocked,
		"avg_latency_ms":   avgLatency,
		"top_blocked":      blockedList,
	})
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
	duration := c.DefaultQuery("duration", "24h")
	ago := time.Now()
	switch duration {
	case "7d":
		ago = ago.Add(-7 * 24 * time.Hour)
	case "30d":
		ago = ago.Add(-30 * 24 * time.Hour)
	case "all":
		ago = time.Time{}
	default:
		ago = ago.Add(-24 * time.Hour)
	}
	rows, err := h.analRepo.GetTimeSeries(c.Request.Context(), projectID, ago, time.Now(), bucket)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
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
