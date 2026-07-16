package services

import (
	"log"
	"time"

	"limiter.io/internal/models"
	"gorm.io/gorm"
)

type BillingService struct {
	db *gorm.DB
}

func NewBillingService(db *gorm.DB) *BillingService {
	return &BillingService{db: db}
}

func (s *BillingService) RecordUsage(projectID string, requestCount, blockedCount int64) error {
	now := time.Now().UTC()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second)

	record := models.UsageRecord{
		ProjectID:    projectID,
		RequestCount: requestCount,
		BlockedCount: blockedCount,
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
		Tier:         "pro",
	}
	return s.db.Create(&record).Error
}

func (s *BillingService) GetUsage(projectID string, year, month int) (*models.UsageRecord, error) {
	var record models.UsageRecord
	err := s.db.Where("project_id = ? AND EXTRACT(YEAR FROM period_start) = ? AND EXTRACT(MONTH FROM period_start) = ?",
		projectID, year, month).First(&record).Error
	return &record, err
}

func (s *BillingService) CheckUsageLimit(projectID string, maxRequests int64) (bool, error) {
	now := time.Now().UTC()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	var total int64
	s.db.Model(&models.UsageRecord{}).
		Where("project_id = ? AND period_start >= ?", projectID, periodStart).
		Select("COALESCE(SUM(request_count), 0)").Scan(&total)
	return total < maxRequests, nil
}

func (s *BillingService) GenerateInvoice(orgID, projectID string, amount int64, currency string) (*models.Invoice, error) {
	now := time.Now().UTC()
	invoice := &models.Invoice{
		OrganizationID: orgID,
		ProjectID:      projectID,
		Amount:         amount,
		Currency:       currency,
		Status:         "pending",
		PeriodStart:    time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:      now,
	}
	if err := s.db.Create(invoice).Error; err != nil {
		return nil, err
	}
	log.Printf("Invoice created: %s for $%.2f", invoice.ID, float64(amount)/100)
	return invoice, nil
}

func (s *BillingService) ListInvoices(projectID string) ([]models.Invoice, error) {
	var invoices []models.Invoice
	err := s.db.Where("project_id = ?", projectID).Order("created_at DESC").Find(&invoices).Error
	return invoices, err
}

func (s *BillingService) GetInvoice(id string) (*models.Invoice, error) {
	var invoice models.Invoice
	err := s.db.First(&invoice, "id = ?", id).Error
	return &invoice, err
}

func (s *BillingService) MarkInvoicePaid(id, stripeInvoiceID string) error {
	now := time.Now()
	return s.db.Model(&models.Invoice{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":           "paid",
		"paid_at":          &now,
		"stripe_invoice_id": stripeInvoiceID,
	}).Error
}

func (s *BillingService) GetSLAConfig(orgID string) (*models.SLAConfig, error) {
	var cfg models.SLAConfig
	err := s.db.Where("organization_id = ?", orgID).First(&cfg).Error
	if err != nil {
		return &models.SLAConfig{
			OrganizationID: orgID,
			UptimeSLA:      99.9,
			ResponseTimeP99: 100,
			SupportLevel:    "standard",
		}, nil
	}
	return &cfg, nil
}

func (s *BillingService) UpdateSLAConfig(cfg *models.SLAConfig) error {
	var existing models.SLAConfig
	result := s.db.Where("organization_id = ?", cfg.OrganizationID).First(&existing)
	if result.Error != nil {
		return s.db.Create(cfg).Error
	}
	cfg.ID = existing.ID
	cfg.CreatedAt = existing.CreatedAt
	cfg.UpdatedAt = time.Now()
	return s.db.Save(cfg).Error
}

func (s *BillingService) GetEmailTemplate(orgID, name string) (*models.EmailTemplate, error) {
	var tmpl models.EmailTemplate
	err := s.db.Where("organization_id = ? AND name = ? AND is_active = ?", orgID, name, true).First(&tmpl).Error
	return &tmpl, err
}

func (s *BillingService) SaveEmailTemplate(tmpl *models.EmailTemplate) error {
	var existing models.EmailTemplate
	result := s.db.Where("organization_id = ? AND name = ?", tmpl.OrganizationID, tmpl.Name).First(&existing)
	if result.Error != nil {
		return s.db.Create(tmpl).Error
	}
	tmpl.ID = existing.ID
	tmpl.CreatedAt = existing.CreatedAt
	tmpl.UpdatedAt = time.Now()
	return s.db.Save(tmpl).Error
}

func (s *BillingService) GetRegionConfig(orgID, region string) (*models.RegionConfig, error) {
	var cfg models.RegionConfig
	err := s.db.Where("organization_id = ? AND region = ?", orgID, region).First(&cfg).Error
	return &cfg, err
}

func (s *BillingService) ListRegionConfigs(orgID string) ([]models.RegionConfig, error) {
	var configs []models.RegionConfig
	err := s.db.Where("organization_id = ?", orgID).Find(&configs).Error
	return configs, err
}

func (s *BillingService) SaveRegionConfig(cfg *models.RegionConfig) error {
	var existing models.RegionConfig
	result := s.db.Where("organization_id = ? AND region = ?", cfg.OrganizationID, cfg.Region).First(&existing)
	if result.Error != nil {
		return s.db.Create(cfg).Error
	}
	cfg.ID = existing.ID
	cfg.CreatedAt = existing.CreatedAt
	cfg.UpdatedAt = time.Now()
	return s.db.Save(cfg).Error
}
