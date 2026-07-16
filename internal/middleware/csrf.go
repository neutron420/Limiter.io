package middleware

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"sync"
	"time"

	"limiter.io/internal/dto"

	"github.com/gin-gonic/gin"
)

type csrfEntry struct {
	tokenHash string
	expiresAt time.Time
}

var (
	csrfTokens     = make(map[string]*csrfEntry)
	csrfTokensLock sync.RWMutex
)

const csrfTokenTTL = 2 * time.Hour

func init() {
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			csrfTokensLock.Lock()
			for id, entry := range csrfTokens {
				if time.Now().After(entry.expiresAt) {
					delete(csrfTokens, id)
				}
			}
			csrfTokensLock.Unlock()
		}
	}()
}

func GenerateCSRFToken(sessionID string) string {
	buf := make([]byte, 32)
	_, _ = io.ReadFull(rand.Reader, buf)
	token := hex.EncodeToString(buf)
	hash := sha256.Sum256([]byte(token))
	csrfTokensLock.Lock()
	csrfTokens[sessionID] = &csrfEntry{
		tokenHash: hex.EncodeToString(hash[:]),
		expiresAt: time.Now().Add(csrfTokenTTL),
	}
	csrfTokensLock.Unlock()
	return token
}

func ValidateCSRFToken(sessionID, token string) bool {
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])
	csrfTokensLock.RLock()
	entry, exists := csrfTokens[sessionID]
	csrfTokensLock.RUnlock()
	if !exists {
		return false
	}
	if time.Now().After(entry.expiresAt) {
		csrfTokensLock.Lock()
		delete(csrfTokens, sessionID)
		csrfTokensLock.Unlock()
		return false
	}
	return entry.tokenHash == tokenHash
}

func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet ||
			c.Request.Method == http.MethodHead ||
			c.Request.Method == http.MethodOptions ||
			c.Request.Method == http.MethodTrace {
			c.Next()
			return
		}

		sessionID := c.GetHeader("X-Session-ID")
		csrfToken := c.GetHeader("X-CSRF-Token")

		if sessionID == "" || csrfToken == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, dto.ErrorResponse{
				Error: "CSRF protection: missing X-Session-ID or X-CSRF-Token header",
			})
			return
		}

		if !ValidateCSRFToken(sessionID, csrfToken) {
			c.AbortWithStatusJSON(http.StatusForbidden, dto.ErrorResponse{
				Error: "CSRF protection: invalid or expired token",
			})
			return
		}

		c.Next()
	}
}
