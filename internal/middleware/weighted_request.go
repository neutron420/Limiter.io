package middleware

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func WeightedRequestMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		weight := 1
		path := c.Request.URL.Path
		method := c.Request.Method

		if strings.HasPrefix(path, "/api/export") || strings.HasPrefix(path, "/api/bulk") {
			weight = 10
		} else if strings.HasPrefix(path, "/api/reports") || strings.HasPrefix(path, "/api/analytics") {
			weight = 5
		} else if method == "POST" && (strings.Contains(path, "/import") || strings.Contains(path, "/upload")) {
			weight = 20
		} else if method == "DELETE" {
			weight = 3
		}

		if w := c.GetHeader("X-Request-Weight"); w != "" {
			if parsed, err := strconv.Atoi(w); err == nil && parsed > 0 && parsed <= 100 {
				weight = parsed
			}
		}

		c.Set("request_weight", weight)
		c.Next()
	}
}
