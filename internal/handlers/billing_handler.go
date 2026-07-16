package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"limiter.io/internal/config"
	"limiter.io/internal/models"
	"limiter.io/internal/repository"
	"limiter.io/internal/services"
)

type BillingHandler struct {
	cfg       *config.Config
	userRepo  repository.UserRepository
	subSvc    services.SubscriptionService
	webhookRepo repository.WebhookEventRepository
	svc       *services.BillingService
}

func NewBillingHandler(cfg *config.Config, userRepo repository.UserRepository, subSvc services.SubscriptionService, webhookRepo repository.WebhookEventRepository) *BillingHandler {
	return &BillingHandler{
		cfg: cfg, userRepo: userRepo, subSvc: subSvc, webhookRepo: webhookRepo,
	}
}

func (h *BillingHandler) SetDB(db *gorm.DB) {
	h.svc = services.NewBillingService(db)
}

func (h *BillingHandler) LemonSqueezyWebhook(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "webhook received"})
}

func (h *BillingHandler) ListWebhooks(c *gin.Context) {
	c.JSON(http.StatusOK, []models.WebhookEvent{})
}

func (h *BillingHandler) GetUsage(c *gin.Context) {
	if h.svc == nil {
		c.JSON(http.StatusNotFound, gin.H{"usage": gin.H{"request_count": 0, "blocked_count": 0}})
		return
	}
	record, err := h.svc.GetUsage(c.Param("id"), 0, 0)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"usage": gin.H{"request_count": 0, "blocked_count": 0}})
		return
	}
	c.JSON(http.StatusOK, record)
}

func (h *BillingHandler) ListInvoices(c *gin.Context) {
	if h.svc == nil {
		c.JSON(http.StatusOK, []models.Invoice{})
		return
	}
	invoices, _ := h.svc.ListInvoices(c.Param("id"))
	c.JSON(http.StatusOK, invoices)
}

func (h *BillingHandler) GetInvoice(c *gin.Context) {
	if h.svc == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
		return
	}
	invoice, err := h.svc.GetInvoice(c.Param("invoiceId"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
		return
	}
	c.JSON(http.StatusOK, invoice)
}

func (h *BillingHandler) GetSLAConfig(c *gin.Context) {
	if h.svc == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "SLA config not found"})
		return
	}
	cfg, _ := h.svc.GetSLAConfig(c.Param("orgId"))
	c.JSON(http.StatusOK, cfg)
}

func (h *BillingHandler) UpdateSLAConfig(c *gin.Context) {
	if h.svc == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "SLA config not found"})
		return
	}
	var cfg models.SLAConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cfg.OrganizationID = c.Param("orgId")
	h.svc.UpdateSLAConfig(&cfg)
	c.JSON(http.StatusOK, cfg)
}

func (h *BillingHandler) GetEmailTemplate(c *gin.Context) {
	if h.svc == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}
	tmpl, err := h.svc.GetEmailTemplate(c.Param("orgId"), c.Query("name"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}
	c.JSON(http.StatusOK, tmpl)
}

func (h *BillingHandler) SaveEmailTemplate(c *gin.Context) {
	if h.svc == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}
	var tmpl models.EmailTemplate
	if err := c.ShouldBindJSON(&tmpl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tmpl.OrganizationID = c.Param("orgId")
	h.svc.SaveEmailTemplate(&tmpl)
	c.JSON(http.StatusOK, tmpl)
}

func (h *BillingHandler) ListRegionConfigs(c *gin.Context) {
	if h.svc == nil {
		c.JSON(http.StatusOK, []models.RegionConfig{})
		return
	}
	configs, _ := h.svc.ListRegionConfigs(c.Param("orgId"))
	c.JSON(http.StatusOK, configs)
}

func (h *BillingHandler) GetRegionConfig(c *gin.Context) {
	if h.svc == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "region config not found"})
		return
	}
	cfg, err := h.svc.GetRegionConfig(c.Param("orgId"), c.Param("region"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "region config not found"})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

func (h *BillingHandler) SaveRegionConfig(c *gin.Context) {
	if h.svc == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "region config not found"})
		return
	}
	var cfg models.RegionConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cfg.OrganizationID = c.Param("orgId")
	h.svc.SaveRegionConfig(&cfg)
	c.JSON(http.StatusOK, cfg)
}
