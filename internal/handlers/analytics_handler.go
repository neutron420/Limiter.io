package handlers

import (
	"net/http"
	"strconv"
	"time"

	"limiter.io/internal/dto"
	"limiter.io/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AnalyticsHandler struct {
	analyticsService services.AnalyticsService
}

func NewAnalyticsHandler(analyticsService services.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsService: analyticsService}
}

func (h *AnalyticsHandler) GetStats(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid project ID format"})
		return
	}

	// Parse duration query parameter (e.g. 24h, 7d, 90d, all)
	durationStr := c.DefaultQuery("duration", "24h")
	var duration time.Duration
	if durationStr == "all" {
		duration = 10 * 365 * 24 * time.Hour
	} else {
		// Convert 'd' suffix to hours since time.ParseDuration doesn't support 'd'
		if len(durationStr) > 1 && durationStr[len(durationStr)-1] == 'd' {
			daysStr := durationStr[:len(durationStr)-1]
			days, err := strconv.Atoi(daysStr)
			if err == nil {
				durationStr = strconv.Itoa(days*24) + "h"
			}
		}
		var err error
		duration, err = time.ParseDuration(durationStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid duration format. Examples: 24h, 7d, 30d, all"})
			return
		}
	}

	userID := uuid.MustParse(userIDStr.(string))
	stats, err := h.analyticsService.GetProjectStats(c.Request.Context(), userID, projectID, duration)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *AnalyticsHandler) GetLogs(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid project ID format"})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	userID := uuid.MustParse(userIDStr.(string))
	logs, err := h.analyticsService.GetProjectLogs(c.Request.Context(), userID, projectID, limit, offset)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, logs)
}

func (h *AnalyticsHandler) GetTimeSeries(c *gin.Context) {
	userIDStr, exists := c.Get("UserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	projectIDStr := c.Param("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid project ID format"})
		return
	}

	durationStr := c.DefaultQuery("duration", "24h")
	var duration time.Duration
	if durationStr == "all" {
		duration = 10 * 365 * 24 * time.Hour
	} else {
		if len(durationStr) > 1 && durationStr[len(durationStr)-1] == 'd' {
			daysStr := durationStr[:len(durationStr)-1]
			days, err := strconv.Atoi(daysStr)
			if err == nil {
				durationStr = strconv.Itoa(days*24) + "h"
			}
		}
		var err error
		duration, err = time.ParseDuration(durationStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid duration format. Examples: 24h, 7d, 30d, all"})
			return
		}
	}

	bucket := c.DefaultQuery("bucket", "hour")

	userID := uuid.MustParse(userIDStr.(string))
	series, err := h.analyticsService.GetTimeSeries(c.Request.Context(), userID, projectID, duration, bucket)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, series)
}
