package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"limiter.io/internal/dto"
	"limiter.io/internal/models"
	"limiter.io/internal/repository"
	"limiter.io/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var rawToken string

		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid authorization header format"})
				return
			}
			rawToken = parts[1]
		} else if qToken := c.Query("token"); qToken != "" {
			// Fallback for browser WebSocket connections, which cannot set the
			// Authorization header. Used by the dashboard live analytics stream.
			rawToken = qToken
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "missing authorization header"})
			return
		}

		claims, err := utils.ValidateAccessToken(rawToken, secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid or expired token"})
			return
		}

		c.Set("UserID", claims.UserID.String())
		c.Set("Email", claims.Email)
		c.Next()
	}
}

func APIKeyAuth(apiKeyRepo repository.APIKeyRepository, cacheRepo repository.CacheRepository) gin.HandlerFunc {
	// A channel to buffer API Key usage updates
	type usageUpdate struct {
		KeyID    uuid.UUID
		LastUsed time.Time
	}
	updateChan := make(chan usageUpdate, 1000)

	// Background worker to write API Key last_used_at timestamps to PostgreSQL
	go func() {
		for update := range updateChan {
			// Perform non-blocking database updates in background
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			apiKey, err := apiKeyRepo.GetByID(ctx, update.KeyID)
			if err == nil {
				apiKey.LastUsedAt = &update.LastUsed
				_ = apiKeyRepo.Update(ctx, apiKey)
			}
			cancel()
		}
	}()

	return func(c *gin.Context) {
		apiKeyHeader := c.GetHeader("X-API-Key")
		if apiKeyHeader == "" {
			// Check Authorization header for ApiKey prefix as well
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "ApiKey ") {
				apiKeyHeader = strings.TrimPrefix(authHeader, "ApiKey ")
			}
		}

		if apiKeyHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "missing API Key"})
			return
		}

		// Hash the incoming key
		keyHash := utils.HashAPIKey(apiKeyHeader)

		// Try cache lookup
		var apiKey *models.APIKey
		var err error

		dbStart := time.Now()
		apiKey, err = cacheRepo.GetAPIKey(c.Request.Context(), keyHash)
		if err != nil {
			// Cache miss: search PostgreSQL database
			apiKey, err = apiKeyRepo.GetByKeyHash(c.Request.Context(), keyHash)
			dbDuration := time.Since(dbStart).Seconds()
			PostgresLatency.WithLabelValues("get_apikey").Observe(dbDuration)

			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid API Key"})
				return
			}

			// Store key in Redis cache (5 minutes TTL)
			_ = cacheRepo.SetAPIKey(c.Request.Context(), keyHash, apiKey, 5*time.Minute)
		}

		// Check if key is expired
		if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "API Key expired"})
			return
		}

		// Check if key is revoked
		if apiKey.RevokedAt != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "API Key revoked"})
			return
		}

		// Save project ID and key ID to Gin context
		c.Set("ProjectID", apiKey.ProjectID.String())
		c.Set("APIKeyID", apiKey.ID.String())

		// Asynchronously update last used time
		select {
		case updateChan <- usageUpdate{KeyID: apiKey.ID, LastUsed: time.Now()}:
		default:
			// If buffer is full, drop to prevent blocking critical request path
		}

		c.Next()
	}
}
