package otel

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func TracingMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		spanName := c.Request.Method + " " + c.FullPath()
		ctx, span := StartSpan(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		AddSpanAttributes(span,
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.path", c.Request.URL.Path),
			attribute.String("http.host", c.Request.Host),
			attribute.String("http.user_agent", c.Request.UserAgent()),
		)

		c.Request = c.Request.WithContext(ctx)
		start := time.Now()

		c.Next()

		latency := time.Since(start).Milliseconds()
		statusCode := c.Writer.Status()

		AddSpanAttributes(span,
			attribute.Int("http.status_code", statusCode),
			attribute.Int64("http.latency_ms", latency),
		)

		if statusCode >= 500 {
			span.SetStatus(codes.Error, "server error")
		}

		span.SetName(c.Request.Method + " " + c.FullPath() + " " + strconv.Itoa(statusCode))
	}
}
