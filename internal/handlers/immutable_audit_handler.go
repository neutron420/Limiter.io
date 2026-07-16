package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"limiter.io/internal/services"
)

type ImmutableAuditHandler struct {
	svc *services.ImmutableAuditService
}

func NewImmutableAuditHandler(db *gorm.DB) *ImmutableAuditHandler {
	return &ImmutableAuditHandler{
		svc: services.NewImmutableAuditService(db),
	}
}

func (h *ImmutableAuditHandler) List(c *gin.Context) {
	projectID := c.Query("project_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit > 200 {
		limit = 200
	}
	logs, total, err := h.svc.List(projectID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"logs": logs, "total": total, "limit": limit, "offset": offset})
}

func (h *ImmutableAuditHandler) GetByID(c *gin.Context) {
	log, err := h.svc.GetByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "log not found"})
		return
	}
	c.JSON(http.StatusOK, log)
}

func (h *ImmutableAuditHandler) VerifyChain(c *gin.Context) {
	valid, err := h.svc.VerifyChain()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"valid": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"valid": valid})
}
