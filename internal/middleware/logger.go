package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Logger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		reqID, _ := c.Get("RequestID")
		userID, _ := c.Get("UserID")
		projectID, _ := c.Get("ProjectID")
		apiKeyID, _ := c.Get("APIKeyID")

		fields := []zap.Field{
			zap.Int("status", status),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.Duration("latency", latency),
			zap.String("request_id", reqID.(string)),
		}

		if userID != nil {
			fields = append(fields, zap.String("user_id", userID.(string)))
		}
		if projectID != nil {
			fields = append(fields, zap.String("project_id", projectID.(string)))
		}
		if apiKeyID != nil {
			fields = append(fields, zap.String("api_key_id", apiKeyID.(string)))
		}

		if len(c.Errors) > 0 {
			for _, e := range c.Errors.Errors() {
				log.Error(e, fields...)
			}
		} else {
			log.Info("HTTP Request", fields...)
		}
	}
}
