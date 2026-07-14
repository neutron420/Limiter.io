package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"limiter.io/internal/dto"
	"limiter.io/internal/kafka"
	"limiter.io/internal/models"
	"limiter.io/internal/ratelimiter"
	"limiter.io/internal/repository"
	internalws "limiter.io/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func RateLimit(
	limiter ratelimiter.RateLimiter,
	ruleRepo repository.RateLimitRuleRepository,
	cacheRepo repository.CacheRepository,
	producer kafka.Producer,
	analyticsRepo repository.AnalyticsRepository,
	rc *redis.Client,
	hub *internalws.Hub,
) gin.HandlerFunc {
	// A buffered channel to handle analytics asynchronously
	analyticsChan := make(chan models.AnalyticsLog, 5000)

	// Background worker: persists each event directly to Postgres (so analytics
	// work even without a running Kafka consumer) and best-effort publishes to
	// Kafka for any downstream consumers.
	go func() {
		for event := range analyticsChan {
			// Primary path: write straight to Postgres.
			dbCtx, dbCancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := analyticsRepo.Store(dbCtx, &event); err != nil {
				log.Printf("Failed to persist analytics log to Postgres: %v", err)
			}
			dbCancel()

			// Secondary path: publish to Kafka (best effort — non-fatal if broker is down).
			data, err := json.Marshal(event)
			if err != nil {
				log.Printf("Failed to marshal analytics log for Kafka: %v", err)
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			start := time.Now()
			if err := producer.PublishEvent(ctx, event.ProjectID.String(), data); err != nil {
				log.Printf("Kafka publish failed (analytics already persisted): %v", err)
			}
			KafkaPublishLatency.Observe(time.Since(start).Seconds())
			cancel()
		}
	}()

	return func(c *gin.Context) {
		start := time.Now()
		projectIDStr, exists := c.Get("ProjectID")
		if !exists {
			c.Next()
			return
		}

		apiKeyIDStr, _ := c.Get("APIKeyID")
		reqIDStr, _ := c.Get("RequestID")

		projectID := uuid.MustParse(projectIDStr.(string))
		apiKeyID := uuid.MustParse(apiKeyIDStr.(string))
		reqID := uuid.MustParse(reqIDStr.(string))

		// Fetch rules for this project (cached for 1 minute in Redis)
		rules, err := getProjectRulesCached(c.Request.Context(), rc, ruleRepo, projectID)
		if err != nil {
			log.Printf("Error fetching rate limit rules: %v. Proceeding without rate limit.", err)
			c.Next()
			return
		}

		path := strings.TrimPrefix(c.Request.URL.Path, "/api/v1/gateway")
		var matchedRule *models.RateLimitRule

		// Find the first matching rule
		for _, r := range rules {
			if r.IsActive && matchRoute(r.RoutePattern, path) {
				matchedRule = &r
				break
			}
		}

		// If no rule matches, allow request (default open policy).
		// Signal to clients that no rate-limit policy applied to this route.
		if matchedRule == nil {
			c.Header("X-RateLimit-Policy", "none")
			c.Next()
			// Record analytics after response is written
			recordAnalytics(c, projectID, apiKeyID, reqID, "allowed", "", start, analyticsChan, hub)
			return
		}

		policy := ratelimiter.Policy{
			Limit:     matchedRule.Limit,
			Period:    time.Duration(matchedRule.Period) * time.Second,
			Burst:     matchedRule.Burst,
			Algorithm: ratelimiter.ParseAlgorithm(matchedRule.Algorithm),
		}

		// Run rate limiter decision against Redis
		var clientKeyPart string
		switch {
		case matchedRule.KeyStrategy == "ip":
			clientKeyPart = c.ClientIP()
		case strings.HasPrefix(matchedRule.KeyStrategy, "header:"):
			headerName := strings.TrimPrefix(matchedRule.KeyStrategy, "header:")
			clientKeyPart = c.GetHeader(headerName)
			if clientKeyPart == "" {
				clientKeyPart = apiKeyID.String() // fallback
			}
		default:
			clientKeyPart = apiKeyID.String()
		}

		clientKey := fmt.Sprintf("%s:%s:%s", projectID.String(), clientKeyPart, matchedRule.ID.String())
		
		redisStart := time.Now()
		result, err := limiter.Allow(c.Request.Context(), clientKey, policy)
		redisDuration := time.Since(redisStart).Seconds()
		RedisLatency.WithLabelValues(string(policy.Algorithm)).Observe(redisDuration)

		if err != nil {
			log.Printf("Rate limiter error: %v. Permitting request.", err)
			c.Next()
			recordAnalytics(c, projectID, apiKeyID, reqID, "allowed", "", start, analyticsChan, hub)
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", result.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", result.Remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", int(result.Reset.Seconds())))

		if !result.Allowed {
			RateLimiterDecisions.WithLabelValues(projectID.String(), "blocked").Inc()
			blockedReason := fmt.Sprintf("Rate limit exceeded for rule: %s", matchedRule.Name)

			c.AbortWithStatusJSON(http.StatusTooManyRequests, dto.ErrorResponse{
				Error: "Too Many Requests. Rate limit exceeded.",
			})

			recordAnalytics(c, projectID, apiKeyID, reqID, "blocked", blockedReason, start, analyticsChan, hub)
			return
		}

		RateLimiterDecisions.WithLabelValues(projectID.String(), "allowed").Inc()
		c.Next()

		recordAnalytics(c, projectID, apiKeyID, reqID, "allowed", "", start, analyticsChan, hub)
	}
}

func matchRoute(pattern, path string) bool {
	if pattern == "*" || pattern == "" {
		return true
	}
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}
	return pattern == path
}

func getProjectRulesCached(
	ctx context.Context,
	rc *redis.Client,
	ruleRepo repository.RateLimitRuleRepository,
	projectID uuid.UUID,
) ([]models.RateLimitRule, error) {
	cacheKey := fmt.Sprintf("rate_limit:cache:rules:%s", projectID.String())

	// Try reading cache
	val, err := rc.Get(ctx, cacheKey).Result()
	if err == nil {
		var rules []models.RateLimitRule
		if err := json.Unmarshal([]byte(val), &rules); err == nil {
			return rules, nil
		}
	}

	// Cache miss: load from DB
	dbStart := time.Now()
	rules, err := ruleRepo.ListByProjectID(ctx, projectID)
	dbDuration := time.Since(dbStart).Seconds()
	PostgresLatency.WithLabelValues("list_rules").Observe(dbDuration)

	if err != nil {
		return nil, err
	}

	// Save to cache (1 minute TTL)
	if data, err := json.Marshal(rules); err == nil {
		_ = rc.Set(ctx, cacheKey, data, 1*time.Minute)
	}

	return rules, nil
}

func recordAnalytics(
	c *gin.Context,
	projectID uuid.UUID,
	apiKeyID uuid.UUID,
	reqID uuid.UUID,
	decision string,
	blockedReason string,
	startTime time.Time,
	analyticsChan chan<- models.AnalyticsLog,
	hub *internalws.Hub,
) {
	latency := time.Since(startTime).Milliseconds()
	statusCode := c.Writer.Status()
	// If it was aborted, status code might still be 200 if not set, but Gin middlewares set it correctly
	
	logRecord := models.AnalyticsLog{
		ID:            uuid.New(),
		ProjectID:     projectID,
		APIKeyID:      apiKeyID,
		RequestID:     reqID,
		ClientIP:      c.ClientIP(),
		Route:         c.Request.URL.Path,
		StatusCode:    statusCode,
		LatencyMs:     int(latency),
		Decision:      decision,
		BlockedReason: blockedReason,
		Timestamp:     time.Now(),
	}

	// Push to buffered channel without blocking request execution
	select {
	case analyticsChan <- logRecord:
	default:
		// If queue is full in extreme peak load, drop to preserve low API latency
		log.Printf("Analytics buffer full. Dropping log record for request: %s", reqID.String())
	}

	// Broadcast to active WebSocket connections
	hub.Broadcast(projectID.String(), logRecord)
}
