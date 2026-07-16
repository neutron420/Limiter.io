package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"limiter.io/internal/config"
	"limiter.io/internal/dto"

	"github.com/gin-gonic/gin"
)

type ipRateLimitEntry struct {
	count    int
	windowAt time.Time
}

var (
	authRateLimiters     = make(map[string]*ipRateLimitEntry)
	authRateLimitersLock sync.RWMutex
)

const (
	authRateLimitPerMinute = 10
	authCleanupInterval    = 5 * time.Minute
)

func init() {
	go func() {
		ticker := time.NewTicker(authCleanupInterval)
		defer ticker.Stop()
		for range ticker.C {
			authRateLimitersLock.Lock()
			for ip, entry := range authRateLimiters {
				if time.Since(entry.windowAt) > time.Minute {
					delete(authRateLimiters, ip)
				}
			}
			authRateLimitersLock.Unlock()
		}
	}()
}

func RateLimitAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.Env == "test" {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		key := fmt.Sprintf("auth:%s:%s", clientIP, c.FullPath())

		authRateLimitersLock.Lock()
		entry, exists := authRateLimiters[key]
		now := time.Now()

		if !exists || now.Sub(entry.windowAt) > time.Minute {
			authRateLimiters[key] = &ipRateLimitEntry{
				count:    1,
				windowAt: now,
			}
			authRateLimitersLock.Unlock()
			c.Next()
			return
		}

		entry.count++
		if entry.count > authRateLimitPerMinute {
			authRateLimitersLock.Unlock()
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, dto.ErrorResponse{
				Error: "Too many authentication requests. Please try again later.",
			})
			return
		}
		authRateLimitersLock.Unlock()
		c.Next()
	}
}
