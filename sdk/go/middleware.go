package sdk

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GinRateLimit returns a Gin middleware that rate limits endpoints using the SDK Client.
func GinRateLimit(client *Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use the request path as the verification route
		allowed, err := client.Verify(c.Request.Context(), c.Request.URL.Path)
		if err != nil {
			// Fail-open: If the rate limiting service is down, allow request but log warning
			c.Next()
			return
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too Many Requests. Rate limit exceeded.",
			})
			return
		}

		c.Next()
	}
}
