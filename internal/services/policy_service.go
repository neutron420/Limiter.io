package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"limiter.io/internal/dto"
	"limiter.io/internal/models"
	"limiter.io/internal/repository"

	"github.com/google/uuid"
)

type PolicyService interface {
	CreateRule(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, req dto.CreateRuleRequest) (*models.RateLimitRule, error)
	GetRule(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, ruleID uuid.UUID) (*models.RateLimitRule, error)
	ListRules(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) ([]models.RateLimitRule, error)
	UpdateRule(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, ruleID uuid.UUID, req dto.UpdateRuleRequest) (*models.RateLimitRule, error)
	DeleteRule(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, ruleID uuid.UUID) error
	SimulateRule(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, ruleID uuid.UUID, numRequests int, reqsPerSecond float64) ([]dto.SimulationStep, error)
}

type policyService struct {
	ruleRepo    repository.RateLimitRuleRepository
	projectRepo repository.ProjectRepository
	subRepo     repository.SubscriptionRepository
	memberRepo  repository.ProjectMemberRepository
}

func NewPolicyService(
	ruleRepo repository.RateLimitRuleRepository,
	projectRepo repository.ProjectRepository,
	subRepo repository.SubscriptionRepository,
	memberRepo repository.ProjectMemberRepository,
) PolicyService {
	return &policyService{
		ruleRepo:    ruleRepo,
		projectRepo: projectRepo,
		subRepo:     subRepo,
		memberRepo:  memberRepo,
	}
}

func (s *policyService) checkProjectAccess(ctx context.Context, userID, projectID uuid.UUID) error {
	role := roleForProject(ctx, s.projectRepo, s.memberRepo, userID, projectID)
	if !canRead(role) {
		return errors.New("unauthorized to access this project")
	}
	return nil
}

func (s *policyService) checkProjectWriteAccess(ctx context.Context, userID, projectID uuid.UUID) error {
	role := roleForProject(ctx, s.projectRepo, s.memberRepo, userID, projectID)
	if !canWrite(role) {
		return errors.New("insufficient role: read-only members cannot modify the project")
	}
	return nil
}

func (s *policyService) CreateRule(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, req dto.CreateRuleRequest) (*models.RateLimitRule, error) {
	proj, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}
	if err := s.checkProjectWriteAccess(ctx, userID, projectID); err != nil {
		return nil, err
	}

	// Retrieve project owner subscription details for limits
	sub, err := s.subRepo.GetByUserID(ctx, proj.UserID)
	if err != nil {
		return nil, errors.New("subscription not found for project owner")
	}

	// Verify allowed algorithms for the subscription plan
	allowedAlgos := strings.Split(sub.Plan.AllowedAlgorithms, ",")
	algoAllowed := false
	for _, algo := range allowedAlgos {
		if strings.TrimSpace(algo) == req.Algorithm {
			algoAllowed = true
			break
		}
	}

	if !algoAllowed {
		return nil, errors.New("the selected rate-limiting algorithm is not available on your plan. Please upgrade to Pro")
	}

	// Validate key strategy if provided
	keyStrategy := req.KeyStrategy
	if keyStrategy == "" {
		keyStrategy = "api_key"
	}
	if keyStrategy != "api_key" && keyStrategy != "ip" && !strings.HasPrefix(keyStrategy, "header:") {
		return nil, errors.New("invalid key strategy. Must be 'api_key', 'ip', or 'header:<name>'")
	}

	rule := &models.RateLimitRule{
		ID:           uuid.New(),
		ProjectID:    projectID,
		Name:         req.Name,
		RoutePattern: req.RoutePattern,
		Algorithm:    req.Algorithm,
		KeyStrategy:  keyStrategy,
		Limit:        req.Limit,
		Period:       req.Period,
		Burst:        req.Burst,
		IsActive:     true,
	}

	if err := s.ruleRepo.Create(ctx, rule); err != nil {
		return nil, err
	}

	return rule, nil
}

func (s *policyService) GetRule(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, ruleID uuid.UUID) (*models.RateLimitRule, error) {
	if err := s.checkProjectAccess(ctx, userID, projectID); err != nil {
		return nil, err
	}

	rule, err := s.ruleRepo.GetByID(ctx, ruleID)
	if err != nil {
		return nil, errors.New("rate limit rule not found")
	}

	if rule.ProjectID != projectID {
		return nil, errors.New("rule does not belong to this project")
	}

	return rule, nil
}

func (s *policyService) ListRules(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) ([]models.RateLimitRule, error) {
	if err := s.checkProjectAccess(ctx, userID, projectID); err != nil {
		return nil, err
	}

	return s.ruleRepo.ListByProjectID(ctx, projectID)
}

func (s *policyService) UpdateRule(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, ruleID uuid.UUID, req dto.UpdateRuleRequest) (*models.RateLimitRule, error) {
	proj, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}
	if err := s.checkProjectWriteAccess(ctx, userID, projectID); err != nil {
		return nil, err
	}

	rule, err := s.ruleRepo.GetByID(ctx, ruleID)
	if err != nil {
		return nil, errors.New("rule not found")
	}

	if rule.ProjectID != projectID {
		return nil, errors.New("rule does not belong to this project")
	}

	// Update fields if provided
	if req.Name != nil {
		rule.Name = *req.Name
	}
	if req.RoutePattern != nil {
		rule.RoutePattern = *req.RoutePattern
	}
	if req.IsActive != nil {
		rule.IsActive = *req.IsActive
	}
	if req.Limit != nil {
		rule.Limit = *req.Limit
	}
	if req.Period != nil {
		rule.Period = *req.Period
	}
	if req.Burst != nil {
		rule.Burst = *req.Burst
	}
	if req.KeyStrategy != nil {
		strategy := *req.KeyStrategy
		if strategy != "api_key" && strategy != "ip" && !strings.HasPrefix(strategy, "header:") {
			return nil, errors.New("invalid key strategy. Must be 'api_key', 'ip', or 'header:<name>'")
		}
		rule.KeyStrategy = strategy
	}

	if req.Algorithm != nil {
		// Verify allowed algorithms for the subscription plan of project owner
		sub, err := s.subRepo.GetByUserID(ctx, proj.UserID)
		if err != nil {
			return nil, errors.New("subscription not found for project owner")
		}

		allowedAlgos := strings.Split(sub.Plan.AllowedAlgorithms, ",")
		algoAllowed := false
		for _, algo := range allowedAlgos {
			if strings.TrimSpace(algo) == *req.Algorithm {
				algoAllowed = true
				break
			}
		}

		if !algoAllowed {
			return nil, errors.New("the selected rate-limiting algorithm is not available on your plan. Please upgrade to Pro")
		}
		rule.Algorithm = *req.Algorithm
	}

	if err := s.ruleRepo.Update(ctx, rule); err != nil {
		return nil, err
	}

	return rule, nil
}

func (s *policyService) DeleteRule(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, ruleID uuid.UUID) error {
	if err := s.checkProjectWriteAccess(ctx, userID, projectID); err != nil {
		return err
	}

	rule, err := s.ruleRepo.GetByID(ctx, ruleID)
	if err != nil {
		return errors.New("rule not found")
	}

	if rule.ProjectID != projectID {
		return errors.New("rule does not belong to this project")
	}

	return s.ruleRepo.Delete(ctx, ruleID)
}

func (s *policyService) SimulateRule(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, ruleID uuid.UUID, numRequests int, reqsPerSecond float64) ([]dto.SimulationStep, error) {
	if err := s.checkProjectWriteAccess(ctx, userID, projectID); err != nil {
		return nil, err
	}

	rule, err := s.ruleRepo.GetByID(ctx, ruleID)
	if err != nil {
		return nil, errors.New("rule not found")
	}

	if rule.ProjectID != projectID {
		return nil, errors.New("rule does not belong to this project")
	}

	if numRequests <= 0 || numRequests > 500 {
		numRequests = 50
	}
	if reqsPerSecond <= 0 {
		reqsPerSecond = 10
	}

	steps := make([]dto.SimulationStep, numRequests)
	startTime := time.Now()
	interval := time.Duration(float64(time.Second) / reqsPerSecond)

	// Simulation states per algorithm
	// 1. Token Bucket state
	tbTokens := float64(rule.Limit)
	if rule.Burst > 0 {
		tbTokens = float64(rule.Burst)
	}
	tbLastRefill := startTime

	// 2. Fixed Window state
	fwCount := 0
	fwWindowStart := startTime.Unix() / int64(rule.Period)

	// 3. Sliding Window Counter state
	swcPrevCount := 0
	swcCurrCount := 0
	swcWindowStart := startTime.Unix() / int64(rule.Period)

	// 4. Sliding Window Log state
	var swlLog []time.Time

	// 5. Leaky Bucket state
	lbQueueSize := 0.0
	lbLastLeak := startTime

	for i := 0; i < numRequests; i++ {
		reqTime := startTime.Add(time.Duration(i) * interval)
		allowed := false
		remaining := 0

		switch rule.Algorithm {
		case "token_bucket":
			elapsed := reqTime.Sub(tbLastRefill).Seconds()
			tbLastRefill = reqTime
			refillRate := float64(rule.Limit) / float64(rule.Period)
			capacity := float64(rule.Limit)
			if rule.Burst > 0 {
				capacity = float64(rule.Burst)
			}
			tbTokens = tbTokens + (elapsed * refillRate)
			if tbTokens > capacity {
				tbTokens = capacity
			}

			if tbTokens >= 1.0 {
				tbTokens -= 1.0
				allowed = true
			}
			remaining = int(tbTokens)

		case "fixed_window":
			window := reqTime.Unix() / int64(rule.Period)
			if window != fwWindowStart {
				fwWindowStart = window
				fwCount = 0
			}
			if fwCount < rule.Limit {
				fwCount++
				allowed = true
			}
			remaining = rule.Limit - fwCount

		case "sliding_window_counter":
			window := reqTime.Unix() / int64(rule.Period)
			if window != swcWindowStart {
				if window == swcWindowStart+1 {
					swcPrevCount = swcCurrCount
				} else {
					swcPrevCount = 0
				}
				swcCurrCount = 0
				swcWindowStart = window
			}

			fraction := float64(reqTime.Unix()%int64(rule.Period)) / float64(rule.Period)
			estimatedRate := float64(swcPrevCount)*(1.0-fraction) + float64(swcCurrCount)

			if estimatedRate < float64(rule.Limit) {
				swcCurrCount++
				allowed = true
			}
			remaining = int(float64(rule.Limit) - estimatedRate)
			if remaining < 0 {
				remaining = 0
			}

		case "sliding_window_log":
			cutoff := reqTime.Add(-time.Duration(rule.Period) * time.Second)
			var newLog []time.Time
			for _, t := range swlLog {
				if t.After(cutoff) {
					newLog = append(newLog, t)
				}
			}
			swlLog = newLog

			if len(swlLog) < rule.Limit {
				swlLog = append(swlLog, reqTime)
				allowed = true
			}
			remaining = rule.Limit - len(swlLog)

		case "leaky_bucket":
			elapsed := reqTime.Sub(lbLastLeak).Seconds()
			lbLastLeak = reqTime
			leakRate := float64(rule.Limit) / float64(rule.Period)
			lbQueueSize = lbQueueSize - (elapsed * leakRate)
			if lbQueueSize < 0 {
				lbQueueSize = 0
			}

			capacity := float64(rule.Limit)
			if rule.Burst > 0 {
				capacity = float64(rule.Burst)
			}

			if lbQueueSize+1.0 <= capacity {
				lbQueueSize += 1.0
				allowed = true
			}
			remaining = int(capacity - lbQueueSize)

		default:
			allowed = true
		}

		steps[i] = dto.SimulationStep{
			RequestNumber: i + 1,
			Timestamp:     reqTime,
			Allowed:       allowed,
			Remaining:     remaining,
			Limit:         rule.Limit,
		}
	}

	return steps, nil
}
