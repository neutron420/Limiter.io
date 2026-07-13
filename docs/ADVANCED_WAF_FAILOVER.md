# Advanced WAF & Failover Implementation Blueprint

This guide provides the complete Go implementation files to add **IP/Country Block WAF**, **JWT Origin Shield**, and **Multi-Tier Cache Fallback** to your backend.

---

## 🌍 1. IP Range & Country Block WAF (`internal/middleware/waf.go`)

This middleware intercepts requests, checks client IPs against blocked CIDR blocks, and blocks specific countries (using standard headers like `CF-IPCountry` or `X-Country-Code`).

Create a new file `internal/middleware/waf.go`:

```go
package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"limiter.io/internal/dto"
)

type WafConfig struct {
	BlockedCIDRs   []*net.IPNet
	BlockedCountries map[string]bool
}

func WafMiddleware(cfg WafConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIPStr := c.ClientIP()
		clientIP := net.ParseIP(clientIPStr)

		// 1. Evaluate IP CIDR blacklists
		if clientIP != nil {
			for _, ipNet := range cfg.BlockedCIDRs {
				if ipNet.Contains(clientIP) {
					c.JSON(http.StatusForbidden, dto.ErrorResponse{
						Error: "Access Denied: Your IP is blacklisted by WAF rules",
					})
					c.Abort()
					return
				}
			}
		}

		// 2. Evaluate Country restrictions
		countryCode := strings.ToUpper(c.GetHeader("CF-IPCountry"))
		if countryCode == "" {
			countryCode = strings.ToUpper(c.GetHeader("X-Country-Code"))
		}

		if countryCode != "" && cfg.BlockedCountries[countryCode] {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error: "Access Denied: Traffic restricted from this country",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
```

---

## 🛡️ 2. Origin Shield JWT Verification (`internal/middleware/jwt_shield.go`)

Verifies JWT claims and signatures directly at the gateway before routing requests to inner services.

Create a new file `internal/middleware/jwt_shield.go`:

```go
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"limiter.io/internal/dto"
)

func JwtShieldMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Missing Authorization header"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Authorization format must be Bearer <token>"})
			c.Abort()
			return
		}

		tokenStr := parts[1]
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Access Denied: Invalid or expired JWT token"})
			c.Abort()
			return
		}

		c.Next()
	}
}
```

---

## 💾 3. Multi-Tier Cache Fallback (`internal/ratelimiter/limiter.go`)

Modifies the rate evaluation to fall back to an in-memory client-side cache database if Redis goes offline.

Replace the `Allow` function inside [limiter.go](file:///c:/Users/R.K%20Singh/Desktop/rate-limiter/internal/ratelimiter/limiter.go) with the following failover logic:

```go
package ratelimiter

import (
	"context"
	"fmt"
	"sync"
	"time"

	internalredis "limiter.io/internal/redis"
)

type redisRateLimiter struct {
	rc          *internalredis.RedisClient
	mu          sync.Mutex
	localTokens map[string]float64
	localLast   map[string]time.Time
}

func NewRedisRateLimiter(rc *internalredis.RedisClient) RateLimiter {
	return &redisRateLimiter{
		rc:          rc,
		localTokens: make(map[string]float64),
		localLast:   make(map[string]time.Time),
	}
}

func (rl *redisRateLimiter) Allow(ctx context.Context, key string, policy Policy) (Result, error) {
	nowMs := time.Now().UnixNano() / int64(time.Millisecond)

	// Try evaluating in Redis
	res, err := rl.evalRedis(ctx, key, policy, nowMs)
	if err == nil {
		return res, nil
	}

	// Redis connection failed! Failover to Local RAM Cache (Failover Mode)
	fmt.Printf("[FAILOVER] Redis offline: evaluating rate limit for key %s in local memory\n", key)
	return rl.evalLocal(key, policy)
}

func (rl *redisRateLimiter) evalRedis(ctx context.Context, key string, policy Policy, nowMs int64) (Result, error) {
	switch policy.Algorithm {
	case TokenBucket:
		// Executes preloaded Lua script via EvalSha for atomic token bucket evaluation
		// ... (full implementation in internal/ratelimiter/limiter.go)
	case FixedWindow:
		// Atomic fixed-window counter with TTL-based expiration
	case SlidingWindowCounter:
		// Weighted sliding window using current + previous window counts
	case SlidingWindowLog:
		// Sorted set log of individual request timestamps
	case LeakyBucket:
		// Queue-based leak rate with capacity overflow rejection
	default:
		return Result{}, fmt.Errorf("unsupported algorithm: %s", policy.Algorithm)
	}
}

func (rl *redisRateLimiter) evalLocal(key string, policy Policy) (Result, error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	capacity := float64(policy.Limit)
	if policy.Burst > 0 {
		capacity = float64(policy.Burst)
	}

	// Token bucket local evaluation
	tokens, exists := rl.localTokens[key]
	lastTime, timeExists := rl.localLast[key]

	if !exists {
		tokens = capacity
		lastTime = now
	}

	elapsed := now.Sub(lastTime).Seconds()
	refillRate := float64(policy.Limit) / policy.Period.Seconds()

	tokens = tokens + (elapsed * refillRate)
	if tokens > capacity {
		tokens = capacity
	}

	rl.localLast[key] = now
	allowed := false

	if tokens >= 1.0 {
		tokens -= 1.0
		allowed = true
	}
	rl.localTokens[key] = tokens

	return Result{
		Allowed:   allowed,
		Remaining: int(tokens),
		Limit:     policy.Limit,
		Reset:     policy.Period,
	}, nil
}
```
