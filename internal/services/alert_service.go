package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"limiter.io/internal/dto"
	"limiter.io/internal/mailer"
	"limiter.io/internal/models"
	"limiter.io/internal/repository"

	"github.com/google/uuid"
)

var validAlertMetrics = map[string]bool{
	"block_rate":    true, // % of requests blocked in window
	"traffic_spike": true, // total requests in window
	"avg_latency":   true, // average latency ms in window
}

type AlertService interface {
	CreateRule(ctx context.Context, userID, projectID uuid.UUID, req dto.CreateAlertRequest) (*models.AlertRule, error)
	ListRules(ctx context.Context, userID, projectID uuid.UUID) ([]models.AlertRule, error)
	UpdateRule(ctx context.Context, userID, projectID, ruleID uuid.UUID, req dto.UpdateAlertRequest) (*models.AlertRule, error)
	DeleteRule(ctx context.Context, userID, projectID, ruleID uuid.UUID) error
	ListEvents(ctx context.Context, userID, projectID uuid.UUID, limit int) ([]models.AlertEvent, error)
	// StartEvaluator runs the periodic alert evaluation loop until ctx is done.
	StartEvaluator(ctx context.Context, interval time.Duration)
}

type alertService struct {
	alertRepo   repository.AlertRepository
	analRepo    repository.AnalyticsRepository
	projectRepo repository.ProjectRepository
	memberRepo  repository.ProjectMemberRepository
	mail        mailer.Mailer
	httpClient  *http.Client
}

func NewAlertService(
	alertRepo repository.AlertRepository,
	analRepo repository.AnalyticsRepository,
	projectRepo repository.ProjectRepository,
	memberRepo repository.ProjectMemberRepository,
	mail mailer.Mailer,
) AlertService {
	return &alertService{
		alertRepo:   alertRepo,
		analRepo:    analRepo,
		projectRepo: projectRepo,
		memberRepo:  memberRepo,
		mail:        mail,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *alertService) canManage(ctx context.Context, userID, projectID uuid.UUID) error {
	proj, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return errors.New("project not found")
	}
	if proj.UserID == userID {
		return nil
	}
	// admins may manage alerts; read-only members may not
	members, err := s.memberRepo.ListByProject(ctx, projectID)
	if err == nil {
		for _, m := range members {
			if m.UserID == userID && m.Role == "admin" {
				return nil
			}
		}
	}
	return errors.New("insufficient role to manage alerts for this project")
}

func (s *alertService) canView(ctx context.Context, userID, projectID uuid.UUID) error {
	proj, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return errors.New("project not found")
	}
	if proj.UserID == userID {
		return nil
	}
	isMem, _ := s.memberRepo.IsMember(ctx, projectID, userID)
	if !isMem {
		return errors.New("unauthorized to view alerts for this project")
	}
	return nil
}

func (s *alertService) CreateRule(ctx context.Context, userID, projectID uuid.UUID, req dto.CreateAlertRequest) (*models.AlertRule, error) {
	if err := s.canManage(ctx, userID, projectID); err != nil {
		return nil, err
	}
	if !validAlertMetrics[req.Metric] {
		return nil, errors.New("invalid metric: use block_rate, traffic_spike or avg_latency")
	}
	if req.Channel != "email" && req.Channel != "webhook" && req.Channel != "slack" {
		return nil, errors.New("invalid channel: use email, webhook, or slack")
	}
	if req.Target == "" {
		return nil, errors.New("target (email address or webhook URL) is required")
	}
	window := req.WindowMinutes
	if window <= 0 || window > 1440 {
		window = 5
	}

	rule := &models.AlertRule{
		ID:            uuid.New(),
		ProjectID:     projectID,
		Name:          req.Name,
		Metric:        req.Metric,
		Threshold:     req.Threshold,
		WindowMinutes: window,
		Channel:       req.Channel,
		Target:        req.Target,
		IsActive:      true,
	}
	if err := s.alertRepo.CreateRule(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *alertService) ListRules(ctx context.Context, userID, projectID uuid.UUID) ([]models.AlertRule, error) {
	if err := s.canView(ctx, userID, projectID); err != nil {
		return nil, err
	}
	return s.alertRepo.ListRulesByProject(ctx, projectID)
}

func (s *alertService) UpdateRule(ctx context.Context, userID, projectID, ruleID uuid.UUID, req dto.UpdateAlertRequest) (*models.AlertRule, error) {
	if err := s.canManage(ctx, userID, projectID); err != nil {
		return nil, err
	}
	rule, err := s.alertRepo.GetRule(ctx, ruleID)
	if err != nil || rule.ProjectID != projectID {
		return nil, errors.New("alert rule not found")
	}
	if req.Name != nil {
		rule.Name = *req.Name
	}
	if req.Threshold != nil {
		rule.Threshold = *req.Threshold
	}
	if req.WindowMinutes != nil && *req.WindowMinutes > 0 && *req.WindowMinutes <= 1440 {
		rule.WindowMinutes = *req.WindowMinutes
	}
	if req.Channel != nil {
		if *req.Channel != "email" && *req.Channel != "webhook" && *req.Channel != "slack" {
			return nil, errors.New("invalid channel")
		}
		rule.Channel = *req.Channel
	}
	if req.Target != nil {
		rule.Target = *req.Target
	}
	if req.IsActive != nil {
		rule.IsActive = *req.IsActive
	}
	if err := s.alertRepo.UpdateRule(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *alertService) DeleteRule(ctx context.Context, userID, projectID, ruleID uuid.UUID) error {
	if err := s.canManage(ctx, userID, projectID); err != nil {
		return err
	}
	rule, err := s.alertRepo.GetRule(ctx, ruleID)
	if err != nil || rule.ProjectID != projectID {
		return errors.New("alert rule not found")
	}
	return s.alertRepo.DeleteRule(ctx, ruleID)
}

func (s *alertService) ListEvents(ctx context.Context, userID, projectID uuid.UUID, limit int) ([]models.AlertEvent, error) {
	if err := s.canView(ctx, userID, projectID); err != nil {
		return nil, err
	}
	return s.alertRepo.ListEventsByProject(ctx, projectID, limit)
}

// ---------------------------------------------------------------------------
// Evaluator
// ---------------------------------------------------------------------------

func (s *alertService) StartEvaluator(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = time.Minute
	}
	ticker := time.NewTicker(interval)
	log.Printf("[alerts] evaluator started (interval %s)", interval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Println("[alerts] evaluator stopped")
				return
			case <-ticker.C:
				s.evaluateAll(ctx)
			}
		}
	}()
}

func (s *alertService) evaluateAll(ctx context.Context) {
	rules, err := s.alertRepo.ListActiveRules(ctx)
	if err != nil {
		log.Printf("[alerts] failed listing rules: %v", err)
		return
	}
	for _, rule := range rules {
		// Cooldown: don't re-fire within one window of the last firing.
		if rule.LastFiredAt != nil &&
			time.Since(*rule.LastFiredAt) < time.Duration(rule.WindowMinutes)*time.Minute {
			continue
		}
		s.evaluateRule(ctx, rule)
	}
}

func (s *alertService) evaluateRule(ctx context.Context, rule models.AlertRule) {
	end := time.Now()
	start := end.Add(-time.Duration(rule.WindowMinutes) * time.Minute)

	stats, err := s.analRepo.GetAggregatedStats(ctx, rule.ProjectID, start, end)
	if err != nil {
		return
	}

	total := toFloat(stats["total_requests"])
	blocked := toFloat(stats["blocked_requests"])
	avgLatency := toFloat(stats["avg_latency_ms"])

	var value float64
	switch rule.Metric {
	case "block_rate":
		if total == 0 {
			return
		}
		value = blocked / total * 100
	case "traffic_spike":
		value = total
	case "avg_latency":
		if total == 0 {
			return
		}
		value = avgLatency
	default:
		return
	}

	if value <= rule.Threshold {
		return
	}

	msg := fmt.Sprintf("Alert %q fired: %s = %.2f exceeded threshold %.2f over the last %d minute(s).",
		rule.Name, rule.Metric, value, rule.Threshold, rule.WindowMinutes)

	delivered := s.deliver(ctx, rule, value, msg)

	event := &models.AlertEvent{
		RuleID:    rule.ID,
		ProjectID: rule.ProjectID,
		Metric:    rule.Metric,
		Value:     value,
		Threshold: rule.Threshold,
		Message:   msg,
		Delivered: delivered,
	}
	_ = s.alertRepo.CreateEvent(ctx, event)

	now := time.Now()
	rule.LastFiredAt = &now
	_ = s.alertRepo.UpdateRule(ctx, &rule)
}

func (s *alertService) deliver(ctx context.Context, rule models.AlertRule, value float64, msg string) bool {
	switch rule.Channel {
	case "email":
		html := fmt.Sprintf(`
			<div style="font-family:monospace;max-width:480px;margin:auto">
				<h2 style="text-transform:uppercase;letter-spacing:2px">⚠ Limiter.io Alert</h2>
				<p>%s</p>
				<table style="border-collapse:collapse;font-size:13px">
					<tr><td style="padding:4px 12px 4px 0;color:#888">METRIC</td><td><b>%s</b></td></tr>
					<tr><td style="padding:4px 12px 4px 0;color:#888">VALUE</td><td><b>%.2f</b></td></tr>
					<tr><td style="padding:4px 12px 4px 0;color:#888">THRESHOLD</td><td><b>%.2f</b></td></tr>
				</table>
			</div>`, msg, rule.Metric, value, rule.Threshold)
		if err := s.mail.Send(ctx, rule.Target, "⚠ Limiter.io alert: "+rule.Name, html); err != nil {
			log.Printf("[alerts] email delivery failed for rule %s: %v", rule.ID, err)
			return false
		}
		return true

	case "slack":
		slackPayload := map[string]interface{}{
			"text": fmt.Sprintf("⚠ *Limiter.io Alert: %s*\n> %s\n> *Metric:* %s\n> *Value:* %.2f\n> *Threshold:* %.2f",
				rule.Name, msg, rule.Metric, value, rule.Threshold),
			"mrkdwn": true,
		}
		payload, _ := json.Marshal(slackPayload)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, rule.Target, bytes.NewReader(payload))
		if err != nil {
			return false
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "limiter-io-alerts/1.0")
		resp, err := s.httpClient.Do(req)
		if err != nil {
			log.Printf("[alerts] slack delivery failed for rule %s: %v", rule.ID, err)
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode < 300

	case "webhook":
		payload, _ := json.Marshal(map[string]interface{}{
			"alert":      rule.Name,
			"project_id": rule.ProjectID,
			"metric":     rule.Metric,
			"value":      value,
			"threshold":  rule.Threshold,
			"message":    msg,
			"fired_at":   time.Now().UTC(),
		})
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, rule.Target, bytes.NewReader(payload))
		if err != nil {
			return false
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "limiter-io-alerts/1.0")
		resp, err := s.httpClient.Do(req)
		if err != nil {
			log.Printf("[alerts] webhook delivery failed for rule %s: %v", rule.ID, err)
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode < 300
	}
	return false
}

func toFloat(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int64:
		return float64(n)
	case int:
		return float64(n)
	}
	return 0
}
